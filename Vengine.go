package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

func Vengine() {
	dat, err := os.ReadFile("alu.vi")

	if (err!=nil) {
		fmt.Println(err)
		return
	}
	code:=string(dat)
	
	parse_results:=file_parser(code)
	codex,string_consts:=parse_results["code_lines"],parse_results["data_constants"]

	var num_constants []float64;
	var arr_constants []VI_Object;
	var dict_constants []VI_Object;

	byte_code:=make([][]int,0)
	symbol_table:=make([]VI_Object,0)
	jump_table:=make(map[string]int,0)
	gas_limit:=0
	current_gas:=1

	for i := 0; i < len(codex); i++ {
		args:=strings.Split(codex[i], " ")
		if len(args)>=2 {
			args=strings.Split(args[1], ",")
		} else {
			args=make([]string, 0)
		}
		current_byte_code:=make([]int,0)
		switch opcode:=strings.Split(codex[i], " ")[0]; opcode {
		case "set":
			num,err:=strconv.ParseFloat(args[1],64)
			if err!=nil {
				fmt.Println(err)
				return
			}
			index:=contains_float64(num,num_constants)
			if index["contains"]==0 {
				num_constants = append(num_constants, num)
				index["index"]=len(num_constants)-1
			}
			res:=plain_in_arr_VI_Object(VI_Object{var_name: args[0],object_type: get_plain_type("num")},symbol_table,"var_name")
			if res["error"]==1 {
				fmt.Println("Data types did not match")
				return
			}
			if res["result"]==0 {
				symbol_table = append(symbol_table, VI_Object{var_name: args[0],object_type: get_plain_type("num")})
				res["index"]=len(symbol_table)-1
			}
			current_byte_code = append(current_byte_code, 0,res["index"],index["index"])
			byte_code = append(byte_code, current_byte_code)
		case "refset","jump","not":
			set_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[0],object_type: get_plain_type("num")},symbol_table,"var_name")
			reference:=plain_in_arr_VI_Object(VI_Object{var_name: args[1],object_type: get_plain_type("num")},symbol_table,"var_name")
			if set_var["index"]==-1 {
				fmt.Println("Variable",args[0],"has not been initialised yet")
				return
			}
			if set_var["error"]==1 {
				fmt.Println("Data types did not match")
				return
			}
			if reference["index"]==-1 {
				fmt.Println("Variable",args[1],"has not been initialised yet")
				return
			}
			if reference["error"]==1 {
				fmt.Println("Data types did not match")
				return
			}
			var intopcode int;
			switch opcode {
			case "refset":intopcode=1
			case "jump":intopcode=2
			case "not":intopcode=22
			}
			current_byte_code = append(current_byte_code, intopcode, set_var["index"], reference["index"])
			byte_code = append(byte_code, current_byte_code)
		case "equals","greater","add","sub","mult","div","floor","mod","power","round","and","or","xor":
			var_1:=plain_in_arr_VI_Object(VI_Object{var_name: args[0],object_type: get_plain_type("num")},symbol_table,"var_name")
			var_2:=plain_in_arr_VI_Object(VI_Object{var_name: args[1],object_type: get_plain_type("num")},symbol_table,"var_name")
			var_res:=plain_in_arr_VI_Object(VI_Object{var_name: args[2],object_type: get_plain_type("num")},symbol_table,"var_name")
			if var_1["error"]==1 || var_2["error"]==1 || var_res["error"]==1 {
				fmt.Println("Data types did not match")
				return
			}
			if var_1["index"]==-1 {
				fmt.Println("Variable",args[0],"has not been initialised yet")
				return
			}
			if var_2["index"]==-1 {
				fmt.Println("Variable",args[1],"has not been initialised yet")
				return
			}
			if var_res["index"]==-1 {
				fmt.Println("Variable",args[2],"has not been initialised yet")
				return
			}
			var intopcode int;
			switch opcode {
			case "equals":intopcode=3
			case "greater":intopcode=4
			case "add":intopcode=5
			case "sub":intopcode=6
			case "mult":intopcode=7
			case "div":intopcode=8
			case "power":intopcode=9
			case "floor":intopcode=10
			case "mod":intopcode=11
			case "round":intopcode=24
			case "and":intopcode=20
			case "or":intopcode=21
			case "xor":intopcode=23
			}
			current_byte_code = append(current_byte_code, intopcode, var_1["index"], var_2["index"], var_res["index"])
			byte_code = append(byte_code, current_byte_code)
		case "str.set":
			num,err:=strconv.ParseInt(args[1],10,64)
			if err!=nil {
				fmt.Println(err)
				return
			}
			var_1:=plain_in_arr_VI_Object(VI_Object{var_name: args[0],object_type: get_plain_type("string")},symbol_table,"var_name")
			if var_1["error"]==1 {
				fmt.Println("Data types did not match")
				return
			}
			if var_1["result"]==0 {
				var_1["index"]=len(symbol_table)
				symbol_table = append(symbol_table, VI_Object{var_name: args[0],object_type: get_plain_type("string")})
			}
			current_byte_code = append(current_byte_code, 12, var_1["index"], int(num))
			byte_code = append(byte_code, current_byte_code)
		case "str.add","str.mult":
			var_2_type:=make([]string,0)
			switch opcode {
			case "str.add":var_2_type=get_plain_type("string")
			case "str.mult":var_2_type=get_plain_type("num")
			}
			var_1:=plain_in_arr_VI_Object(VI_Object{var_name: args[0],object_type: get_plain_type("string")},symbol_table,"var_name")
			var_2:=plain_in_arr_VI_Object(VI_Object{var_name: args[1],object_type: var_2_type},symbol_table,"var_name")
			var_res:=plain_in_arr_VI_Object(VI_Object{var_name: args[2],object_type: get_plain_type("string")},symbol_table,"var_name")
			if var_1["error"]==1 || var_2["error"]==1 || var_res["error"]==1 {
				fmt.Println("Data types did not match")
				return
			}
			if var_1["index"]==-1 {
				fmt.Println("Variable",args[0],"has not been initialised yet")
				return
			}
			if var_2["index"]==-1 {
				fmt.Println("Variable",args[1],"has not been initialised yet")
				return
			}
			if var_res["index"]==-1 {
				fmt.Println("Variable",args[2],"has not been initialised yet")
				return
			}
			var intopcode int;
			switch opcode {
			case "str.add":intopcode=14
			case "str.mult":intopcode=15
			}
			current_byte_code = append(current_byte_code, intopcode, var_1["index"], var_2["index"], var_res["index"])
			byte_code = append(byte_code, current_byte_code)
		case "str.refset":
			set_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[0],object_type: get_plain_type("string")},symbol_table,"var_name")
			reference:=plain_in_arr_VI_Object(VI_Object{var_name: args[1],object_type: get_plain_type("string")},symbol_table,"var_name")
			if set_var["index"]==-1 {
				fmt.Println("Variable",args[0],"has not been initialised yet")
				return
			}
			if set_var["error"]==1 {
				fmt.Println("Data types did not match")
				return
			}
			if reference["index"]==-1 {
				fmt.Println("Variable",args[1],"has not been initialised yet")
				return
			}
			if reference["error"]==1 {
				fmt.Println("Data types did not match")
				return
			}
			var intopcode int;
			intopcode=16
			current_byte_code = append(current_byte_code, intopcode, set_var["index"], reference["index"])
			byte_code = append(byte_code, current_byte_code)
		case "define.jump":
			num,err:=strconv.ParseInt(args[0],10,64)
			if err!=nil {
				fmt.Println(err)
				return
			}
			jump_table[string_consts[num]]=i
			current_byte_code = append(current_byte_code, 17,int(num),i)
			byte_code = append(byte_code, current_byte_code)
		case "jump.def":
			num,err:=strconv.ParseInt(args[0],10,64)
			condition_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[1],object_type: get_plain_type("num")},symbol_table,"var_name")
			if condition_var["index"]==-1 {
				fmt.Println("Variable",args[1],"has not been initialised yet")
				return
			}
			if condition_var["error"]==1 {
				fmt.Println("Data types did not match")
				return
			}
			if err!=nil {
				fmt.Println(err)
				return
			}
			current_byte_code = append(current_byte_code, 18,int(num),condition_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "endx":
			current_byte_code = append(current_byte_code, 19)
			byte_code = append(byte_code, current_byte_code)
		case "arr.init":
			num,err:=strconv.ParseFloat(args[1],64)
			if err!=nil {
				fmt.Println(err)
				return
			}
			arr_type:=get_init_arr_type(string_consts[int(num)])
			if !type_evaluator(arr_type) {
				fmt.Println("Invalid type for initialising an array")
				return
			}
			index:=plain_in_arr_VI_Object(VI_Object{var_name: args[0],object_type: arr_type},arr_constants,"var_name")
			if index["error"]==1 {
				fmt.Println("Data types did not match")
				return
			}
			arr_default:=VI_Object{var_name: args[0],object_type: arr_type}
			if index["result"]==0 {
				index["index"]=len(arr_constants)
				arr_constants = append(arr_constants, arr_default)
			}
			res:=plain_in_arr_VI_Object(arr_default,symbol_table,"var_name")
			if res["error"]==1 {
				fmt.Println("Data types did not match")
				return
			}
			if res["result"]==0 {
				symbol_table = append(symbol_table, VI_Object{var_name: args[0],object_type: arr_type})
				res["index"]=len(symbol_table)-1
			}
			current_byte_code = append(current_byte_code, 25,res["index"],index["index"])
			byte_code = append(byte_code, current_byte_code)
		case "arr.push":
			arr_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[0]},symbol_table,"var_name")
			if arr_var["result"]==0 {
				fmt.Println("Variable",args[0],"of type array has not been initialised yet")
				return
			}
			if symbol_table[arr_var["index"]].object_type[0]!="arr" {
				fmt.Println("Variable",args[0],"is not an array")
				return
			}
			push_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[1]},symbol_table,"var_name")
			if push_var["result"]==0 {
				fmt.Println("Variable",args[1],"has not been initialised yet")
				return
			}
			if !obj_supports_type(symbol_table[arr_var["index"]].object_type,symbol_table[push_var["index"]].object_type) {
				fmt.Println("Object type is not supported by array",args[0])
				return
			}
			current_byte_code = append(current_byte_code, 26,arr_var["index"],push_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "arr.pull","arr.index.set":
			arr_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[0]},symbol_table,"var_name")
			if arr_var["result"]==0 {
				fmt.Println("Variable",args[0],"of type array has not been initialised yet")
				return
			}
			if symbol_table[arr_var["index"]].object_type[0]!="arr" {
				fmt.Println("Variable",args[0],"is not an array")
				return
			}
			index_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[1],object_type: get_plain_type("num")},symbol_table,"var_name")
			if index_var["result"]==0 {
				fmt.Println("Variable",args[1],"has not been initialised yet")
				return
			}
			if index_var["error"]==1 {
				fmt.Println("Index variable needs to be a number")
				return
			}
			pull_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[2]},symbol_table,"var_name")
			if pull_var["result"]==0 {
				fmt.Println("Variable",args[2],"has not been initialised yet")
				return
			}
			if !obj_supports_type(symbol_table[arr_var["index"]].object_type,symbol_table[pull_var["index"]].object_type) {
				fmt.Println("Object type does not match array",args[0],"object type")
				return
			}
			opcode_num:=0
			switch opcode {
			case "arr.pull":opcode_num=27
			case "arr.index.set":opcode_num=29
			}
			current_byte_code = append(current_byte_code, opcode_num,arr_var["index"],index_var["index"],pull_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "arr.remove","arr.length":
			arr_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[0]},symbol_table,"var_name")
			if arr_var["result"]==0 {
				fmt.Println("Variable",args[0],"of type array has not been initialised yet")
				return
			}
			if symbol_table[arr_var["index"]].object_type[0]!="arr" {
				fmt.Println("Variable",args[0],"is not an array")
				return
			}
			index_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[1],object_type: get_plain_type("num")},symbol_table,"var_name")
			if index_var["result"]==0 {
				fmt.Println("Variable",args[1],"has not been initialised yet")
				return
			}
			if index_var["error"]==1 {
				fmt.Println("Array index is required to be of type number")
				return
			}
			intopcode:=0
			switch opcode {
			case "arr.remove":intopcode=28
			case "arr.length":intopcode=30
			}
			current_byte_code = append(current_byte_code, intopcode,arr_var["index"],index_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "arr.refset":
			arr1_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[0]},symbol_table,"var_name")
			if arr1_var["result"]==0 {
				fmt.Println("Variable",args[0],"of type array has not been initialised yet")
				return
			}
			if symbol_table[arr1_var["index"]].object_type[0]!="arr" {
				fmt.Println("Variable",args[0],"is not an array")
				return
			}
			arr2_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[1],object_type: get_plain_type("num")},symbol_table,"var_name")
			if arr2_var["result"]==0 {
				fmt.Println("Variable",args[1],"has not been initialised yet")
				return
			}
			if (!(string_arr_compare(symbol_table[arr1_var["index"]].object_type,symbol_table[arr2_var["index"]].object_type))) {
				fmt.Println("Both arrays must be of same type")
				return
			}
			current_byte_code = append(current_byte_code, 31,arr1_var["index"],arr2_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "arr.includes":
			arr_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[0]},symbol_table,"var_name")
			if arr_var["result"]==0 {
				fmt.Println("Variable",args[0],"of type array has not been initialised yet")
				return
			}
			if symbol_table[arr_var["index"]].object_type[0]!="arr" {
				fmt.Println("Variable",args[0],"is not an array")
				return
			}
			check_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[1]},symbol_table,"var_name")
			if check_var["result"]==0 {
				fmt.Println("Variable",args[1],"has not been initialised yet")
				return
			}
			index_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[2],object_type: get_plain_type("num")},symbol_table,"var_name")
			if index_var["result"]==0 {
				fmt.Println("Variable",args[2],"has not been initialised yet")
				return
			}
			if index_var["error"]==1 {
				fmt.Println("Result index variable needs to be a number")
				return
			}
			if !obj_supports_type(symbol_table[arr_var["index"]].object_type,symbol_table[check_var["index"]].object_type) {
				fmt.Println("Object type does not match array",args[0],"object type")
				return
			}
			current_byte_code = append(current_byte_code, 32,arr_var["index"],check_var["index"],index_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "obj.equals":
			obj1_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[0]},symbol_table,"var_name")
			if obj1_var["result"]==0 {
				fmt.Println("Variable",args[0],"of type array has not been initialised yet")
				return
			}
			obj2_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[1]},symbol_table,"var_name")
			if obj2_var["result"]==0 {
				fmt.Println("Variable",args[1],"has not been initialised yet")
				return
			}
			if (!(string_arr_compare(symbol_table[obj1_var["index"]].object_type,symbol_table[obj2_var["index"]].object_type))) {
				fmt.Println("Both objects must be of same type")
				return
			}
			res_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[2], object_type: get_plain_type("num")},symbol_table,"var_name")
			if res_var["result"]==0 {
				fmt.Println("Variable",args[2],"has not been initialised yet")
				return
			}
			if res_var["error"]==1 {
				fmt.Println("Result variable needs to be a number")
				return
			}
			current_byte_code = append(current_byte_code, 33,obj1_var["index"],obj2_var["index"],res_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "str.equals":
			var_1:=plain_in_arr_VI_Object(VI_Object{var_name: args[0],object_type: get_plain_type("string")},symbol_table,"var_name")
			var_2:=plain_in_arr_VI_Object(VI_Object{var_name: args[1],object_type: get_plain_type("string")},symbol_table,"var_name")
			var_res:=plain_in_arr_VI_Object(VI_Object{var_name: args[2],object_type: get_plain_type("num")},symbol_table,"var_name")
			if var_1["error"]==1 || var_2["error"]==1 || var_res["error"]==1 {
				fmt.Println("Data types did not match")
				return
			}
			if var_1["index"]==-1 {
				fmt.Println("Variable",args[0],"has not been initialised yet")
				return
			}
			if var_2["index"]==-1 {
				fmt.Println("Variable",args[1],"has not been initialised yet")
				return
			}
			if var_res["index"]==-1 {
				fmt.Println("Variable",args[2],"has not been initialised yet")
				return
			}
			current_byte_code = append(current_byte_code, 34, var_1["index"], var_2["index"], var_res["index"])
			byte_code = append(byte_code, current_byte_code)
		case "dict.init":
			num,err:=strconv.ParseFloat(args[1],64)
			if err!=nil {
				fmt.Println(err)
				return
			}
			dict_type:=get_init_dict_type(string_consts[int(num)])
			if !type_evaluator(dict_type) {
				fmt.Println("Invalid type for initialising a dict")
				return
			}
			index:=plain_in_arr_VI_Object(VI_Object{var_name: args[0],object_type: dict_type},dict_constants,"var_name")
			if index["error"]==1 {
				fmt.Println("Data types did not match")
				return
			}
			dict_default:=VI_Object{var_name: args[0],object_type: dict_type}
			if index["result"]==0 {
				index["index"]=len(dict_constants)
				dict_constants = append(dict_constants, dict_default)
			}
			res:=plain_in_arr_VI_Object(dict_default,symbol_table,"var_name")
			if res["error"]==1 {
				fmt.Println("Data types did not match")
				return
			}
			if res["result"]==0 {
				symbol_table = append(symbol_table, VI_Object{var_name: args[0],object_type: dict_type})
				res["index"]=len(symbol_table)-1
			}
			current_byte_code = append(current_byte_code, 35,res["index"],index["index"])
			byte_code = append(byte_code, current_byte_code)
		case "dict.key.set","dict.pull":
			dict_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[0]},symbol_table,"var_name")
			if dict_var["result"]==0 {
				fmt.Println("Variable",args[0],"of type dict has not been initialised yet")
				return
			}
			if symbol_table[dict_var["index"]].object_type[0]!="dict" {
				fmt.Println("Variable",args[0],"is not a dict")
				return
			}
			key_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[1],object_type: get_plain_type("string")},symbol_table,"var_name")
			if key_var["error"]==1 {
				fmt.Println("Data types did not match")
			}
			if key_var["result"]==0 {
				fmt.Println("Variable",args[0],"of type string has not been initialised yet")
				return
			}
			value_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[2]},symbol_table,"var_name")
			if value_var["result"]==0 {
				fmt.Println("Variable",args[2],"has not been initialised yet")
				return
			}
			if !obj_supports_type(symbol_table[dict_var["index"]].object_type,symbol_table[value_var["index"]].object_type) {
				fmt.Println("Object type is not supported by dict",args[0])
				return
			}
			var int_opcode int;
			switch opcode {
			case "dict.key.set":int_opcode=36
			case "dict.pull":int_opcode=37
			}
			current_byte_code = append(current_byte_code, int_opcode,dict_var["index"],key_var["index"],value_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "dict.delete":
			dict_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[0]},symbol_table,"var_name")
			if dict_var["result"]==0 {
				fmt.Println("Variable",args[0],"of type dict has not been initialised yet")
				return
			}
			if symbol_table[dict_var["index"]].object_type[0]!="dict" {
				fmt.Println("Variable",args[0],"is not a dict")
				return
			}
			key_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[1],object_type: get_plain_type("string")},symbol_table,"var_name")
			if key_var["error"]==1 {
				fmt.Println("Data types did not match")
			}
			if key_var["result"]==0 {
				fmt.Println("Variable",args[0],"of type string has not been initialised yet")
				return
			}
			current_byte_code = append(current_byte_code, 38,dict_var["index"],key_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "dict.keys":
			dict_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[0]},symbol_table,"var_name")
			if dict_var["result"]==0 {
				fmt.Println("Variable",args[0],"of type dict has not been initialised yet")
				return
			}
			if symbol_table[dict_var["index"]].object_type[0]!="dict" {
				fmt.Println("Variable",args[0],"is not a dict")
				return
			}
			arr_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[1],object_type: get_init_arr_type("string")},symbol_table,"var_name")
			if arr_var["error"]==1 {
				fmt.Println("Data types did not match")
			}
			if arr_var["result"]==0 {
				fmt.Println("Variable",args[0],"of type array[string] has not been initialised yet")
				return
			}
			current_byte_code = append(current_byte_code, 39,dict_var["index"],arr_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "dict.refset":
			dict1_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[0]},symbol_table,"var_name")
			if dict1_var["result"]==0 {
				fmt.Println("Variable",args[0],"of type dict has not been initialised yet")
				return
			}
			if symbol_table[dict1_var["index"]].object_type[0]!="dict" {
				fmt.Println("Variable",args[0],"is not a dict")
				return
			}
			dict2_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[1]},symbol_table,"var_name")
			if dict2_var["result"]==0 {
				fmt.Println("Variable",args[1],"of type dict has not been initialised yet")
				return
			}
			if symbol_table[dict2_var["index"]].object_type[0]!="dict" {
				fmt.Println("Variable",args[1],"is not a dict")
				return
			}
			if !string_arr_compare(symbol_table[dict1_var["index"]].object_type,symbol_table[dict2_var["index"]].object_type) {
				fmt.Println("Dictionaries are of different types")
				return
			}
			current_byte_code = append(current_byte_code, 40,dict1_var["index"],dict2_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "dict.key.includes":
			dict_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[0]},symbol_table,"var_name")
			if dict_var["result"]==0 {
				fmt.Println("Variable",args[0],"of type dict has not been initialised yet")
				return
			}
			if symbol_table[dict_var["index"]].object_type[0]!="dict" {
				fmt.Println("Variable",args[0],"is not a dict")
				return
			}
			key_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[1],object_type: get_plain_type("string")},symbol_table,"var_name")
			if key_var["error"]==1 {
				fmt.Println("Data types did not match")
			}
			if key_var["result"]==0 {
				fmt.Println("Variable",args[0],"of type string has not been initialised yet")
				return
			}
			value_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[2],object_type: get_plain_type("num")},symbol_table,"var_name")
			if value_var["error"]==1 {
				fmt.Println("Data types did not match")
			}
			if value_var["result"]==0 {
				fmt.Println("Variable",args[2],"has not been initialised yet")
				return
			}
			current_byte_code = append(current_byte_code, 41,dict_var["index"],key_var["index"],value_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "str.includes":
			str1_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[0],object_type: get_plain_type("string")},symbol_table,"var_name")
			if str1_var["error"]==1 {
				fmt.Println("Data types did not match")
			}
			if str1_var["result"]==0 {
				fmt.Println("Variable",args[0],"of type string has not been initialised yet")
				return
			}
			str2_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[1],object_type: get_plain_type("string")},symbol_table,"var_name")
			if str2_var["error"]==1 {
				fmt.Println("Data types did not match")
			}
			if str2_var["result"]==0 {
				fmt.Println("Variable",args[1],"of type string has not been initialised yet")
				return
			}
			value_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[2],object_type: get_plain_type("num")},symbol_table,"var_name")
			if value_var["error"]==1 {
				fmt.Println("Data types did not match")
			}
			if value_var["result"]==0 {
				fmt.Println("Variable",args[2],"has not been initialised yet")
				return
			}
			current_byte_code = append(current_byte_code, 42,str1_var["index"],str2_var["index"],value_var["index"])
			byte_code = append(byte_code, current_byte_code)
		}
	}
	fmt.Println(byte_code)
	continute_exec:=true
	for i := 0; i < len(byte_code); i++ {
		current_gas+=1
		if current_gas<gas_limit || !continute_exec {
			break
		}
		current_byte_code:=byte_code[i]
		switch opcode:=current_byte_code[0]; opcode {
		case 0:
			symbol_table[current_byte_code[1]].num_value=num_constants[current_byte_code[2]]
		case 1:
			symbol_table[current_byte_code[1]].num_value=symbol_table[current_byte_code[2]].num_value
		case 2:
			if symbol_table[current_byte_code[2]].num_value!=0 {
				i=int(symbol_table[current_byte_code[1]].num_value)-3
			}
		case 3:
			if symbol_table[current_byte_code[1]].num_value == symbol_table[current_byte_code[2]].num_value {
				symbol_table[current_byte_code[3]].num_value=1
			} else {
				symbol_table[current_byte_code[3]].num_value=0
			}
		case 4:
			if symbol_table[current_byte_code[1]].num_value > symbol_table[current_byte_code[2]].num_value {
				symbol_table[current_byte_code[3]].num_value=1
			} else {
				symbol_table[current_byte_code[3]].num_value=0
			}
		case 5:
			symbol_table[current_byte_code[3]].num_value=symbol_table[current_byte_code[1]].num_value + symbol_table[current_byte_code[2]].num_value
		case 6:
			symbol_table[current_byte_code[3]].num_value=symbol_table[current_byte_code[1]].num_value - symbol_table[current_byte_code[2]].num_value
		case 7:
			symbol_table[current_byte_code[3]].num_value=symbol_table[current_byte_code[1]].num_value * symbol_table[current_byte_code[2]].num_value
		case 8:
			symbol_table[current_byte_code[3]].num_value=symbol_table[current_byte_code[1]].num_value / symbol_table[current_byte_code[2]].num_value
		case 9:
			symbol_table[current_byte_code[3]].num_value=math.Pow(symbol_table[current_byte_code[1]].num_value,symbol_table[current_byte_code[2]].num_value)
		case 10:
			symbol_table[current_byte_code[3]].num_value=math.Floor(symbol_table[current_byte_code[1]].num_value / symbol_table[current_byte_code[2]].num_value)
		case 11:
			symbol_table[current_byte_code[3]].num_value=math.Mod(symbol_table[current_byte_code[1]].num_value, symbol_table[current_byte_code[2]].num_value)
		case 12:
			symbol_table[current_byte_code[1]].str_value=string_consts[current_byte_code[2]]
		case 14:
			symbol_table[current_byte_code[3]].str_value=symbol_table[current_byte_code[1]].str_value + symbol_table[current_byte_code[2]].str_value
		case 15:
			symbol_table[current_byte_code[3]].str_value=strings.Repeat(symbol_table[current_byte_code[1]].str_value, int(symbol_table[current_byte_code[2]].num_value))
		case 16:
			symbol_table[current_byte_code[1]].str_value=symbol_table[current_byte_code[2]].str_value
		case 17:
			jump_table[string_consts[current_byte_code[1]]]=current_byte_code[2]
		case 18:
			if (symbol_table[current_byte_code[2]].num_value!=0) {
				i=jump_table[string_consts[current_byte_code[1]]]
			}
		case 19:
			continute_exec=false
		case 20:
			symbol_table[current_byte_code[3]].num_value=bool_to_num(num_to_bool(symbol_table[current_byte_code[1]].num_value) && num_to_bool(symbol_table[current_byte_code[2]].num_value))
		case 21:
			symbol_table[current_byte_code[3]].num_value=bool_to_num(num_to_bool(symbol_table[current_byte_code[1]].num_value) || num_to_bool(symbol_table[current_byte_code[2]].num_value))
		case 22:
			symbol_table[current_byte_code[2]].num_value=bool_to_num(!num_to_bool(symbol_table[current_byte_code[1]].num_value))
		case 23:
			symbol_table[current_byte_code[3]].num_value=float64(int(symbol_table[current_byte_code[1]].num_value) ^ int(symbol_table[current_byte_code[2]].num_value))
		case 24:
			symbol_table[current_byte_code[3]].num_value=round_float64(symbol_table[current_byte_code[1]].num_value,uint(symbol_table[current_byte_code[2]].num_value))
		case 25:
			symbol_table[current_byte_code[1]]=arr_constants[current_byte_code[2]]
		case 26:
			symbol_table[current_byte_code[1]].children = append(symbol_table[current_byte_code[1]].children, symbol_table[current_byte_code[2]])
		case 27:
			temp_var:=symbol_table[current_byte_code[1]].children[int(symbol_table[current_byte_code[2]].num_value)]
			temp_var.var_name=symbol_table[current_byte_code[3]].var_name
			symbol_table[current_byte_code[3]]=temp_var
		case 28:
			symbol_table[current_byte_code[1]].children=remove_VI_Object_from_index(symbol_table[current_byte_code[1]].children,int(symbol_table[current_byte_code[2]].num_value))
		case 29:
			symbol_table[current_byte_code[1]].children[int(symbol_table[current_byte_code[2]].num_value)]=symbol_table[current_byte_code[3]]
		case 30:
			symbol_table[current_byte_code[2]].num_value=float64(len(symbol_table[current_byte_code[1]].children))
		case 31:
			symbol_table[current_byte_code[1]].children=symbol_table[current_byte_code[2]].children
		case 32:
			arr_var:=symbol_table[current_byte_code[1]].children
			check_var:=symbol_table[current_byte_code[2]]
			res_index:=-1
			for i := 0; i < len(symbol_table[current_byte_code[1]].children); i++ {
				if (recursive_VI_Object_match(arr_var[i],check_var)) {
					res_index=i
					break
				}
			}
			symbol_table[current_byte_code[3]].num_value=float64(res_index)
		case 33:
			symbol_table[current_byte_code[3]].num_value=bool_to_num(recursive_VI_Object_match(symbol_table[current_byte_code[1]],symbol_table[current_byte_code[2]]))
		case 34:
			symbol_table[current_byte_code[3]].num_value=bool_to_num(symbol_table[current_byte_code[1]].str_value==symbol_table[current_byte_code[2]].str_value)
		case 35:
			symbol_table[current_byte_code[1]]=dict_constants[current_byte_code[2]]
		case 36:
			index:=str_index_in_arr(symbol_table[current_byte_code[2]].str_value,symbol_table[current_byte_code[1]].dict_keys)
			if index==-1 {
				symbol_table[current_byte_code[1]].dict_keys = append(symbol_table[current_byte_code[1]].dict_keys, symbol_table[current_byte_code[2]].str_value)
				symbol_table[current_byte_code[1]].children = append(symbol_table[current_byte_code[1]].children, symbol_table[current_byte_code[3]])
			} else {
				symbol_table[current_byte_code[1]].children[index]=symbol_table[current_byte_code[3]]
			}
		case 37:
			temp_pull_index:=str_index_in_arr(symbol_table[current_byte_code[2]].str_value,symbol_table[current_byte_code[1]].dict_keys)
			temp_pull:=symbol_table[current_byte_code[1]].children[temp_pull_index]
			temp_pull.var_name=symbol_table[current_byte_code[3]].var_name
			symbol_table[current_byte_code[3]]=temp_pull
		case 38:
			temp_pull_index:=str_index_in_arr(symbol_table[current_byte_code[2]].str_value,symbol_table[current_byte_code[1]].dict_keys)
			symbol_table[current_byte_code[1]].dict_keys=remove_string_from_index(symbol_table[current_byte_code[1]].dict_keys,temp_pull_index)
			symbol_table[current_byte_code[1]].children=remove_VI_Object_from_index(symbol_table[current_byte_code[1]].children,temp_pull_index)
		case 39:
			keys_arr_VI_Object:=make([]VI_Object,0)
			for i := 0; i < len(symbol_table[current_byte_code[1]].dict_keys); i++ {
				keys_arr_VI_Object = append(keys_arr_VI_Object, VI_Object{object_type: get_plain_type("string"),str_value: symbol_table[current_byte_code[1]].dict_keys[i]})
			}
			symbol_table[current_byte_code[2]].children=keys_arr_VI_Object
		case 40:
			symbol_table[current_byte_code[1]].children=symbol_table[current_byte_code[2]].children
			symbol_table[current_byte_code[1]].dict_keys=symbol_table[current_byte_code[2]].dict_keys
		case 41:
			symbol_table[current_byte_code[3]].num_value=bool_to_num(str_index_in_arr(symbol_table[current_byte_code[2]].str_value,symbol_table[current_byte_code[1]].dict_keys)!=-1)
		case 42:
			symbol_table[current_byte_code[3]].num_value=bool_to_num(strings.Contains(symbol_table[current_byte_code[1]].str_value,symbol_table[current_byte_code[2]].str_value))
		}
	}
	fmt.Println(symbol_table,num_constants)
}