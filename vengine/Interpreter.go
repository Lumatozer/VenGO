package Vengine

import (
	"errors"
	"fmt"
	"github.com/lumatozer/VenGO/structs"
)

type Stack struct {
	Locations       []int
	Objects         []*Object
}

func Bool2Int(a bool) int {
	if a {
		return 1
	}
	return 0
}

func Interpreter(function *Function, stack Stack, thread_Mutex *structs.Mutex_Interface, Database_Interface structs.Database_Interface) structs.Execution_Result {
	execution_Result:=structs.Execution_Result{}
	scope := make([]*Object, len(function.Base_Program.Rendered_Scope))
	constructed_Objects:=make(map[int]Object)
	for i:=0; i<len(scope); i++ {
		scope[i] = function.Base_Program.Rendered_Scope[i]
		stack_Index:=int_index_in_int_arr(i, stack.Locations)
		if stack_Index!=-1 {
			scope[i]=stack.Objects[stack_Index]
		} else {
			if int_index_in_int_arr(i, function.Base_Program.Globally_Available)!=-1 && function.Base_Program.Rendered_Scope[i].Value!=nil {
				scope[i]=function.Base_Program.Rendered_Scope[i]
			} else {
				object_Abstract,ok:=function.Stack_Spec[i]
				if ok {
					Constructed_Object:=Default_Object_By_Object_Abstract(object_Abstract)
					constructed_Objects[i]=Constructed_Object
					scope[i]=&Constructed_Object
				}
			}
		}
	}
	if function.Base_Program.Is_Dynamic {
		stack_Interfaces:=make([]*interface{}, 0)
		for _,stack_Object:=range stack.Objects {
			stack_Interfaces = append(stack_Interfaces, &stack_Object.Value)
		}
		return (*function.External_Function)(stack_Interfaces)
	}
	for i := 0; i < len(function.Instructions); i++ {
		if Database_Interface.Sequential {
			if Database_Interface.Sequential_Instructions<=0 {
				Database_Interface.Sequential=false
				Database_Interface.Sequential_Instructions=0
				thread_Mutex.Channel <- 0
			} else {
				Database_Interface.Sequential_Instructions-=1
			}
		}
		instructions := function.Instructions[i]
		switch opcode := instructions[0]; opcode {
		case SET_INSTRUCTION:
			scope[instructions[1]].Value = instructions[2]
		case STRING_SET_INSTRUCTION:
			scope[instructions[1]].Value = function.Constants.STRING[instructions[2]]
		case ADD_INSTRUCTION, SUB_INSTRUCTION, MULT_INSTRUCTION, DIV_INSTRUCTION, FLOOR_INSTRUCTION, MOD_INSTRUCTION, GREATER_INSTRUCTION, SMALLER_INSTRUCTION, AND_INSTRUCTION, OR_INSTRUCTION, XOR_INSTRUCTION, EQUALS_INSTRUCTION, NEQUALS_INSTRUCTION:
			var1,is_int:=scope[instructions[1]].Value.(int)
			if !is_int {
				execution_Result.Error=errors.New("variable is not of type integer")
				return execution_Result
			}
			var2,is_int:=scope[instructions[2]].Value.(int)
			if !is_int {
				execution_Result.Error=errors.New("variable is not of type integer")
				return execution_Result
			}
			switch opcode {
			case ADD_INSTRUCTION:
				scope[instructions[3]].Value = var1 + var2
			case SUB_INSTRUCTION:
				scope[instructions[3]].Value = var1 - var2
			case DIV_INSTRUCTION, FLOOR_INSTRUCTION:
				scope[instructions[3]].Value = var1 / var2
			case MULT_INSTRUCTION:
				scope[instructions[3]].Value = var1 * var2
			case MOD_INSTRUCTION:
				scope[instructions[3]].Value = var1 % var2
			case GREATER_INSTRUCTION:
				scope[instructions[3]].Value = Bool2Int(var1 > var2)
			case SMALLER_INSTRUCTION:
				scope[instructions[3]].Value = Bool2Int(var1 < var2)
			case AND_INSTRUCTION:
				scope[instructions[3]].Value = Bool2Int((var1!=0 && var2!=0))
			case OR_INSTRUCTION:
				scope[instructions[3]].Value = Bool2Int((var1!=0 || var2!=0))
			case EQUALS_INSTRUCTION:
				scope[instructions[3]].Value = Bool2Int((var1 == var2))
			case NEQUALS_INSTRUCTION:
				scope[instructions[3]].Value = Bool2Int((var1 != var2))
			case XOR_INSTRUCTION:
				scope[instructions[3]].Value = var1 ^ var2
			}
		case RETURN_INSTRUCTION:
			execution_Result.Return_Value=scope[instructions[1]].Value
			return execution_Result
		case CALL_INSTRUCTION:
			call_Stack:=Stack{}
			function_to_be_Called:=function.Base_Program.Functions[instructions[1]]
			for i:=0; i<len(function_to_be_Called.Argument_Names); i++ {
				call_Stack.Locations = append(call_Stack.Locations, function_to_be_Called.Argument_Indexes[i])
				call_Stack.Objects = append(call_Stack.Objects, scope[instructions[3+i]])
			}
			perfomed_Execution:=Interpreter(&function_to_be_Called, call_Stack, thread_Mutex, Database_Interface)
			execution_Result.Gas_Used+=perfomed_Execution.Gas_Used
			if perfomed_Execution.Error!=nil {
				execution_Result.Error=perfomed_Execution.Error
				return execution_Result
			}
			scope[instructions[2]].Value=perfomed_Execution.Return_Value
		case DEEP_COPY_OBJECT_INSTRUCTION:
			Copied_Object:=Copy_Object(scope[instructions[2]])
			scope[instructions[1]].Value=Copied_Object.Value
		case JUMP_INSTRUCTION:
			if scope[instructions[2]].Value.(int)!=0 {
				i+=scope[instructions[1]].Value.(int)
			}
		case NOT_INSTRUCTION:
			scope[instructions[1]].Value = Bool2Int(scope[instructions[1]].Value.(int)==0)
		case JUMPTO_INSTRUCTION:
			i=scope[instructions[1]].Value.(int)-1
		case USE_DEFAULT_OBJECT_INSTRUCTION:
			scope[instructions[1]].Value=constructed_Objects[instructions[1]].Value
		case APPEND_INSTRUCTION:
			scope[instructions[3]].Value = append(scope[instructions[1]].Value.([]*Object), &Object{Value: scope[instructions[2]].Value})
		case LEN_INSTRUCTION:
			scope[instructions[2]].Value = len(scope[instructions[1]].Value.([]*Object))
		case SOFT_COPY_OBJECT_INSTRUCTION:
			scope[instructions[1]]=scope[instructions[2]]
		case ARRAY_TYPE_LOOKUP_INSTRUCTION:
			scope[instructions[3]]=scope[instructions[1]].Value.([]*Object)[scope[instructions[2]].Value.(int)]
		case DICT_TYPE_LOOKUP_INSTRUCTION:
			if instructions[4]==int(INT_TYPE) {
				dict:=scope[instructions[1]].Value.(map[int]*Object)
				key:=scope[instructions[2]].Value.(int)
				value,ok:=dict[key]
				if !ok {
					value=&Object{}
					dict[key]=value
				}
				scope[instructions[3]]=value
			}
			if instructions[4]==int(INT64_TYPE) {
				dict:=scope[instructions[1]].Value.(map[int64]*Object)
				key:=scope[instructions[2]].Value.(int64)
				value,ok:=dict[key]
				if !ok {
					value=&Object{}
					dict[key]=value
				}
				scope[instructions[3]]=value
			}
			if instructions[4]==int(STRING_TYPE) {
				dict:=scope[instructions[1]].Value.(map[string]*Object)
				key:=scope[instructions[2]].Value.(string)
				value,ok:=dict[key]
				if !ok {
					value=&Object{}
					dict[key]=value
				}
				scope[instructions[3]]=value
			}
			if instructions[4]==int(FLOAT_TYPE) {
				dict:=scope[instructions[1]].Value.(map[float32]*Object)
				key:=scope[instructions[2]].Value.(float32)
				value,ok:=dict[key]
				if !ok {
					value=&Object{}
					dict[key]=value
				}
				scope[instructions[3]]=value
			}
			if instructions[4]==int(FLOAT64_TYPE) {
				dict:=scope[instructions[1]].Value.(map[float64]*Object)
				key:=scope[instructions[2]].Value.(float64)
				value,ok:=dict[key]
				if !ok {
					value=&Object{}
					dict[key]=value
				}
				scope[instructions[3]]=value
			}
		case DB_WRITE_INSTRUCTION:
			if !Database_Interface.Sequential && int_index_in_int_arr(instructions[1], Database_Interface.Locking_Databases)!=-1 {
				thread_Mutex.Inner_Waiting=true
				Database_Interface.Sequential_Instructions=<-thread_Mutex.Channel
				thread_Mutex.Inner_Waiting=false
				Database_Interface.Sequential=true
			}
			encoded_Object, err:=Encode_Object(struct{Value interface{}}{Value: scope[instructions[3]].Value})
			if err!=nil {
				fmt.Println(err)
				execution_Result.Error=err
				return execution_Result
			}
			db_gas, err:=Database_Interface.DB_Write(instructions[1], scope[instructions[2]].Value.(string), encoded_Object)
			execution_Result.Gas_Used+=db_gas
			if err!=nil {
				fmt.Println(err)
				execution_Result.Error=err
				return execution_Result
			}
		case DB_READ_INSTRUCTION:
			if !Database_Interface.Sequential && int_index_in_int_arr(instructions[1], Database_Interface.Locking_Databases)!=-1 {
				thread_Mutex.Inner_Waiting=true
				Database_Interface.Sequential_Instructions=<-thread_Mutex.Channel
				thread_Mutex.Inner_Waiting=false
				Database_Interface.Sequential=true
			}
			encoded_Object, db_gas, err:=Database_Interface.DB_Read(instructions[1], scope[instructions[2]].Value.(string))
			execution_Result.Gas_Used+=db_gas
			if err!=nil {
				fmt.Println(err)
				execution_Result.Error=err
				return execution_Result
			}
			obj, err:=Decode_Object(encoded_Object)
			if err!=nil {
				fmt.Println(err)
				execution_Result.Error=err
				return execution_Result
			}
			scope[instructions[3]].Value=obj.Value
		case LOCK_INSTRUCTION:
			if Database_Interface.Sequential {
				Database_Interface.Sequential_Instructions=instructions[1]
			} else {
				thread_Mutex.Inner_Waiting=true
				<-thread_Mutex.Channel
				Database_Interface.Sequential_Instructions=instructions[1]
				thread_Mutex.Inner_Waiting=false
				Database_Interface.Sequential=true
			}
		case FIELD_ACCESS_INSTRUCTION:
			Out_Object:=scope[instructions[1]].Value.(map[string]*Object)[function.Constants.STRING[instructions[2]]]
			if Out_Object==nil {
				Out_Object=&Object{}
				scope[instructions[1]].Value.(map[string]*Object)[function.Constants.STRING[instructions[2]]]=Out_Object
			}
			scope[instructions[3]]=Out_Object
		}
	}
	for i := range scope {
		if scope[i]!=nil && scope[i].Value != nil {
			fmt.Print(Object_PrintS(scope[i]), " ")
		}
	}
	fmt.Println()
	return execution_Result
}