package main

import (
	"errors"
	"fmt"
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

func Type_Token_To_Struct(Type_Token Token, program *Program) (*Type, error) {
	if Type_Token.Type=="type" {
		return Type_Token_To_Struct(Type_Token.Tok_Children[0], program)
	}
	if Type_Token.Type=="array" {
		new_Type,err:=Type_Token_To_Struct(Type_Token.Tok_Children[0], program)
		if err!=nil {
			return &Type{}, err
		}
		return& Type{Is_Array: true, Child: new_Type}, nil
	}
	if Type_Token.Type=="dict" {
		new_Type,err:=Type_Token_To_Struct(Type_Token.Tok_Children[0], program)
		if err!=nil {
			return &Type{}, err
		}
		return &Type{Is_Dict: true, Raw_Type: String_Type_To_Int8(Type_Token.Str_Children[0]), Child: new_Type}, nil
	}
	if Type_Token.Type=="pointer" {
		rendered_Type,err:=Type_Token_To_Struct(Type_Token.Tok_Children[0], program)
		if err!=nil {
			return &Type{}, err
		}
		return &Type{Is_Pointer: true, Child: rendered_Type, Raw_Type: POINTER_TYPE}, nil
	}
	if str_index_in_str_arr(Type_Token.Value, []string{"string", "bytes", "int", "int64", "float", "float64", "void"})!=-1 {
		return &Type{Raw_Type: String_Type_To_Int8(Type_Token.Value), Is_Raw: true}, nil
	} else if program.Structs[Type_Token.Value]!=nil {
		return program.Structs[Type_Token.Value], nil
	} else {
		return &Type{}, errors.New("invalid type '"+Type_Token.Value+"'")
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

func Type_Struct_To_Object_Abstract(Type_Object Type) Object_Abstract {
	if Type_Object.Is_Array {
		return Object_Abstract{Is_Array: true}
	}
	if Type_Object.Is_Dict {
		return Object_Abstract{Is_Mapping: true}
	}
	if Type_Object.Is_Pointer {
		return Object_Abstract{Is_Raw: true, Raw_Type: POINTER_TYPE}
	}
	return Object_Abstract{Is_Array: true, Raw_Type: Type_Object.Raw_Type}
}

func Equal_Type(a *Type, b *Type) bool {
	if a.Is_Array!=b.Is_Array {
		fmt.Println(1)
		return false
	}
	if a.Is_Dict!=b.Is_Dict {
		fmt.Println(2)
		return false
	}
	if a.Is_Pointer!=b.Is_Pointer {
		fmt.Println(3)
		return false
	}
	if a.Is_Raw!=b.Is_Raw {
		fmt.Println(4)
		return false
	}
	if a.Is_Struct!=b.Is_Struct {
		fmt.Println(5)
		return false
	}
	if a.Raw_Type!=b.Raw_Type {
		fmt.Println(6, a.Raw_Type, b.Raw_Type)
		return false
	}
	if a.Child!=nil && b.Child!=nil {
		return Equal_Type(a.Child, b.Child)
	}
	if a.Is_Struct && b.Is_Struct {
		for field, field1_Type:=range a.Struct_Details {
			field2_Type, found:=b.Struct_Details[field]
			if !found {
				return false
			}
			return Equal_Type(field1_Type, field2_Type)
		}
	}
	return true
}