// make sure that you do run the smart contract code while verifying so that just in case the vengine crashes you can mark the blocvk as false

package main

import (
	"math"
	"strconv"
	"strings"
	"fmt"
	"reflect"
)

var Debug bool = false;
var registered_structs = map[string][]string{"structs":[]string{"num","string"}}

func contains_float64(a float64, list []float64) map[string]int {
	res := make(map[string]int)
	for i, b := range list {
		if b == a {
			res["index"] = i
			res["contains"] = 1
			return res
		}
	}
	res["contains"] = 0
	res["index"] = -1
	return res
}

func contains_str(a string, list []string) map[string]int64 {
	res := make(map[string]int64)
	for i, b := range list {
		if b == a {
			res["index"] = int64(i)
			res["contains"] = 1
			return res
		}
	}
	res["contains"] = 0
	res["index"] = -1
	return res
}

func plain_in_arr_VI_Object(obj VI_Object, arr []VI_Object, field string) map[string]int {
	res := make(map[string]int)
	for i := 0; i < len(arr); i++ {
		current_obj := arr[i]
		res["error"] = 0
		if strings.Join(current_obj.object_type[:], ",") != strings.Join(obj.object_type[:], ",") {
			res["error"] = 1
		}
		switch field {
		case "var_name":
			if current_obj.var_name == obj.var_name {
				res["index"] = i
				res["result"] = 1
				return res
			}
		case "num_value":
			if current_obj.num_value == obj.num_value {
				res["index"] = i
				res["result"] = 1
				return res
			}
		case "str_value":
			if current_obj.str_value == obj.str_value {
				res["index"] = i
				res["result"] = 1
				return res
			}
		}
	}
	res["index"] = -1
	res["result"] = 0
	res["error"] = 0
	return res
}

func string_arr_compare(arr1 []string, arr2 []string) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for i := 0; i < len(arr1); i++ {
		if arr1[i] != arr2[i] {
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
		if obj_type=="dict" {
			for i := 0; i < len(obj1.dict_keys); i++ {
				key_index:=str_index_in_arr(obj1.dict_keys[i], obj2.dict_keys)
				if key_index==-1 || !recursive_VI_Object_match(obj1.children[i],obj2.children[key_index]) {
					return false
				}
			}
		} else {
			for i := 0; i < len(obj1.children); i++ {
				if (!recursive_VI_Object_match(obj1.children[i],obj2.children[i])) {
					return false
				}
			}
		}
		return true
	default:
		return reflect.DeepEqual(obj1._struct, obj2._struct)
	}
	return true
}

func str_index_in_arr(str_obj string, string_arr []string) int {
	for i := 0; i < len(string_arr); i++ {
		if string_arr[i] == str_obj {
			return i
		}
	}
	return -1
}

func get_plain_type(obj_type string) []string {
	res := make([]string, 0)
	res = append(res, obj_type)
	return res
}

func get_init_arr_type(obj_type string) []string {
	res := make([]string, 0)
	res = append(res, "arr")
	res = append(res, strings.Split(strings.ReplaceAll(obj_type, " ", ""), ",")...)
	return res
}

func get_init_dict_type(obj_type string) []string {
	res := make([]string, 0)
	res = append(res, "dict")
	res = append(res, strings.Split(strings.ReplaceAll(obj_type, " ", ""), ",")...)
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
	var_name    string
	num_value   float64
	str_value   string
	object_type []string
	children    []VI_Object
	dict_keys   []string
	scope       int
	_struct     map[int]VI_Object
}

func (object *VI_Object) fill_defaults() {}

func round_float64(num float64, decimals uint) float64 {
	ratio := math.Pow(10, float64(decimals))
	return math.Round(num*ratio) / ratio
}

func str_parser(dat string) string {
	output := ""
	for i := 0; i < len(dat); i++ {
		if string(dat[i]) == "\\" {
			if i+1 == len(dat) {
				output += "\\"
			} else {
				switch string(dat[i+1]) {
				case "n":
					i += 1
					output += "\n"
				default:
					output += "\\"
				}
			}
		} else {
			output += string(dat[i])
		}
	}
	return output
}

func file_parser(dat string) map[string][]string {
	data_lines := strings.Split(dat, "\n")
	mode := ""
	code_lines := make([]string, 0)
	data_constants := make([]string, 0)
	for i := 0; i < len(data_lines); i++ {
		current_line := strings.Trim(strings.Trim(strings.Trim(data_lines[i], "\r"), "	"), "	")
		if mode=="code" {
			current_line=strings.Split(current_line, "//")[0]
		}
		if current_line == "" {
			continue
		}
		if current_line == ".code" {
			mode = "code"
			continue
		} else if current_line == ".data" {
			mode = "data"
			continue
		}
		if mode == "data" {
			data_string := strings.Trim(current_line, " ")[1:]
			data_string = str_parser(data_string[:len(data_string)-1])
			data_constants = append(data_constants, data_string)
		} else if mode == "code" {
			code_lines = append(code_lines, strings.Trim(current_line, " "))
		}
	}
	res := make(map[string][]string, 0)
	res["code_lines"] = code_lines
	res["data_constants"] = data_constants
	return res
}

