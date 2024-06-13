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

type Stack struct {
	Locations       []int
	Objects         []*Object
}

func Interpreter(function *Function, stack Stack) (int, interface{}) {
	gas:=0
	scope := function.Base_Program.Rendered_Scope
	for i:=0; i<len(scope); i++ {
		stack_Index:=int_index_in_int_arr(i, stack.Locations)
		if stack_Index!=-1 {
			scope[i]=stack.Objects[stack_Index]
		} else {
			if int_index_in_int_arr(i, function.Base_Program.Globally_Available)!=-1 && function.Base_Program.Rendered_Scope[i].Value!=nil {
				scope[i]=function.Base_Program.Rendered_Scope[i]
			} else {
				object_Value:=function.Base_Program.Rendered_Scope[i].Value
				scope[i]=&Object{Value: Copy_Interface(object_Value)}
			}
		}
	}
	if function.Base_Program.Is_External {
		if function.External_Function==nil {
			// add minimum gas required to call a function
			return 0, nil
		}
		return function.External_Function(stack.Objects)
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
				fmt.Println(errors.New("variable is not of type integer"))
				return -1, nil
			}
			var2,is_int:=scope[instructions[2]].Value.(int)
			if !is_int {
				fmt.Println(errors.New("variable is not of type integer"))
				return -1, nil
			}
			scope[instructions[3]].Value = var1+var2
		}
		if opcode == RETURN_INSTRUCTION {
			return gas, scope[instructions[1]].Value
		}
		if opcode == CALL_INSTRUCTION {
			call_Stack:=Stack{}
			function_to_be_Called:=function.Base_Program.Functions[instructions[1]]
			for i:=0; i<len(function_to_be_Called.Argument_Names); i++ {
				call_Stack.Locations = append(call_Stack.Locations, function_to_be_Called.Argument_Indexes[i])
				call_Stack.Objects = append(call_Stack.Objects, function.Base_Program.Rendered_Scope[instructions[3+i]])
			}
			gas_Used,return_Value:=Interpreter(&function_to_be_Called, call_Stack)
			if gas==-1 {
				return -1, nil
			}
			gas+=gas_Used
			scope[instructions[2]].Value=return_Value
		}
	}
	for i := range scope {
		if scope[i]!=nil && scope[i].Value != nil {
			fmt.Print(scope[i].Value, " ")
		}
	}
	fmt.Println()
	return gas, nil
}

// to add struct and pointer capabilities to objects