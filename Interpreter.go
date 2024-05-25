package main

import "fmt"

func Copy_Scope(scope *Scope) Scope {
	copied_Scope := Scope{}
	copied_Scope.Float64_Objects = scope.Float64_Objects
	copied_Scope.Float_Objects = scope.Float_Objects
	copied_Scope.Int64_Objects = scope.Int64_Objects
	copied_Scope.Int_Objects = scope.Int_Objects
	copied_Scope.String_Objects = scope.String_Objects
	copied_Scope.Filename = scope.Filename
	for _, obj := range scope.Objects {
		copied_Scope.Objects = append(copied_Scope.Objects, obj)
	}
	return copied_Scope
}

func Interpreter(entry *Function) Object {
	fmt.Println("hi")
	return_Object:=Object{}
	function_scope := Copy_Scope(entry.Base_Scope)
	for _, index := range entry.Stack_Spec {
		new_Object := Object{Name: function_scope.Objects[index].Name, Type: function_scope.Objects[index].Type, Location: function_scope.Objects[index].Location}
		Initialise_Object_Mapping(&new_Object)
		function_scope.Objects[index] = &new_Object
	}
	for i := 0; i < len(entry.Instructions); i++ {
		bytecode := entry.Instructions[i]
		switch operator := bytecode[0]; operator {
		case SET_INSTRUCTION:
			function_scope.Int_Objects[function_scope.Objects[bytecode[1]].Location] = entry.Int_Constants[bytecode[2]]
		case RETURN_INSTRUCTION:
			i=len(entry.Instructions)
		case CALL_INSTRUCTION:
			Interpreter(entry.Base_Program.Functions[bytecode[1]])
		}
	}
	fmt.Println(function_scope.Int_Objects, function_scope.Filename)
	return return_Object
}