func num_to_bool(num float64) bool {
	if num == 0 {
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

func obj_supports_type(arr_type []string, obj_type []string) bool {
	if strings.Join(arr_type[1:], ",") == strings.Join(obj_type, ",") {
		return true
	} else {
		return false
	}
}

func type_evaluator(obj_type []string) bool {
	if len(obj_type)==0 {
		return false
	}
	if str_index_in_arr(obj_type[len(obj_type)-1], registered_structs["structs"])==-1 {
		return false
	}
	check_arr := remove_string_from_index(obj_type, len(obj_type)-1)
	for i := 0; i < len(check_arr); i++ {
		if check_arr[i] != "dict" && check_arr[i] != "arr" {
			return false
		}
	}
	return true
}

func Debug_print(a ...any) {
	if Debug {
		fmt.Println(a...)
	}
}

func Debug_printf(a ...any) {
	if Debug {
		fmt.Print(a...)
	}
}

func spawn_struct_default(var_name string, object_type string) VI_Object {
	struct_default:=VI_Object{var_name: var_name, object_type: []string{object_type}, _struct: map[int]VI_Object{}}
	for i,value:=range registered_structs[object_type] {
		raw_type:=strings.Split(strings.Split(value, "->")[1], ",")
		if str_index_in_arr(raw_type[0], []string{"arr","dict","string","num"})==-1 && len(raw_type)==1 {
			struct_default._struct[i]=spawn_struct_default(strings.Split(value, "->")[0], raw_type[0])
		} else {
			struct_default._struct[i]=VI_Object{var_name: strings.Split(value, "->")[0], object_type: raw_type}
		}
	}
	return struct_default
}

func copy_VI_Object(a VI_Object) VI_Object {
	new_children:=make([]VI_Object, 0)
	for _,child:=range a.children {
		new_children = append(new_children, copy_VI_Object(child))
	}
	new_dict_keys:=make([]string, 0)
	for _,dict_key:=range a.dict_keys {
		new_dict_keys = append(new_dict_keys, dict_key)
	}
	new_struct:=make(map[int]VI_Object)
	for i,struct_:=range a._struct {
		new_struct[i]=struct_
	}
	return VI_Object{var_name: strings.Clone(a.var_name), num_value: (float64(a.num_value)+-1)+1, str_value: strings.Clone(a.str_value), object_type: a.object_type, children: new_children, scope: int(int64(a.scope)), _struct: new_struct}
}

func Vengine(code string, debug bool) int64 {
	Debug=debug
	parse_results := file_parser(code)
	codex, string_consts := parse_results["code_lines"], parse_results["data_constants"]

	var num_constants []float64
	var arr_constants []VI_Object
	var dict_constants []VI_Object

	byte_code := make([][]int, 0)
	global_table := make([][]VI_Object, 0)
	global_table = append(global_table, make([]VI_Object, 0))
	symbol_table := global_table[0]
	scope_count := 0
	jump_table := make(map[int]int, 0)
	gas_limit := int64(0)
	current_gas := int64(0)

	for i := 0; i < len(codex); i++ {
		args := strings.Split(codex[i], " ")
		if len(args) >= 2 {
			args = strings.Split(args[1], ",")
		} else {
			args = make([]string, 0)
		}
		current_byte_code := make([]int, 0)
		switch opcode := strings.Split(codex[i], " ")[0]; opcode {
		case "set":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			num, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				Debug_print(err)
				return current_gas
			}
			index := contains_float64(num, num_constants)
			if index["contains"] == 0 {
				num_constants = append(num_constants, num)
				index["index"] = len(num_constants) - 1
			}
			res := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if res["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if res["result"] == 0 {
				symbol_table = append(symbol_table, VI_Object{var_name: args[0], object_type: get_plain_type("num")})
				res["index"] = len(symbol_table) - 1
			}
			current_byte_code = append(current_byte_code, 0, res["index"], index["index"])
			byte_code = append(byte_code, current_byte_code)
		case "refset", "jump", "not":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			set_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("num")}, symbol_table, "var_name")
			reference := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if set_var["index"] == -1 {
				Debug_print("Variable", args[0], "has not been initialised yet")
				return current_gas
			}
			if set_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if reference["index"] == -1 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			if reference["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			var intopcode int
			switch opcode {
			case "refset":
				intopcode = 1
			case "jump":
				intopcode = 2
			case "not":
				intopcode = 22
			}
			current_byte_code = append(current_byte_code, intopcode, set_var["index"], reference["index"])
			byte_code = append(byte_code, current_byte_code)
		case "equals", "greater", "add", "sub", "mult", "div", "floor", "mod", "power", "round", "and", "or", "xor","smaller":
			if len(args) != 3 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			var_1 := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("num")}, symbol_table, "var_name")
			var_2 := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("num")}, symbol_table, "var_name")
			var_res := plain_in_arr_VI_Object(VI_Object{var_name: args[2], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if var_1["error"] == 1 || var_2["error"] == 1 || var_res["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if var_1["index"] == -1 {
				Debug_print("Variable", args[0], "has not been initialised yet")
				return current_gas
			}
			if var_2["index"] == -1 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			if var_res["index"] == -1 {
				Debug_print("Variable", args[2], "has not been initialised yet")
				return current_gas
			}
			var intopcode int
			switch opcode {
			case "equals":
				intopcode = 3
			case "greater":
				intopcode = 4
			case "add":
				intopcode = 5
			case "sub":
				intopcode = 6
			case "mult":
				intopcode = 7
			case "div":
				intopcode = 8
			case "power":
				intopcode = 9
			case "floor":
				intopcode = 10
			case "mod":
				intopcode = 11
			case "round":
				intopcode = 24
			case "and":
				intopcode = 20
			case "or":
				intopcode = 21
			case "xor":
				intopcode = 23
			case "smaller":
				intopcode = 64
			}
			current_byte_code = append(current_byte_code, intopcode, var_1["index"], var_2["index"], var_res["index"])
			byte_code = append(byte_code, current_byte_code)
		case "str.set":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			num, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				Debug_print(err)
				return current_gas
			}
			var_1 := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if var_1["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if var_1["result"] == 0 {
				var_1["index"] = len(symbol_table)
				symbol_table = append(symbol_table, VI_Object{var_name: args[0], object_type: get_plain_type("string")})
			}
			current_byte_code = append(current_byte_code, 12, var_1["index"], int(num))
			byte_code = append(byte_code, current_byte_code)
		case "str.add", "str.mult":
			if len(args) != 3 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			var_2_type := make([]string, 0)
			switch opcode {
			case "str.add":
				var_2_type = get_plain_type("string")
			case "str.mult":
				var_2_type = get_plain_type("num")
			}
			var_1 := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("string")}, symbol_table, "var_name")
			var_2 := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: var_2_type}, symbol_table, "var_name")
			var_res := plain_in_arr_VI_Object(VI_Object{var_name: args[2], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if var_1["error"] == 1 || var_2["error"] == 1 || var_res["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if var_1["index"] == -1 {
				Debug_print("Variable", args[0], "has not been initialised yet")
				return current_gas
			}
			if var_2["index"] == -1 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			if var_res["index"] == -1 {
				Debug_print("Variable", args[2], "has not been initialised yet")
				return current_gas
			}
			var intopcode int
			switch opcode {
			case "str.add":
				intopcode = 14
			case "str.mult":
				intopcode = 15
			}
			current_byte_code = append(current_byte_code, intopcode, var_1["index"], var_2["index"], var_res["index"])
			byte_code = append(byte_code, current_byte_code)
		case "str.refset":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			set_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("string")}, symbol_table, "var_name")
			reference := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if set_var["index"] == -1 {
				Debug_print("Variable", args[0], "has not been initialised yet")
				return current_gas
			}
			if set_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if reference["index"] == -1 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			if reference["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			var intopcode int
			intopcode = 16
			current_byte_code = append(current_byte_code, intopcode, set_var["index"], reference["index"])
			byte_code = append(byte_code, current_byte_code)
		case "define.jump":
			if len(args) != 1 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			num, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				Debug_print(err)
				return current_gas
			}
			jump_table[int(num)] = i
			current_byte_code = append(current_byte_code, 17, int(num), i)
			byte_code = append(byte_code, current_byte_code)
		case "jump.def":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			num, err := strconv.ParseInt(args[0], 10, 64)
			condition_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if condition_var["index"] == -1 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			if condition_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if err != nil {
				Debug_print(err)
				return current_gas
			}
			current_byte_code = append(current_byte_code, 18, int(num), condition_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "endx":
			current_byte_code = append(current_byte_code, 19)
			byte_code = append(byte_code, current_byte_code)
		case "arr.init":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			num, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				Debug_print(err)
				return current_gas
			}
			arr_type := get_init_arr_type(string_consts[int(num)])
			if !type_evaluator(arr_type) {
				Debug_print("Invalid type for initialising an array")
				return current_gas
			}
			index := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: arr_type}, arr_constants, "var_name")
			if index["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			arr_default := VI_Object{var_name: args[0], object_type: arr_type}
			if index["result"] == 0 {
				index["index"] = len(arr_constants)
				arr_constants = append(arr_constants, arr_default)
			}
			res := plain_in_arr_VI_Object(arr_default, symbol_table, "var_name")
			if res["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if res["result"] == 0 {
				symbol_table = append(symbol_table, VI_Object{var_name: args[0], object_type: arr_type})
				res["index"] = len(symbol_table) - 1
			}
			current_byte_code = append(current_byte_code, 25, res["index"], index["index"])
			byte_code = append(byte_code, current_byte_code)
		case "arr.push":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			arr_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0]}, symbol_table, "var_name")
			if arr_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type array has not been initialised yet")
				return current_gas
			}
			if symbol_table[arr_var["index"]].object_type[0] != "arr" {
				Debug_print("Variable", args[0], "is not an array")
				return current_gas
			}
			push_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1]}, symbol_table, "var_name")
			if push_var["result"] == 0 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			if !obj_supports_type(symbol_table[arr_var["index"]].object_type, symbol_table[push_var["index"]].object_type) {
				Debug_print("Object type is not supported by array", args[0])
				return current_gas
			}
			current_byte_code = append(current_byte_code, 26, arr_var["index"], push_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "arr.pull", "arr.index.set":
			if len(args) != 3 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			arr_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0]}, symbol_table, "var_name")
			if arr_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type array has not been initialised yet")
				return current_gas
			}
			if symbol_table[arr_var["index"]].object_type[0] != "arr" {
				Debug_print("Variable", args[0], "is not an array")
				return current_gas
			}
			index_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if index_var["result"] == 0 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			if index_var["error"] == 1 {
				Debug_print("Index variable needs to be a number")
				return current_gas
			}
			pull_var := plain_in_arr_VI_Object(VI_Object{var_name: args[2]}, symbol_table, "var_name")
			if pull_var["result"] == 0 {
				Debug_print("Variable", args[2], "has not been initialised yet")
				return current_gas
			}
			if !obj_supports_type(symbol_table[arr_var["index"]].object_type, symbol_table[pull_var["index"]].object_type) {
				Debug_print("Object type does not match array", args[0], "object type")
				return current_gas
			}
			opcode_num := 0
			switch opcode {
			case "arr.pull":
				opcode_num = 27
			case "arr.index.set":
				opcode_num = 29
			}
			current_byte_code = append(current_byte_code, opcode_num, arr_var["index"], index_var["index"], pull_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "arr.remove", "arr.length":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			arr_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0]}, symbol_table, "var_name")
			if arr_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type array has not been initialised yet")
				return current_gas
			}
			if symbol_table[arr_var["index"]].object_type[0] != "arr" {
				Debug_print("Variable", args[0], "is not an array")
				return current_gas
			}
			index_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if index_var["result"] == 0 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			if index_var["error"] == 1 {
				Debug_print("Array index is required to be of type number")
				return current_gas
			}
			intopcode := 0
			switch opcode {
			case "arr.remove":
				intopcode = 28
			case "arr.length":
				intopcode = 30
			}
			current_byte_code = append(current_byte_code, intopcode, arr_var["index"], index_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "arr.refset":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			arr1_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0]}, symbol_table, "var_name")
			if arr1_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type array has not been initialised yet")
				return current_gas
			}
			if symbol_table[arr1_var["index"]].object_type[0] != "arr" {
				Debug_print("Variable", args[0], "is not an array")
				return current_gas
			}
			arr2_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if arr2_var["result"] == 0 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			if !(string_arr_compare(symbol_table[arr1_var["index"]].object_type, symbol_table[arr2_var["index"]].object_type)) {
				Debug_print("Both arrays must be of same type")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 31, arr1_var["index"], arr2_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "arr.includes":
			if len(args) != 3 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			arr_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0]}, symbol_table, "var_name")
			if arr_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type array has not been initialised yet")
				return current_gas
			}
			if symbol_table[arr_var["index"]].object_type[0] != "arr" {
				Debug_print("Variable", args[0], "is not an array")
				return current_gas
			}
			check_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1]}, symbol_table, "var_name")
			if check_var["result"] == 0 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			index_var := plain_in_arr_VI_Object(VI_Object{var_name: args[2], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if index_var["result"] == 0 {
				Debug_print("Variable", args[2], "has not been initialised yet")
				return current_gas
			}
			if index_var["error"] == 1 {
				Debug_print("Result index variable needs to be a number")
				return current_gas
			}
			if !obj_supports_type(symbol_table[arr_var["index"]].object_type, symbol_table[check_var["index"]].object_type) {
				Debug_print("Object type does not match array", args[0], "object type")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 32, arr_var["index"], check_var["index"], index_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "obj.equals":
			if len(args) != 3 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			obj1_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0]}, symbol_table, "var_name")
			if obj1_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type object has not been initialised yet")
				return current_gas
			}
			obj2_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1]}, symbol_table, "var_name")
			if obj2_var["result"] == 0 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			if !(string_arr_compare(symbol_table[obj1_var["index"]].object_type, symbol_table[obj2_var["index"]].object_type)) {
				Debug_print("Both objects must be of same type")
				return current_gas
			}
			res_var := plain_in_arr_VI_Object(VI_Object{var_name: args[2], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if res_var["result"] == 0 {
				Debug_print("Variable", args[2], "has not been initialised yet")
				return current_gas
			}
			if res_var["error"] == 1 {
				Debug_print("Result variable needs to be a number")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 33, obj1_var["index"], obj2_var["index"], res_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "str.equals":
			if len(args) != 3 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			var_1 := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("string")}, symbol_table, "var_name")
			var_2 := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("string")}, symbol_table, "var_name")
			var_res := plain_in_arr_VI_Object(VI_Object{var_name: args[2], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if var_1["error"] == 1 || var_2["error"] == 1 || var_res["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if var_1["index"] == -1 {
				Debug_print("Variable", args[0], "has not been initialised yet")
				return current_gas
			}
			if var_2["index"] == -1 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			if var_res["index"] == -1 {
				Debug_print("Variable", args[2], "has not been initialised yet")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 34, var_1["index"], var_2["index"], var_res["index"])
			byte_code = append(byte_code, current_byte_code)
		case "dict.init":
			if len(args) != 1 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			num, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				Debug_print(err)
				return current_gas
			}
			dict_type := get_init_dict_type(string_consts[int(num)])
			if !type_evaluator(dict_type) {
				Debug_print("Invalid type for initialising a dict")
				return current_gas
			}
			index := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: dict_type}, dict_constants, "var_name")
			if index["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			dict_default := VI_Object{var_name: args[0], object_type: dict_type}
			if index["result"] == 0 {
				index["index"] = len(dict_constants)
				dict_constants = append(dict_constants, dict_default)
			}
			res := plain_in_arr_VI_Object(dict_default, symbol_table, "var_name")
			if res["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if res["result"] == 0 {
				symbol_table = append(symbol_table, VI_Object{var_name: args[0], object_type: dict_type})
				res["index"] = len(symbol_table) - 1
			}
			current_byte_code = append(current_byte_code, 35, res["index"], index["index"])
			byte_code = append(byte_code, current_byte_code)
		case "dict.key.set", "dict.pull":
			if len(args) != 3 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			dict_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0]}, symbol_table, "var_name")
			if dict_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type dict has not been initialised yet")
				return current_gas
			}
			if symbol_table[dict_var["index"]].object_type[0] != "dict" {
				Debug_print("Variable", args[0], "is not a dict")
				return current_gas
			}
			key_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if key_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if key_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type string has not been initialised yet")
				return current_gas
			}
			value_var := plain_in_arr_VI_Object(VI_Object{var_name: args[2]}, symbol_table, "var_name")
			if value_var["result"] == 0 {
				Debug_print("Variable", args[2], "has not been initialised yet")
				return current_gas
			}
			if !obj_supports_type(symbol_table[dict_var["index"]].object_type, symbol_table[value_var["index"]].object_type) {
				Debug_print("Object type is not supported by dict", args[0])
				return current_gas
			}
			var int_opcode int
			switch opcode {
			case "dict.key.set":
				int_opcode = 36
			case "dict.pull":
				int_opcode = 37
			}
			current_byte_code = append(current_byte_code, int_opcode, dict_var["index"], key_var["index"], value_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "dict.delete":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			dict_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0]}, symbol_table, "var_name")
			if dict_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type dict has not been initialised yet")
				return current_gas
			}
			if symbol_table[dict_var["index"]].object_type[0] != "dict" {
				Debug_print("Variable", args[0], "is not a dict")
				return current_gas
			}
			key_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if key_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if key_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type string has not been initialised yet")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 38, dict_var["index"], key_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "dict.keys":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			dict_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0]}, symbol_table, "var_name")
			if dict_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type dict has not been initialised yet")
				return current_gas
			}
			if symbol_table[dict_var["index"]].object_type[0] != "dict" {
				Debug_print("Variable", args[0], "is not a dict")
				return current_gas
			}
			arr_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_init_arr_type("string")}, symbol_table, "var_name")
			if arr_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if arr_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type array[string] has not been initialised yet")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 39, dict_var["index"], arr_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "dict.refset":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			dict1_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0]}, symbol_table, "var_name")
			if dict1_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type dict has not been initialised yet")
				return current_gas
			}
			if symbol_table[dict1_var["index"]].object_type[0] != "dict" {
				Debug_print("Variable", args[0], "is not a dict")
				return current_gas
			}
			dict2_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1]}, symbol_table, "var_name")
			if dict2_var["result"] == 0 {
				Debug_print("Variable", args[1], "of type dict has not been initialised yet")
				return current_gas
			}
			if symbol_table[dict2_var["index"]].object_type[0] != "dict" {
				Debug_print("Variable", args[1], "is not a dict")
				return current_gas
			}
			if !string_arr_compare(symbol_table[dict1_var["index"]].object_type, symbol_table[dict2_var["index"]].object_type) {
				Debug_print("Dictionaries are of different types")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 40, dict1_var["index"], dict2_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "dict.key.includes":
			if len(args) != 3 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			dict_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0]}, symbol_table, "var_name")
			if dict_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type dict has not been initialised yet")
				return current_gas
			}
			if symbol_table[dict_var["index"]].object_type[0] != "dict" {
				Debug_print("Variable", args[0], "is not a dict")
				return current_gas
			}
			key_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if key_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if key_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type string has not been initialised yet")
				return current_gas
			}
			value_var := plain_in_arr_VI_Object(VI_Object{var_name: args[2], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if value_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if value_var["result"] == 0 {
				Debug_print("Variable", args[2], "has not been initialised yet")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 41, dict_var["index"], key_var["index"], value_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "str.includes":
			if len(args) != 3 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			str1_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if str1_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if str1_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type string has not been initialised yet")
				return current_gas
			}
			str2_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if str2_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if str2_var["result"] == 0 {
				Debug_print("Variable", args[1], "of type string has not been initialised yet")
				return current_gas
			}
			value_var := plain_in_arr_VI_Object(VI_Object{var_name: args[2], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if value_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if value_var["result"] == 0 {
				Debug_print("Variable", args[2], "has not been initialised yet")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 42, str1_var["index"], str2_var["index"], value_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "str.replace":
			if len(args) != 4 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			str1_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if str1_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if str1_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type string has not been initialised yet")
				return current_gas
			}
			str2_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if str2_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if str2_var["result"] == 0 {
				Debug_print("Variable", args[1], "of type string has not been initialised yet")
				return current_gas
			}
			str3_var := plain_in_arr_VI_Object(VI_Object{var_name: args[2], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if str3_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if str3_var["result"] == 0 {
				Debug_print("Variable", args[2], "of type string has not been initialised yet")
				return current_gas
			}
			value_var := plain_in_arr_VI_Object(VI_Object{var_name: args[3], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if value_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if value_var["result"] == 0 {
				Debug_print("Variable", args[3], "has not been initialised yet")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 43, str1_var["index"], str2_var["index"], str3_var["index"], value_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "str.index":
			if len(args) != 3 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			str1_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if str1_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if str1_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type string has not been initialised yet")
				return current_gas
			}
			str2_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if str2_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if str2_var["result"] == 0 {
				Debug_print("Variable", args[1], "of type string has not been initialised yet")
				return current_gas
			}
			value_var := plain_in_arr_VI_Object(VI_Object{var_name: args[2], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if value_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if value_var["result"] == 0 {
				Debug_print("Variable", args[2], "has not been initialised yet")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 44, str1_var["index"], str2_var["index"], value_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "str.split":
			if len(args) != 3 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			str1_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if str1_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if str1_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type string has not been initialised yet")
				return current_gas
			}
			str2_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if str2_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if str2_var["result"] == 0 {
				Debug_print("Variable", args[1], "of type string has not been initialised yet")
				return current_gas
			}
			arr1_var := plain_in_arr_VI_Object(VI_Object{var_name: args[2]}, symbol_table, "var_name")
			if arr1_var["result"] == 0 {
				Debug_print("Variable", args[2], "of type array has not been initialised yet")
				return current_gas
			}
			if symbol_table[arr1_var["index"]].object_type[0] != "arr" {
				Debug_print("Variable", args[2], "is not an array")
				return current_gas
			}
			if !obj_supports_type(symbol_table[arr1_var["index"]].object_type, symbol_table[str2_var["index"]].object_type) {
				Debug_print("Object type is not supported by array", args[2])
				return current_gas
			}
			current_byte_code = append(current_byte_code, 45, str1_var["index"], str2_var["index"], arr1_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "str.pull":
			if len(args) != 3 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			str1_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if str1_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if str1_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type string has not been initialised yet")
				return current_gas
			}
			index_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if index_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if index_var["result"] == 0 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			value_var := plain_in_arr_VI_Object(VI_Object{var_name: args[2], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if value_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if value_var["result"] == 0 {
				Debug_print("Variable", args[2], "has not been initialised yet")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 46, str1_var["index"], index_var["index"], value_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "str.slice_n":
			if len(args) != 3 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			str1_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if str1_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if str1_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type string has not been initialised yet")
				return current_gas
			}
			index_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if index_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if index_var["result"] == 0 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			arr1_var := plain_in_arr_VI_Object(VI_Object{var_name: args[2]}, symbol_table, "var_name")
			if arr1_var["result"] == 0 {
				Debug_print("Variable", args[2], "of type array has not been initialised yet")
				return current_gas
			}
			if symbol_table[arr1_var["index"]].object_type[0] != "arr" {
				Debug_print("Variable", args[2], "is not an array")
				return current_gas
			}
			if !obj_supports_type(symbol_table[arr1_var["index"]].object_type, symbol_table[str1_var["index"]].object_type) {
				Debug_print("Object type is not supported by array", args[2])
				return current_gas
			}
			current_byte_code = append(current_byte_code, 47, str1_var["index"], index_var["index"], arr1_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "scope.new":
			current_byte_code = append(current_byte_code, 48)
			byte_code = append(byte_code, current_byte_code)
		case "scope.exit":
			current_byte_code = append(current_byte_code, 49)
			byte_code = append(byte_code, current_byte_code)
		case "pointer.init":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			pointing_variable := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if pointing_variable["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if pointing_variable["result"] == 0 {
				Debug_print("Variable", args[0], "has not been initialised yet")
				return current_gas
			}
			pointing_to_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1]}, symbol_table, "var_name")
			if pointing_to_var["result"] == 0 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 50, pointing_variable["index"], pointing_to_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "pointer.dereference":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			pointer := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if pointer["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if pointer["result"] == 0 {
				Debug_print("Variable", args[0], "has not been initialised yet")
				return current_gas
			}
			dereferencing_variable := plain_in_arr_VI_Object(VI_Object{var_name: args[1]}, symbol_table, "var_name")
			if dereferencing_variable["result"] == 0 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 51, pointer["index"], dereferencing_variable["index"])
			byte_code = append(byte_code, current_byte_code)
		case "num_to_str":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			num_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if num_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if num_var["result"] == 0 {
				Debug_print("Variable", args[0], "has not been initialised yet")
				return current_gas
			}
			str2_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if str2_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if str2_var["result"] == 0 {
				Debug_print("Variable", args[1], "of type string has not been initialised yet")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 52, num_var["index"], str2_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "str_to_num":
			if len(args) != 3 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			str1_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if str1_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if str1_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type string has not been initialised yet")
				return current_gas
			}
			num_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if num_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if num_var["result"] == 0 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			error_var := plain_in_arr_VI_Object(VI_Object{var_name: args[2], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if error_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if error_var["result"] == 0 {
				Debug_print("Variable", args[2], "of type string has not been initialised yet")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 53, str1_var["index"], num_var["index"], error_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "str.length":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			str1_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if str1_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if str1_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type string has not been initialised yet")
				return current_gas
			}
			num_var := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if num_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if num_var["result"] == 0 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 54, str1_var["index"], num_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "debug.print":
			if len(args) !=1 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			str1_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("string")}, symbol_table, "var_name")
			if str1_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if str1_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type string has not been initialised yet")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 55, str1_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "register_struct":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			floatnum, err := strconv.ParseFloat(args[1], 64)
			num:=uint(int(floatnum))
			if err != nil {
				Debug_print(err)
				return current_gas
			}
			if len(string_consts)<int(num) {
				Debug_print("Data index not found", len(string_consts), num)
				return current_gas
			}
			if str_index_in_arr(args[0], registered_structs["structs"])!=-1 {
				Debug_print("Cannot reassign a struct twice")
				return current_gas
			}
			registered_structs[args[0]]=make([]string, 0)
			if string_consts[num]!="" {
				for _,definition:=range strings.Split(string_consts[num], ";") {
					if !valid_var_name(strings.Split(definition, "->")[0]) {
						Debug_print("Invalid struct field name")
						return current_gas
					}
					if !type_evaluator(strings.Split(strings.Split(definition, "->")[1], ",")) {
						Debug_print("Invalid struct type",strings.Split(strings.Split(definition, "->")[1], ","))
						return current_gas
					}
					registered_structs[args[0]] = append(registered_structs[args[0]], definition)
				}
			}
			registered_structs["structs"] = append(registered_structs["structs"], args[0])
		case "struct.init":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			floatnum, err := strconv.ParseFloat(args[1], 64)
			num:=uint(int(floatnum))
			if err != nil {
				Debug_print(err)
				return current_gas
			}
			if len(string_consts)<int(num) {
				Debug_print("Data index not found", len(string_consts), num)
				return current_gas
			}
			args[1]=string_consts[num]
			if str_index_in_arr(args[1], registered_structs["structs"])==-1 {
				Debug_print("Struct does not exist", args[1], registered_structs["structs"])
				return current_gas
			}
			index := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: []string{args[1]}}, arr_constants, "var_name")
			if index["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			struct_default := spawn_struct_default(args[0], args[1])
			if index["result"] == 0 {
				index["index"] = len(arr_constants)
				arr_constants = append(arr_constants, struct_default)
			}
			res := plain_in_arr_VI_Object(struct_default, symbol_table, "var_name")
			if res["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if res["result"] == 0 {
				symbol_table = append(symbol_table, struct_default)
				res["index"] = len(symbol_table) - 1
			}
			current_byte_code = append(current_byte_code, 57, res["index"], int(num))
			byte_code = append(byte_code, current_byte_code)
		case "struct.set":
			if len(args) != 3 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			floatnum, err := strconv.ParseFloat(args[1], 64)
			num:=uint(int(floatnum))
			if err != nil {
				Debug_print(err)
				return current_gas
			}
			if len(string_consts)<int(num) {
				Debug_print("Data index not found", len(string_consts), num)
				return current_gas
			}
			args[1]=string_consts[num]
			struct_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0]}, symbol_table, "var_name")
			if struct_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type struct has not been initialised yet")
				return current_gas
			}
			if len(symbol_table[struct_var["index"]].object_type)!=1 || str_index_in_arr(symbol_table[struct_var["index"]].object_type[0], registered_structs["structs"])==-1 || str_index_in_arr(symbol_table[struct_var["index"]].object_type[0], []string{"num","string"})!=-1 {
				Debug_print("Invalid struct variable", )
				return current_gas
			}
			struct_field_type:=make([]string, 0)
			for _,value:=range registered_structs[symbol_table[struct_var["index"]].object_type[0]] {
				if strings.Split(value, "->")[0]==args[1] {
					struct_field_type=strings.Split(strings.Split(value, "->")[1], ",")
				}
			}
			if len(struct_field_type)==0 {
				Debug_print("Field does not exist")
				return current_gas
			}
			push_var := plain_in_arr_VI_Object(VI_Object{var_name: args[2]}, symbol_table, "var_name")
			if push_var["result"] == 0 {
				Debug_print("Variable", args[2], "has not been initialised yet")
				return current_gas
			}
			if !string_arr_compare(struct_field_type, symbol_table[push_var["index"]].object_type) {
				Debug_print("Object type is not supported by struct", args[0])
				return current_gas
			}
			current_byte_code = append(current_byte_code, 58, struct_var["index"], int(num), push_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "struct.pull":
			if len(args) != 3 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			floatnum, err := strconv.ParseFloat(args[1], 64)
			num:=uint(int(floatnum))
			if err != nil {
				Debug_print(err)
				return current_gas
			}
			if len(string_consts)<int(num) {
				Debug_print("Data index not found", len(string_consts), num)
				return current_gas
			}
			args[1]=string_consts[num]
			struct_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0]}, symbol_table, "var_name")
			if struct_var["result"] == 0 {
				Debug_print("Variable", args[0], "of type struct has not been initialised yet")
				return current_gas
			}
			if len(symbol_table[struct_var["index"]].object_type)!=1 || str_index_in_arr(symbol_table[struct_var["index"]].object_type[0], registered_structs["structs"])==-1 || str_index_in_arr(symbol_table[struct_var["index"]].object_type[0], []string{"num","string"})!=-1 {
				Debug_print("Invalid struct variable", )
				return current_gas
			}
			struct_field_type:=make([]string, 0)
			for _,value:=range registered_structs[symbol_table[struct_var["index"]].object_type[0]] {
				if strings.Split(value, "->")[0]==args[1] {
					struct_field_type=strings.Split(strings.Split(value, "->")[1], ",")
				}
			}
			if len(struct_field_type)==0 {
				Debug_print("Field does not exist")
				return current_gas
			}
			push_var := plain_in_arr_VI_Object(VI_Object{var_name: args[2]}, symbol_table, "var_name")
			if push_var["result"] == 0 {
				Debug_print("Variable", args[2], "has not been initialised yet")
				return current_gas
			}
			if !string_arr_compare(struct_field_type, symbol_table[push_var["index"]].object_type) {
				Debug_print("Object type is not supported by struct", args[0])
				return current_gas
			}
			current_byte_code = append(current_byte_code, 59, struct_var["index"], int(num), push_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "jump.def.always":
			if len(args) != 1 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			num, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				Debug_print(err)
				return current_gas
			}
			current_byte_code = append(current_byte_code, 60, int(num))
			byte_code = append(byte_code, current_byte_code)
		case "jump_n_lines":
			if len(args) != 2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			set_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("num")}, symbol_table, "var_name")
			reference := plain_in_arr_VI_Object(VI_Object{var_name: args[1], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if set_var["index"] == -1 {
				Debug_print("Variable", args[0], "has not been initialised yet")
				return current_gas
			}
			if set_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			if reference["index"] == -1 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			if reference["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 61, set_var["index"], reference["index"])
			byte_code = append(byte_code, current_byte_code)
		case "padding":
			if len(args) != 0 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 62)
			byte_code = append(byte_code, current_byte_code)
		case "obj_copy":
			if len(args)!=2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			reference := plain_in_arr_VI_Object(VI_Object{var_name: args[1]}, symbol_table, "var_name")
			set_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0]}, symbol_table, "var_name")
			if set_var["index"] == -1 {
				Debug_print("Variable", args[0], "has not been initialised yet")
				return current_gas
			}
			if reference["index"] == -1 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			if !string_arr_compare(symbol_table[set_var["index"]].object_type, symbol_table[reference["index"]].object_type) {
				Debug_print("Object b must have same type as object a")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 65, set_var["index"], reference["index"])
			byte_code = append(byte_code, current_byte_code)
		case "obj_copy_update_scope":
			if len(args)!=2 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			reference := plain_in_arr_VI_Object(VI_Object{var_name: args[1]}, symbol_table, "var_name")
			set_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0]}, symbol_table, "var_name")
			if set_var["index"] == -1 {
				Debug_print("Variable", args[0], "has not been initialised yet")
				return current_gas
			}
			if reference["index"] == -1 {
				Debug_print("Variable", args[1], "has not been initialised yet")
				return current_gas
			}
			if !string_arr_compare(symbol_table[set_var["index"]].object_type, symbol_table[reference["index"]].object_type) {
				Debug_print("Object b must have same type as object a", symbol_table[set_var["index"]], symbol_table[reference["index"]])
				return current_gas
			}
			current_byte_code = append(current_byte_code, 66, set_var["index"], reference["index"])
			byte_code = append(byte_code, current_byte_code)
		case "jump.always.var":
			if len(args) != 1 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			set_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if set_var["index"] == -1 {
				Debug_print("Variable", args[0], "has not been initialised yet")
				return current_gas
			}
			if set_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 67, set_var["index"])
			byte_code = append(byte_code, current_byte_code)
		case "current_line_number":
			if len(args) != 1 {
				Debug_print("Invalid number of arguments")
				return current_gas
			}
			set_var := plain_in_arr_VI_Object(VI_Object{var_name: args[0], object_type: get_plain_type("num")}, symbol_table, "var_name")
			if set_var["index"] == -1 {
				Debug_print("Variable", args[0], "has not been initialised yet")
				return current_gas
			}
			if set_var["error"] == 1 {
				Debug_print("Data types did not match")
				return current_gas
			}
			current_byte_code = append(current_byte_code, 68, set_var["index"])
			byte_code = append(byte_code, current_byte_code)
		}
	}
	Debug_print(byte_code)
	continue_exec := true
	for i := 0; i < len(byte_code); i++ {
		current_gas += 1
		if current_gas < gas_limit || !continue_exec {
			break
		}
		current_byte_code := byte_code[i]
		switch opcode := current_byte_code[0]; opcode {
		case 0:
			symbol_table[current_byte_code[1]].num_value = num_constants[current_byte_code[2]]
			symbol_table[current_byte_code[1]].scope = scope_count
		case 1:
			symbol_table[current_byte_code[1]].num_value = symbol_table[current_byte_code[2]].num_value
		case 2:
			if symbol_table[current_byte_code[2]].num_value != 0 {
				i = int(symbol_table[current_byte_code[1]].num_value) - 1
			}
		case 3:
			if symbol_table[current_byte_code[1]].num_value == symbol_table[current_byte_code[2]].num_value {
				symbol_table[current_byte_code[3]].num_value = 1
			} else {
				symbol_table[current_byte_code[3]].num_value = 0
			}
		case 4:
			if symbol_table[current_byte_code[1]].num_value > symbol_table[current_byte_code[2]].num_value {
				symbol_table[current_byte_code[3]].num_value = 1
			} else {
				symbol_table[current_byte_code[3]].num_value = 0
			}
		case 5:
			symbol_table[current_byte_code[3]].num_value = symbol_table[current_byte_code[1]].num_value + symbol_table[current_byte_code[2]].num_value
		case 6:
			symbol_table[current_byte_code[3]].num_value = symbol_table[current_byte_code[1]].num_value - symbol_table[current_byte_code[2]].num_value
		case 7:
			symbol_table[current_byte_code[3]].num_value = symbol_table[current_byte_code[1]].num_value * symbol_table[current_byte_code[2]].num_value
		case 8:
			symbol_table[current_byte_code[3]].num_value = symbol_table[current_byte_code[1]].num_value / symbol_table[current_byte_code[2]].num_value
		case 9:
			symbol_table[current_byte_code[3]].num_value = math.Pow(symbol_table[current_byte_code[1]].num_value, symbol_table[current_byte_code[2]].num_value)
		case 10:
			symbol_table[current_byte_code[3]].num_value = math.Floor(symbol_table[current_byte_code[1]].num_value / symbol_table[current_byte_code[2]].num_value)
		case 11:
			symbol_table[current_byte_code[3]].num_value = math.Mod(symbol_table[current_byte_code[1]].num_value, symbol_table[current_byte_code[2]].num_value)
		case 12:
			symbol_table[current_byte_code[1]].str_value = string_consts[current_byte_code[2]]
			symbol_table[current_byte_code[1]].scope = scope_count
		case 14:
			symbol_table[current_byte_code[3]].str_value = symbol_table[current_byte_code[1]].str_value + symbol_table[current_byte_code[2]].str_value
		case 15:
			symbol_table[current_byte_code[3]].str_value = strings.Repeat(symbol_table[current_byte_code[1]].str_value, int(symbol_table[current_byte_code[2]].num_value))
		case 16:
			symbol_table[current_byte_code[1]].str_value = symbol_table[current_byte_code[2]].str_value
		case 17:
			jump_table[current_byte_code[1]] = current_byte_code[2]
		case 18:
			if symbol_table[current_byte_code[2]].num_value != 0 {
				i = jump_table[current_byte_code[1]]
			}
		case 19:
			continue_exec = false
		case 20:
			symbol_table[current_byte_code[3]].num_value = bool_to_num(num_to_bool(symbol_table[current_byte_code[1]].num_value) && num_to_bool(symbol_table[current_byte_code[2]].num_value))
		case 21:
			symbol_table[current_byte_code[3]].num_value = bool_to_num(num_to_bool(symbol_table[current_byte_code[1]].num_value) || num_to_bool(symbol_table[current_byte_code[2]].num_value))
		case 22:
			symbol_table[current_byte_code[2]].num_value = bool_to_num(!num_to_bool(symbol_table[current_byte_code[1]].num_value))
		case 23:
			symbol_table[current_byte_code[3]].num_value = float64(int(symbol_table[current_byte_code[1]].num_value) ^ int(symbol_table[current_byte_code[2]].num_value))
		case 24:
			symbol_table[current_byte_code[3]].num_value = round_float64(symbol_table[current_byte_code[1]].num_value, uint(symbol_table[current_byte_code[2]].num_value))
		case 25:
			symbol_table[current_byte_code[1]] = arr_constants[current_byte_code[2]]
			symbol_table[current_byte_code[1]].scope = scope_count
		case 26:
			symbol_table[current_byte_code[1]].children = append(symbol_table[current_byte_code[1]].children, symbol_table[current_byte_code[2]])
		case 27:
			arr_index := int(symbol_table[current_byte_code[2]].num_value)
			if arr_index >= len(symbol_table[current_byte_code[1]].children) {
				Debug_print("Index out of range")
				return current_gas
			}
			temp_var := symbol_table[current_byte_code[1]].children[arr_index]
			temp_var.var_name = symbol_table[current_byte_code[3]].var_name
			symbol_table[current_byte_code[3]] = temp_var
		case 28:
			symbol_table[current_byte_code[1]].children = remove_VI_Object_from_index(symbol_table[current_byte_code[1]].children, int(symbol_table[current_byte_code[2]].num_value))
		case 29:
			arr_index := int(symbol_table[current_byte_code[2]].num_value)
			if arr_index >= len(symbol_table[current_byte_code[1]].children) {
				Debug_print("Index out of range")
				return current_gas
			}
			symbol_table[current_byte_code[1]].children[arr_index] = symbol_table[current_byte_code[3]]
		case 30:
			symbol_table[current_byte_code[2]].num_value = float64(len(symbol_table[current_byte_code[1]].children))
		case 31:
			symbol_table[current_byte_code[1]].children = symbol_table[current_byte_code[2]].children
		case 32:
			arr_var := symbol_table[current_byte_code[1]].children
			check_var := symbol_table[current_byte_code[2]]
			res_index := -1
			for i := 0; i < len(symbol_table[current_byte_code[1]].children); i++ {
				if recursive_VI_Object_match(arr_var[i], check_var) {
					res_index = i
					break
				}
			}
			symbol_table[current_byte_code[3]].num_value = float64(res_index)
		case 33:
			symbol_table[current_byte_code[3]].num_value = bool_to_num(recursive_VI_Object_match(symbol_table[current_byte_code[1]], symbol_table[current_byte_code[2]]))
		case 34:
			symbol_table[current_byte_code[3]].num_value = bool_to_num(symbol_table[current_byte_code[1]].str_value == symbol_table[current_byte_code[2]].str_value)
		case 35:
			symbol_table[current_byte_code[1]] = dict_constants[current_byte_code[2]]
			symbol_table[current_byte_code[1]].scope = scope_count
		case 36:
			index := str_index_in_arr(symbol_table[current_byte_code[2]].str_value, symbol_table[current_byte_code[1]].dict_keys)
			if index == -1 {
				symbol_table[current_byte_code[1]].dict_keys = append(symbol_table[current_byte_code[1]].dict_keys, symbol_table[current_byte_code[2]].str_value)
				symbol_table[current_byte_code[1]].children = append(symbol_table[current_byte_code[1]].children, symbol_table[current_byte_code[3]])
			} else {
				symbol_table[current_byte_code[1]].children[index] = symbol_table[current_byte_code[3]]
			}
		case 37:
			temp_pull_index := str_index_in_arr(symbol_table[current_byte_code[2]].str_value, symbol_table[current_byte_code[1]].dict_keys)
			if temp_pull_index == -1 {
				Debug_print("String not found in array")
				return current_gas
			}
			temp_pull := symbol_table[current_byte_code[1]].children[temp_pull_index]
			temp_pull.var_name = symbol_table[current_byte_code[3]].var_name
			symbol_table[current_byte_code[3]] = temp_pull
		case 38:
			temp_pull_index := str_index_in_arr(symbol_table[current_byte_code[2]].str_value, symbol_table[current_byte_code[1]].dict_keys)
			symbol_table[current_byte_code[1]].dict_keys = remove_string_from_index(symbol_table[current_byte_code[1]].dict_keys, temp_pull_index)
			symbol_table[current_byte_code[1]].children = remove_VI_Object_from_index(symbol_table[current_byte_code[1]].children, temp_pull_index)
		case 39:
			keys_arr_VI_Object := make([]VI_Object, 0)
			for i := 0; i < len(symbol_table[current_byte_code[1]].dict_keys); i++ {
				keys_arr_VI_Object = append(keys_arr_VI_Object, VI_Object{object_type: get_plain_type("string"), str_value: symbol_table[current_byte_code[1]].dict_keys[i]})
			}
			symbol_table[current_byte_code[2]].children = keys_arr_VI_Object
		case 40:
			symbol_table[current_byte_code[1]].children = symbol_table[current_byte_code[2]].children
			symbol_table[current_byte_code[1]].dict_keys = symbol_table[current_byte_code[2]].dict_keys
		case 41:
			symbol_table[current_byte_code[3]].num_value = bool_to_num(str_index_in_arr(symbol_table[current_byte_code[2]].str_value, symbol_table[current_byte_code[1]].dict_keys) != -1)
		case 42:
			symbol_table[current_byte_code[3]].num_value = bool_to_num(strings.Contains(symbol_table[current_byte_code[1]].str_value, symbol_table[current_byte_code[2]].str_value))
		case 43:
			symbol_table[current_byte_code[4]].str_value = strings.ReplaceAll(symbol_table[current_byte_code[1]].str_value, symbol_table[current_byte_code[2]].str_value, symbol_table[current_byte_code[3]].str_value)
		case 44:
			symbol_table[current_byte_code[3]].num_value = float64(strings.Index(symbol_table[current_byte_code[1]].str_value, symbol_table[current_byte_code[2]].str_value))
		case 45:
			splitted_string := strings.Split(symbol_table[current_byte_code[1]].str_value, symbol_table[current_byte_code[2]].str_value)
			splitted_strings_objects := make([]VI_Object, 0)
			for _, string := range splitted_string {
				splitted_strings_objects = append(splitted_strings_objects, VI_Object{object_type: get_plain_type("string"), str_value: string})
			}
			symbol_table[current_byte_code[3]].children = splitted_strings_objects
		case 46:
			symbol_table[current_byte_code[3]].str_value = string(symbol_table[current_byte_code[1]].str_value[int(symbol_table[current_byte_code[2]].num_value)])
		case 47:
			symbol_table[current_byte_code[3]].children = []VI_Object{
				VI_Object{object_type: get_plain_type("string"), str_value: symbol_table[current_byte_code[1]].str_value[:int(symbol_table[current_byte_code[2]].num_value)]},
				VI_Object{object_type: get_plain_type("string"), str_value: symbol_table[current_byte_code[1]].str_value[int(symbol_table[current_byte_code[2]].num_value):]}}
		case 48:
			global_table[scope_count] = symbol_table
			scope_count += 1
			previous_scope := make([]VI_Object, 0)
			for _, object := range symbol_table {
				previous_scope = append(previous_scope, VI_Object{var_name: object.var_name, num_value: object.num_value, str_value: object.str_value, object_type: object.object_type, children: object.children, scope: object.scope, dict_keys: object.dict_keys})
			}
			global_table = append(global_table, previous_scope)
			symbol_table = previous_scope
		case 49:
			if scope_count == 0 {
				continue
			}
			for i_, variable := range symbol_table {
				if variable.scope < scope_count {
					if global_table[scope_count-1][i_].var_name=="return_to" {
					}
					global_table[scope_count-1][i_] = variable
				}
			}
			global_table = global_table[:len(global_table)-1]
			scope_count -= 1
			symbol_table = global_table[scope_count]
		case 50:
			symbol_table[current_byte_code[1]].num_value = float64(current_byte_code[2])
		case 51:
			memory_location := symbol_table[current_byte_code[1]].num_value
			if len(symbol_table) <= int(memory_location) {
				Debug_print("Illegal memory access")
				continue_exec = false
				continue
			}
			dereferenced_variable := symbol_table[int(memory_location)]
			dereferencing_to_variable := symbol_table[current_byte_code[2]]
			if !string_arr_compare(dereferenced_variable.object_type, dereferencing_to_variable.object_type) {
				Debug_print("Invalid deferenced variable type")
				continue_exec = false
				continue
			}
			dereferenced_variable.var_name = dereferencing_to_variable.var_name
			symbol_table[current_byte_code[2]] = dereferenced_variable
		case 52:
			symbol_table[current_byte_code[2]].str_value = strconv.FormatFloat(symbol_table[current_byte_code[1]].num_value, 'f', 8, 64)
		case 53:
			float_64_num, err := strconv.ParseFloat(symbol_table[current_byte_code[1]].str_value, 64)
			if err != nil {
				symbol_table[current_byte_code[2]].num_value = 0
				symbol_table[current_byte_code[3]].num_value = 1
			} else {
				symbol_table[current_byte_code[2]].num_value = float_64_num
				symbol_table[current_byte_code[3]].num_value = 0
			}
		case 54:
			symbol_table[current_byte_code[2]].num_value = float64(len(symbol_table[current_byte_code[1]].str_value))
		case 55:
			Debug_printf(symbol_table[current_byte_code[1]].str_value)
		case 57:
			current_var_name:=symbol_table[current_byte_code[1]].var_name
			symbol_table[current_byte_code[1]]=spawn_struct_default("", string_consts[current_byte_code[2]])
			symbol_table[current_byte_code[1]].var_name=current_var_name
		case 58:
			symbol_table[current_byte_code[1]]._struct[current_byte_code[2]]=symbol_table[current_byte_code[3]]
		case 59:
			current_var_name:=symbol_table[current_byte_code[3]].var_name
			symbol_table[current_byte_code[3]]=symbol_table[current_byte_code[1]]._struct[current_byte_code[2]]
			symbol_table[current_byte_code[3]].var_name=current_var_name
		case 60:
			i = jump_table[current_byte_code[1]]
		case 61:
			if symbol_table[current_byte_code[2]].num_value != 0 {
				i+=int(symbol_table[current_byte_code[1]].num_value)
			}
		case 64:
			if symbol_table[current_byte_code[1]].num_value < symbol_table[current_byte_code[2]].num_value {
				symbol_table[current_byte_code[3]].num_value = 1
			} else {
				symbol_table[current_byte_code[3]].num_value = 0
			}
		case 65:
			symbol_table[current_byte_code[1]]=copy_VI_Object(symbol_table[current_byte_code[2]])
		case 66:
			symbol_table[current_byte_code[1]]=copy_VI_Object(symbol_table[current_byte_code[2]])
			symbol_table[current_byte_code[1]].scope = scope_count
		case 67:
			i = int(symbol_table[current_byte_code[1]].num_value)-1
		case 68:
			symbol_table[current_byte_code[1]].num_value = float64(i)
		}
	}
	global_table[scope_count] = symbol_table
	for current_global_table_index := range global_table {
		Debug_print(global_table[current_global_table_index])
	}
	return current_gas
}
