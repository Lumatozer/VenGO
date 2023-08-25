package main

import (
	"strings"
	"github.com/lumatozer/VenGO/database"
	"math"
)

// func do_nothing(x interface{}) {}

func contains_float64(a float64, list []float64) map[string]int {
	res:=make(map[string]int)
    for i, b := range list {
        if b == a {
			res["index"]=i
			res["contains"]=1
            return res
        }
    }
	res["contains"]=0
	res["index"]=-1
    return res
}

func contains_str(a string, list []string) map[string]int64 {
	res:=make(map[string]int64)
    for i, b := range list {
        if b == a {
			res["index"]=int64(i)
			res["contains"]=1
            return res
        }
    }
	res["contains"]=0
	res["index"]=-1
    return res
}

func plain_in_arr_VI_Object(obj VI_Object, arr []VI_Object, field string ) map[string]int {
	res:=make(map[string]int)
	for i := 0; i < len(arr); i++ {
		current_obj:=arr[i]
		res["error"]=0
		if (strings.Join(current_obj.object_type[:],",")!=strings.Join(obj.object_type[:],",")) {
			res["error"]=1
		}
		switch field {
		case "var_name":
			if (current_obj.var_name==obj.var_name) {
				res["index"]=i
				res["result"]=1
				return res
			}
		case "num_value":
			if (current_obj.num_value==obj.num_value) {
				res["index"]=i
				res["result"]=1
				return res
			}
		case "str_value":
			if (current_obj.str_value==obj.str_value) {
				res["index"]=i
				res["result"]=1
				return res
			}
		}
	}
	res["index"]=-1
	res["result"]=0
	res["error"]=0
	return res
}

func string_arr_compare(arr1 []string, arr2[]string) bool {
	if len(arr1)!=len(arr2) {
		return false
	}
	for i := 0; i < len(arr1); i++ { 
		if arr1[i]!=arr2[i] {
			return false
		}
	}
	return true
}

func recursive_VI_Object_match(obj1 VI_Object, obj2 VI_Object) bool {
	if !(string_arr_compare(obj1.object_type,obj2.object_type)) {
		return false
	}
	switch obj_type:=obj1.object_type[0]; obj_type {
	case "num":
		if (obj1.num_value!=obj2.num_value) {
			return false
		}
	case "string":
		if (obj1.str_value!=obj2.str_value) {
			return false
		}
	case "arr","dict":
		if len(obj1.children)!=len(obj2.children) {
			return false
		}
		for i := 0; i < len(obj1.dict_keys); i++ {
			isok:=false
			for j := 0; j < len(obj1.dict_keys); j++ {
				if (obj1.dict_keys[i]==obj2.dict_keys[j]) {
					isok=true
				}
			}
			if !(isok) {
				return false
			}
		}
		for i := 0; i < len(obj1.children); i++ {
			isok:=false
			for j := 0; j < len(obj1.children); j++ {
				if (recursive_VI_Object_match(obj1.children[i],obj2.children[j])) {
					isok=true
				}
			}
			if !(isok) {
				return false
			}
		}
		return true
	}
	return true
}

func str_index_in_arr(str_obj string, string_arr []string) int {
	for i := 0; i < len(string_arr); i++ {
		if string_arr[i]==str_obj {
			return i
		}
	}
	return -1
}

func get_plain_type(obj_type string) []string {
	res:=make([]string,0)
	res = append(res, obj_type)
	return res
}

func get_init_arr_type(obj_type string) []string {
	res:=make([]string,0)
	res = append(res, "arr")
	res = append(res, strings.Split(strings.ReplaceAll(obj_type," ",""),",")...)
	return res
}

func get_init_dict_type(obj_type string) []string {
	res:=make([]string,0)
	res = append(res, "dict")
	res = append(res, strings.Split(strings.ReplaceAll(obj_type," ",""),",")...)
	return res
}

func remove_VI_Object_from_index(slice []VI_Object, i int) []VI_Object {
	copy(slice[i:], slice[i+1:])
	return slice[:len(slice)-1]
}

func remove_string_from_index(slice []string, i int) []string {
	copy(slice[i:], slice[i+1:])
	return slice[:len(slice)-1]
}

type VI_Object struct {
	var_name string
	num_value float64
	str_value string
	object_type []string
	children []VI_Object
	dict_keys []string
}

func (object *VI_Object) fill_defaults() {}

func round_float64(num float64, decimals uint) float64 {
    ratio := math.Pow(10, float64(decimals))
    return math.Round(num * ratio) / ratio
}

func file_parser(dat string) map[string][]string {
	data_lines:=strings.Split(dat, "\n")
	mode:=""
	code_lines:=make([]string,0)
	data_constants:=make([]string,0)
	for i := 0; i < len(data_lines); i++ {
		current_line:=strings.Trim(strings.Trim(strings.Trim(data_lines[i],"\r"),"	"),"	")
		if (current_line=="") {
			continue
		}
		if (current_line==".code") {
			mode="code"
			continue
		} else if (current_line==".data") {
			mode="data"
			continue
		}
		if (mode=="data") {
			data_constants = append(data_constants, strings.Trim(strings.Trim(current_line," "),"\""))
		} else if (mode=="code") {
			code_lines = append(code_lines, strings.Trim(current_line," "))
		}
	}
	res:=make(map[string][]string,0)
	res["code_lines"]=code_lines
	res["data_constants"]=data_constants
	return res
}

func num_to_bool(num float64) bool {
	if num==0 {
		return false
	} else {
		return true
	}
}

func bool_to_num(x bool) float64 {
	if x {
		return 1
	} else {
		return 0
	}
}

func obj_supports_type(arr_type []string,obj_type []string) bool {
	if strings.Join(arr_type[1:],"")==strings.Join(obj_type,"") {
		return true
	} else {
		return false
	}
}

func type_evaluator(obj_type []string) bool {
	if obj_type[len(obj_type)-1]!="string" && obj_type[len(obj_type)-1]!="num" {
		return false
	}
	check_arr:=remove_string_from_index(obj_type,len(obj_type)-1)
	for i := 0; i < len(check_arr); i++ {
		if check_arr[i]=="string" || check_arr[i]=="num" {
			return false
		}
	}
	return true
}

func main() {
	if (database.Init())==0 {
		return
	}
	// database.DB_set("alu","aa")
	// database.DB_set("alux","bb")
	// database.DB_set("alu","a")
	// database.DB_set("alu","aa")
	// database.DB_delete("alux")
	// fmt.Println(database.DB_get("alu"))
	// Vengine()
}