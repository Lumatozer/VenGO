package database

import (
	"io"
	"encoding/json"
	"fmt"
	"os"
)

var err error;
var db_read *os.File;
var db_write *os.File;
var db_mapping_raw_read *os.File;
var db_mapping_raw_write *os.File;
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

func save_mapping_and_free() {
	var overall_dict map[string]interface{}=make(map[string]interface{})
	overall_dict["locs"]=db_mapping
	overall_dict["free"]=free
	dat,err:=json.Marshal(overall_dict)
	db_mapping_raw_write_local, err := os.OpenFile("database/mapping.vx",os.O_WRONLY, 0644)
	if (err!=nil) {
		fmt.Println(err)
		return
	}
	db_mapping_raw_write_local.Write(dat)
}

func remove_free_space_index(remove_index int) {
	new_free:=make([][]int64, 0)
	for i := 0; i < len(free); i++ {
		if i!=remove_index {
			new_free = append(new_free, free[i])
		}
	}
}

func DB_get(key string) string {
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
			free[space_index]=[]int64{space_info[0]+space_info[1]-1-(space_info[1]-int64(len(val))-1),space_info[1]-int64(len(val))}
		} else {
			write_at:=fi.Size()
			db_write.WriteAt([]byte(val),write_at)
			db_mapping[key]=[]int64{write_at,int64(len([]byte(val)))}
		}
	} else {
		if int64(len(val))==db_mapping[key][1] {
			db_write.WriteAt([]byte(val),db_mapping[key][0])
		}
		if int64(len(val))<db_mapping[key][1] {
			free_index:=db_mapping[key][0]+db_mapping[key][1]-1-(db_mapping[key][1]-int64(len(val))-1)
			free_space:=db_mapping[key][1]-int64(len(val))
			free = append(free, []int64{free_index,free_space})
			db_mapping[key][1]=int64(len(val))
		}
		if int64(len(val))>db_mapping[key][1] {
			free = append(free, []int64{db_mapping[key][0],db_mapping[key][1]})
			write_at:=fi.Size()
			db_write.WriteAt([]byte(val),write_at)
			db_mapping[key]=[]int64{write_at,int64(len([]byte(val)))}
		}
	}
	save_mapping_and_free()
}

func Init() int {
	os.OpenFile("database/db.vx",os.O_CREATE, 0644)
	os.OpenFile("database/mapping.vx",os.O_CREATE, 0644)
	db_read, err = os.OpenFile("database/db.vx",os.O_APPEND, 0644)
	fi, err = db_read.Stat()
	db_write, err = os.OpenFile("database/db.vx",os.O_WRONLY, 0644)
	db_mapping_raw_read, err = os.OpenFile("database/mapping.vx",os.O_APPEND, 0644)
	db_mapping_raw_write, err = os.OpenFile("database/mapping.vx",os.O_WRONLY, 0644)

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