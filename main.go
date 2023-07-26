package main

import (
	"fmt"
	"os"
	"strings"
	"strconv"
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
	return res
}

func get_plain_type(obj_type string) []string {
	res:=make([]string,0)
	res = append(res, obj_type)
	return res
}

func get_init_arr_type(obj_type string) []string {
	res:=make([]string,0)
	res = append(res, "arr")
	res = append(res, obj_type)
	return res
}

func add_arr_depth(current_obj_type []string) []string {
	res:=make([]string,0)
	res = append([]string{"arr"}, current_obj_type...)
	return res
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

func main() {
	dat, err := os.ReadFile("alu.vi")

	if (err!=nil) {
		fmt.Println(err)
		return
	}
	code:=string(dat)
	codex := strings.Split(code, "\n")

	var num_constants []float64;

	byte_code:=make([][]int,0)
	symbol_table:=make([]VI_Object,0)
	gas_limit:=0
	current_gas:=1

	for i := 0; i < len(codex); i++ {
        args:=strings.Split(strings.Split(codex[i], " ")[1], ",")
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
			res:=plain_in_arr_VI_Object(VI_Object{var_name: args[0]},symbol_table,"var_name")
			if res["result"]==0 {
				symbol_table = append(symbol_table, VI_Object{var_name: args[0],object_type: get_plain_type("num")})
				res["index"]=len(symbol_table)-1
			}
			current_byte_code = append(current_byte_code, 0,res["index"],index["index"])
			byte_code = append(byte_code, current_byte_code)
		case "refset","jump":
			set_var:=plain_in_arr_VI_Object(VI_Object{var_name: args[0]},symbol_table,"var_name")
			reference:=plain_in_arr_VI_Object(VI_Object{var_name: args[1]},symbol_table,"var_name")
			if set_var["index"]==-1 {
				fmt.Println("Variable",args[0],"has not been initialised yet")
				return
			}
			if reference["index"]==-1 {
				fmt.Println("Variable",args[1],"has not been initialised yet")
				return
			}
			var intopcode int;
			if (opcode=="refset") {intopcode=1}
			if (opcode=="jump") {intopcode=2}
			current_byte_code = append(current_byte_code, intopcode, set_var["index"], reference["index"])
			byte_code = append(byte_code, current_byte_code)
		case "equals","greater","add","sub","mult","div","floor","mod","power","xor.num","round":
			var_1:=plain_in_arr_VI_Object(VI_Object{var_name: args[0]},symbol_table,"var_name")
			var_2:=plain_in_arr_VI_Object(VI_Object{var_name: args[1]},symbol_table,"var_name")
			var_res:=plain_in_arr_VI_Object(VI_Object{var_name: args[2]},symbol_table,"var_name")
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
			case "xor.num":intopcode=24
			case "round":intopcode=26
			}
			current_byte_code = append(current_byte_code, intopcode, var_1["index"], var_2["index"], var_res["index"])
			byte_code = append(byte_code, current_byte_code)
		}
	}
	fmt.Println(byte_code)
	for i := 0; i < len(byte_code); i++ {
		current_gas+=1
		if current_gas<gas_limit {
			return
		}
		current_byte_code:=byte_code[i]
		switch opcode:=current_byte_code[0]; opcode {
		case 0:
			symbol_table[current_byte_code[1]].num_value=num_constants[current_byte_code[2]]
		case 1:
			symbol_table[current_byte_code[1]].num_value=symbol_table[current_byte_code[2]].num_value
		case 2:
			if symbol_table[current_byte_code[2]].num_value!=0 {
				i=int(symbol_table[current_byte_code[1]].num_value)-2
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
		}
	}
	fmt.Println(symbol_table,num_constants)
}