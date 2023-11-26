package database

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var err error;
var db_read *os.File;
var db_write *os.File;
var db_mapping map[string][][]int64=make(map[string][][]int64)
var free [][]int64=make([][]int64, 0);
var temp interface{};
var fi os.FileInfo;
var instructions []map[string][]byte=make([]map[string][]byte,0)
var initialised bool=false;

func json_dumps(x interface{}) []byte {
	y,_:=json.Marshal(x)
	return y
}

func save_free_and_mapping_to_disk() {
	os.WriteFile("database/mapping.vx",json_dumps(db_mapping), 0644)
	os.WriteFile("database/free.vx",json_dumps(free), 0644)
}

func get_db_file_size() int64 {
	fi,err:=os.Stat("database/db.vx")
	if err!=nil {
		fmt.Println(err,"get_db_file_size")
		os.Exit(1)
	}
	return fi.Size()
}

func remove_dual_dimension_int64(slice [][]int64, s int) [][]int64 {
    return append(slice[:s], slice[s+1:]...)
}

func get_sectors(length int64,remove bool) [][]int64 {
	sectors:=make([][]int64,0)
	completed_length:=int64(0)
	copy_free:=free
	to_append:=make([][]int64,0)
	to_remove:=make([][]int64,0)
	for i,s:=range copy_free {
		current_space_index:=s[0]
		current_space_size:=s[1]
		if current_space_size==0 {
			to_remove = append(to_remove, copy_free[i])
			continue
		}
		// just checking whether adding this sector + all sector spaces we have added till now would exceed the length required or not
		if (completed_length+current_space_size) <= length {
			sectors = append(sectors, []int64{current_space_index,current_space_size})
			completed_length+=current_space_size
			if remove {
				to_remove = append(to_remove, copy_free[i])
			}
		} else {
			if length-completed_length==0 {
				fmt.Println("0 REQUESTED!?!?",completed_length,current_space_size,length)
				os.Exit(1)
			}
			sectors = append(sectors, []int64{current_space_index,length-completed_length})
			if remove {
				to_remove = append(to_remove, copy_free[i])
			}
			// readjusting this sector if located sector has more space than required
			if current_space_size-(length-completed_length)==0 {
				fmt.Println("0 SIZE")
				os.Exit(1)
			}
			to_append = append(to_append, []int64{current_space_index+length-completed_length-1,current_space_size-(length-completed_length)})
			completed_length=length
		}
		if length==completed_length {
			break
		}
	}
	if completed_length!=length {
		sectors = append(sectors, []int64{get_db_file_size(),length-completed_length})
	}
	for _,s := range to_append {
		free = append(free, s)
	}
	if remove {
		new_free:=make([][]int64,0)
		for _,k:=range copy_free {
			to_add:=true
			for _,s:= range to_remove {
				if s[0]==k[0] && s[1]==k[1] {
					to_add=false
					break
				}
			}
			if to_add {
				new_free = append(new_free, k)
			}
		}
		free=new_free
	}
	return sectors
}

func Get(table string, key string) []byte {
	key=table+"_"+key
	prev_val, does_exist := db_mapping[key]
	if (!does_exist) {
		return make([]byte,0)
	} else {
		out:=make([]byte,0)
		for _,s:=range prev_val {
			read_buffer:=make([]byte, s[1])
			db_read.ReadAt(read_buffer,s[0])
			out = append(out, read_buffer...)
		}
		return out
	}
}

func Set(table string, key string, val []byte) {
	in:=make(map[string][]byte)
	in["table"]=[]byte(table)
	in["key"]=[]byte(key)
	in["val"]=val
	instructions = append(instructions, in)
}

func internal_set(table string, key string, val []byte) {
	key=table+"_"+key
	prev_state, does_exist := db_mapping[key]
	if len(val)==0 {
		for _,s:=range prev_state {
			if s[1]!=0 {
				free = append(free, s)
			}
		}
		delete(db_mapping,key)
		return
	}
	if does_exist {
		for _,s:=range prev_state {
			if s[1]!=0 {
				free = append(free, s)
			}
		}
		db_mapping[key]=get_sectors(int64(len(val)),true)
		cached_val:=val
		// save data in sectors
		for _,s:=range db_mapping[key] {
			current_part:=cached_val[:s[1]]
			cached_val=cached_val[s[1]:]
			db_write.WriteAt([]byte(current_part),s[0])
		}

	} else {
		// removes required sectors out of the free list globally
		sectors_needed:=get_sectors(int64(len(val)),true)
		db_mapping[key]=sectors_needed
		cached_val:=val
		// save data in sectors
		for _,s:=range sectors_needed {
			current_part:=cached_val[:s[1]]
			cached_val=cached_val[s[1]:]
			db_write.WriteAt([]byte(current_part),s[0])
		}
	}
	save_free_and_mapping_to_disk()
}

func write_listener() {
	for {
		if len(instructions)==0 {
			time.Sleep(time.Millisecond*10)
			continue
		}
		current_instruction:=instructions[0]
		instructions=instructions[1:]
		internal_set(string(current_instruction["table"]),string(current_instruction["key"]),current_instruction["val"])
	}
}

func Init() int {
	os.OpenFile("database/db.vx",os.O_CREATE, 0644)
	os.OpenFile("database/mapping.vx",os.O_CREATE, 0644)
	os.OpenFile("database/free.vx",os.O_CREATE, 0644)
	db_read, err = os.OpenFile("database/db.vx",os.O_APPEND, 0644)
	fi, err = db_read.Stat()
	db_write, err = os.OpenFile("database/db.vx",os.O_WRONLY, 0644)

	if (err!=nil) {
		fmt.Println(err)
		return 0
	}

	// loading mapping only

	temp_raw_data_mapping, err := os.ReadFile("database/mapping.vx")
	var temp_interface map[string][][]int64
	err=json.Unmarshal(temp_raw_data_mapping,&temp_interface)

	if (err!=nil) {
		overall_dict:=make(map[string][][]int64)
		temp_interface=overall_dict
	}
	db_mapping=temp_interface

	// loading free only

	temp_free, err := os.ReadFile("database/free.vx")
	temp_interface_1:=make([][]int64,0)
	err=json.Unmarshal(temp_free,&temp_interface_1)
	free=temp_interface_1
	
	// when file is empty (mapping)
	if err != nil {
		save_free_and_mapping_to_disk()
	}

	go write_listener()
	initialised=true

	return 1
}