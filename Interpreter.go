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
	out := Execution_Result{}
	scope := function.Base_Program.Rendered_Scope
	for stack_index, object := range stack {
		scope[stack_index] = object
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
	}
	for i := range scope {
		if scope[i].Value != nil {
			fmt.Println(scope[i].Value)
		}
	}
	return out
}