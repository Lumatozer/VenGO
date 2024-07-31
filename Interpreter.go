package main

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

func Interpreter(function *Function, stack Stack) structs.Execution_Result {
	execution_Result:=structs.Execution_Result{}
	scope := make([]*Object, len(function.Base_Program.Rendered_Scope))
	for i:=0; i<len(scope); i++ {
		scope[i] = function.Base_Program.Rendered_Scope[i]
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
	if function.Base_Program.Is_Dynamic {
		stack_Interfaces:=make([]*interface{}, 0)
		for _,stack_Object:=range stack.Objects {
			stack_Interfaces = append(stack_Interfaces, &stack_Object.Value)
		}
		return (*function.External_Function)(stack_Interfaces)
	}
	for i := 0; i < len(function.Instructions); i++ {
		instructions := function.Instructions[i]
		switch opcode := instructions[0]; opcode {
		case SET_INSTRUCTION:
			scope[instructions[1]].Value = instructions[2]
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
			perfomed_Execution:=Interpreter(&function_to_be_Called, call_Stack)
			execution_Result.Gas_Used+=perfomed_Execution.Gas_Used
			if perfomed_Execution.Error!=nil {
				execution_Result.Error=perfomed_Execution.Error
				return execution_Result
			}
			scope[instructions[2]].Value=perfomed_Execution.Return_Value
		}
	}
	for i := range scope {
		if scope[i]!=nil && scope[i].Value != nil {
			fmt.Print(scope[i].Value, " ")
		}
	}
	fmt.Println()
	return execution_Result
}