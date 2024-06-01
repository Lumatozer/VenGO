package main

// import "fmt"

// func Interpreter(entry *Function) Object {
// 	return_Object:=Object{}
// 	function_scope := Copy_Scope(entry.Base_Scope)
// 	for _, index := range entry.Stack_Spec {
// 		new_Object := Object{Name: function_scope.Objects[index].Name, Type: function_scope.Objects[index].Type}
// 		to_copy:=false
// 		for _,local_var:=range entry.Local_Variables {
// 			if *local_var==index {
// 				to_copy=true
// 			}
// 		}
// 		if to_copy {
// 			new_Object=*Deep_Copy(function_scope.Objects[index])
// 		} else {
// 			Initialise_Object_Mapping(&new_Object)
// 			Initialise_Object_Values(&new_Object)
// 		}
// 		function_scope.Objects[index] = &new_Object
// 	}
// 	for i := 0; i < len(entry.Instructions); i++ {
// 		bytecode := entry.Instructions[i]
// 		switch operator := bytecode[0]; operator {
// 		case SET_INSTRUCTION:
// 			*function_scope.Objects[bytecode[1]].Int_Value = entry.Int_Constants[bytecode[2]]
// 		case RETURN_INSTRUCTION:
// 			i=len(entry.Instructions)
// 		case CALL_INSTRUCTION:
// 			Interpreter(entry.Base_Program.Functions[bytecode[1]])
// 		}
// 	}
// 	covered:=make([]*int, 0)
// 	for _,object:=range function_scope.Objects {
// 		found:=false
// 		for _,covered_Object:=range covered {
// 			if covered_Object==object.Int_Value {
// 				found=true
// 			}
// 		}
// 		if !found {
// 			if object.Int_Value!=nil {
// 				fmt.Print(*object.Int_Value, " ")
// 			}
// 			covered = append(covered, object.Int_Value)
// 		}
// 	}
// 	fmt.Println()
// 	return return_Object
// }