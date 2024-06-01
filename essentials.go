package main

import (
	"errors"
	"strings"
)

func str_index_in_str_arr(a string, b []string) int {
	for i := 0; i < len(b); i++ {
		if b[i] == a {
			return i
		}
	}
	return -1
}

func int_index_in_int_arr(a int, b []int) int {
	for i := 0; i < len(b); i++ {
		if b[i] == a {
			return i
		}
	}
	return -1
}


func Can_access(index int, arr_len int) bool {
	if index < 0 {
		return false
	}
	return arr_len > index
}

func Is_Valid_Variable_Name(name string) bool {
	if strings.Contains(name, ".") {
		return false
	}
	return true
}

func String_Type_To_Int8(string_Type string) int8 {
	if string_Type=="int" {
		return INT_TYPE
	}
	if string_Type=="int64" {
		return INT64_TYPE
	}
	if string_Type=="string" {
		return STRING_TYPE
	}
	if string_Type=="float" {
		return FLOAT_TYPE
	}
	if string_Type=="float64" {
		return FLOAT64_TYPE
	}
	return 0
}

func Type_Token_To_Struct(Type_Token Token, program *Program) (Type, error) {
	if Type_Token.Type=="type" {
		return Type_Token_To_Struct(Type_Token.Tok_Children[0], program)
	}
	if Type_Token.Type=="array" {
		new_Type,err:=Type_Token_To_Struct(Type_Token.Tok_Children[0], program)
		if err!=nil {
			return Type{}, err
		}
		return Type{Is_Array: true, Child: &new_Type}, nil
	}
	if Type_Token.Type=="dict" {
		new_Type,err:=Type_Token_To_Struct(Type_Token.Tok_Children[0], program)
		if err!=nil {
			return Type{}, err
		}
		return Type{Is_Dict: true, Raw_Type: String_Type_To_Int8(Type_Token.Str_Children[0]), Child: &new_Type}, nil
	}
	if str_index_in_str_arr(Type_Token.Value, []string{"string", "bytes", "int", "int64", "float", "float64", "void"})!=-1 {
		return Type{Raw_Type: String_Type_To_Int8(Type_Token.Value)}, nil
	} else if program.Structs[Type_Token.Value]!=nil {
		return Type{Struct_Details: program.Structs[Type_Token.Value]}, nil
	} else {
		return Type{}, errors.New("invalid type '"+Type_Token.Value+"'")
	}
}

func Copy_Function(function *Function) *Function {
	copied_Function:=*function
	return &copied_Function
}

func Default_Object_By_Type(variable_Type Type) interface{} {
	if variable_Type.Raw_Type==INT_TYPE {
		return int(0)
	}
	if variable_Type.Raw_Type==INT64_TYPE {
		return int64(0)
	}
	if variable_Type.Raw_Type==STRING_TYPE {
		return ""
	}
	if variable_Type.Raw_Type==FLOAT_TYPE {
		return float32(0)
	}
	if variable_Type.Raw_Type==FLOAT64_TYPE {
		return float64(0)
	}
	if variable_Type.Is_Array {
		return make([]*Object, 0)
	}
	if variable_Type.Is_Dict {
		if variable_Type.Raw_Type==INT_TYPE {
			return make(map[int]*Object)
		}
		if variable_Type.Raw_Type==INT64_TYPE {
			return make(map[int64]*Object)
		}
		if variable_Type.Raw_Type==STRING_TYPE {
			return make(map[string]*Object)
		}
		if variable_Type.Raw_Type==FLOAT_TYPE {
			return make(map[float32]*Object)
		}
		if variable_Type.Raw_Type==FLOAT64_TYPE {
			return make(map[float64]*Object)
		}
	}
	return int(0)
}