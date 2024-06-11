package main

import (
	"errors"
	"fmt"
)

type Execution_Result struct {
	Gas_Used        int
	Return_Value    interface{}
	Error           error
}

func Interpreter(function *Function, stack map[int]*Object) Execution_Result {
	fmt.Println(function.Base_Program.Rendered_Scope)
	out := Execution_Result{}
	scope := function.Base_Program.Rendered_Scope
	stack_Indices:=make([]int, 0)
	for stack_index := range stack {
		stack_Indices = append(stack_Indices, stack_index)
	}
	for i:=0; i<len(scope); i++ {
		stack_Index:=int_index_in_int_arr(i, stack_Indices)
		if stack_Index!=-1 {
			scope[i]=stack[stack_Index]
		} else {
			object_Value:=function.Base_Program.Rendered_Scope[i].Value
			scope[i]=&Object{Value: Copy_Interface(object_Value)}
		}
	}
	for i := 0; i < len(function.Instructions); i++ {
		instructions := function.Instructions[i]
		opcode := instructions[0]
		if opcode == SET_INSTRUCTION {
			scope[instructions[1]].Value = instructions[2]
		}
		if opcode == ADD_INSTRUCTION {
			var1,is_int:=scope[instructions[1]].Value.(int)
			if !is_int {
				out.Error=errors.New("variable is not of type integer")
				return out
			}
			var2,is_int:=scope[instructions[2]].Value.(int)
			if !is_int {
				out.Error=errors.New("variable is not of type integer")
				return out
			}
			scope[instructions[3]].Value = var1+var2
		}
		if opcode == RETURN_INSTRUCTION {
			out.Return_Value=scope[instructions[1]].Value
			break
		}
		if opcode == CALL_INSTRUCTION {
			call_Stack:=make(map[int]*Object)
			function_to_be_Called:=function.Base_Program.Functions[instructions[1]]
			for i:=0; i<len(function_to_be_Called.Argument_Names); i++ {
				object_Value:=function.Base_Program.Rendered_Scope[instructions[3+i]].Value
				fmt.Println("putting", object_Value)
			}
			execution_Result:=Interpreter(&function_to_be_Called, call_Stack)
			if execution_Result.Error!=nil {
				out.Error=execution_Result.Error
				return out
			}
			out.Gas_Used+=execution_Result.Gas_Used
			scope[instructions[2]].Value=execution_Result.Return_Value
		}
	}
	for i := range scope {
		if scope[i]!=nil && scope[i].Value != nil {
			fmt.Println(scope[i].Value)
		}
	}
	fmt.Println()
	return out
}

// to add struct capabilities to object_abstract