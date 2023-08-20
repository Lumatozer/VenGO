package database

import (
	"io"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

var err error;
var db_read *os.File;
var db_write *os.File;
var db_mapping map[string][]int64=make(map[string][]int64)
var free [][]int64=make([][]int64, 0);
var temp interface{};
var fi os.FileInfo;

func get_all_db_keys() []string {
	all_keys:=make([]string,0)
	for key, value := range db_mapping {
		temp=value
		all_keys = append(all_keys, key)
	}
	return all_keys
}

func arr_contains_str(a string, arr []string) bool {
	for i := 0; i < len(arr); i++ {
		if arr[i]==a {
			return true
		}
	}
	return false
}

func space_available(length int64) int64 {
	for i := 0; i < len(free); i++ {
		if free[i][1]>=length {
			return int64(i)
		}
	}
	return -1
}

func space_available_index(original int64,index int, length int64) []int64 {
	res:=make([]int64,2)
	for i := 0; i < len(free); i++ {
		if free[i][0]==int64(index) {
			res[0]=1
			res[1]=int64(i)
			return res
		}
	}
	res[0]=0
	res[1]=space_available(length)
	return res
}

func find_int64_index(num int64, arr []int64) int {
	for i := 0; i < len(arr); i++ {
		if arr[i]==num {
			return i
		}
	}
	return -1
}

func merge_free_locs()  {
	new_free:=make([][]int64, len(free))
	free_indexes:=make([]int64,0)
	for i := 0; i < len(free); i++ {
		free_indexes = append(free_indexes, free[i][0])
	}
	free_indexes_copy:=free_indexes
	sort.Slice(free_indexes_copy,func(i, j int) bool {
		return  free_indexes_copy[i] < free_indexes_copy[j]
	})
	for i := 0; i < len(free_indexes_copy); i++ {
		new_free[i]=free[find_int64_index(free_indexes_copy[i],free_indexes)]
	}
	free=new_free
	new_free=make([][]int64, 0)
	connected_free:=make([]int64, 0)
	for i := 0; i < len(free); i++ {
		if (len(free)-i!=1) {
			if free[i][0]+free[i][1]==free[i+1][0] {
				if len(connected_free)==0 {
					connected_free = append(connected_free, free[i][0],free[i][1]+free[i+1][1])
				} else {
					connected_free[1]+=free[i+1][1]
				}
			} else {
				if len(connected_free)==0 {
					new_free = append(new_free, free[i])
				}
				if len(connected_free)!=0 {
					new_free = append(new_free, connected_free)
					connected_free=make([]int64, 0)
				}
				if (len(free)-i)==2 {
					new_free = append(new_free, free[i+1])
				}
			}
		} else {
			break
		}
	}
	if (len(free)==1) {
		new_free = append(new_free, free[0])
	}
	free=new_free
}

func save_mapping_and_free() {
	merge_free_locs()
	var overall_dict map[string]interface{}=make(map[string]interface{})
	overall_dict["locs"]=db_mapping
	overall_dict["free"]=free
	dat,err:=json.Marshal(overall_dict)
	if (err!=nil) {
		fmt.Println(err)
		return
	}
	os.WriteFile("database/mapping.vx", dat, 0644)
}

func remove_free_space_index(remove_index int) {
	new_free:=make([][]int64, 0)
	for i := 0; i < len(free); i++ {
		if i!=remove_index {
			new_free = append(new_free, free[i])
		}
	}
	free=new_free
}

func DB_get(key string) string {
	if db_mapping[key]==nil {
		return ""
	}
	db_read.Seek(db_mapping[key][0],0)
	constx:=int(db_mapping[key][1])
	header:=make([]byte,constx)
	n, err := io.ReadFull(db_read, header[:])
	if err != nil || n==-1 {
		return ""
	}
	db_read.Seek(0,0)
	return string(header)
}

func DB_set(key string,val string) {
	if len(val)==0 {
		return
	}
	fi, err := os.Stat("database/db.vx")
	if err != nil {
		fmt.Println(err)
		return
	}
	if (!arr_contains_str(key,get_all_db_keys())) {
		space_info:=space_available(int64(len(val)))
		if (space_info!=-1) {
			space_index:=space_info
			space_info:=free[space_info]
			write_at:=space_info[0]
			db_write.WriteAt([]byte(val),write_at)
			db_mapping[key]=[]int64{write_at,int64(len([]byte(val)))}
			if (space_info[1]-int64(len(val)))==0 {
				remove_free_space_index(int(space_index))
			} else {
				free[space_index]=[]int64{space_info[0]+space_info[1]-1-(space_info[1]-int64(len(val))-1),space_info[1]-int64(len(val))}
			}
		} else {
			write_at:=fi.Size()
			db_write.WriteAt([]byte(val),write_at)
			db_mapping[key]=[]int64{write_at,int64(len([]byte(val)))}
		}
	} else {
		local_db_mapping:=db_mapping[key]
		if int64(len(val))==local_db_mapping[1] {
			db_write.WriteAt([]byte(val),local_db_mapping[0])
		}
		if int64(len(val))<local_db_mapping[1] {
			free_index:=local_db_mapping[0]+local_db_mapping[1]-1-(local_db_mapping[1]-int64(len(val))-1)
			free_space:=local_db_mapping[1]-int64(len(val))
			free = append(free, []int64{free_index,free_space})
			local_db_mapping[1]=int64(len(val))
			db_write.WriteAt([]byte(val),local_db_mapping[0])
		}
		if int64(len(val))>local_db_mapping[1] {
			space_info_xy:=space_available_index(local_db_mapping[0],int(local_db_mapping[0]+local_db_mapping[1]),int64(len(val)))
			expand:=space_info_xy[0]
			space_info:=space_info_xy[1]
			if (space_info!=-1) {
				space_index:=space_info
				space_info:=free[space_info]
				if expand==1 {
					write_at:=local_db_mapping[0]
					db_write.WriteAt([]byte(val),write_at)
					db_mapping[key]=[]int64{write_at,int64(len([]byte(val)))}
					if (space_info[1]-(int64(len(val))-local_db_mapping[1]))==0 {
						remove_free_space_index(int(space_index))
					} else {
						free[space_index]=[]int64{space_info[0]+space_info[1]-1-(space_info[1]-(int64(len(val))-local_db_mapping[1])-1),space_info[1]-(int64(len(val))-local_db_mapping[1])}
					}
				} else {
					write_at:=space_info[0]
					db_write.WriteAt([]byte(val),write_at)
					db_mapping[key]=[]int64{write_at,int64(len([]byte(val)))}
					if (space_info[1]-int64(len(val)))==0 {
						remove_free_space_index(int(space_index))
					} else {
						free[space_index]=[]int64{space_info[0]+space_info[1]-1-(space_info[1]-int64(len(val))-1),space_info[1]-int64(len(val))}
					}
				}
			} else {
				if (local_db_mapping[0]+local_db_mapping[1]==fi.Size()) {
					write_at:=local_db_mapping[0]
					db_write.WriteAt([]byte(val),write_at)
					db_mapping[key]=[]int64{write_at,int64(len([]byte(val)))}
				} else {
					free = append(free, []int64{local_db_mapping[0],local_db_mapping[1]})
					write_at:=fi.Size()
					db_write.WriteAt([]byte(val),write_at)
					db_mapping[key]=[]int64{write_at,int64(len([]byte(val)))}
				}
			}
		}
	}
	save_mapping_and_free()
}

func DB_delete(key string) {
	if db_mapping[key]==nil {
		return
	}
	local_db_mapping:=db_mapping[key]
	free = append(free, []int64{local_db_mapping[0],local_db_mapping[1]})
	delete(db_mapping,key)
	save_mapping_and_free()
}

func Init() int {
	os.OpenFile("database/db.vx",os.O_CREATE, 0644)
	os.OpenFile("database/mapping.vx",os.O_CREATE, 0644)
	db_read, err = os.OpenFile("database/db.vx",os.O_APPEND, 0644)
	fi, err = db_read.Stat()
	db_write, err = os.OpenFile("database/db.vx",os.O_WRONLY, 0644)

	if (err!=nil) {
		fmt.Println(err)
		return 0
	}

	temp_raw_data_mapping, err := os.ReadFile("database/mapping.vx")
	var temp_interface interface{}
	err=json.Unmarshal(temp_raw_data_mapping,&temp_interface)

	if (err!=nil) {
		var overall_dict map[string]interface{}=make(map[string]interface{})
		overall_dict["locs"]=make(map[string]interface{})
		overall_dict["free"]=make([]interface{},0)
		temp_interface=overall_dict
	}

	for key, value := range temp_interface.(map[string]interface{})["locs"].(map[string]interface{}) {
		db_mapping[key]=[]int64{int64(value.([]interface{})[0].(float64)),int64(value.([]interface{})[1].(float64))}
	}

	temp_free:=temp_interface.(map[string]interface{})["free"].([]interface{})

	for i := 0; i < len(temp_free); i++ {
		temp_each_free:=temp_free[i].([]interface{})
		new_temp_each_free:=make([]int64,0)
		for y := 0; y < len(temp_each_free); y++ {
			new_temp_each_free = append(new_temp_each_free, int64(temp_each_free[y].(float64)))
		}
		free = append(free, new_temp_each_free)
	}

	return 1
}