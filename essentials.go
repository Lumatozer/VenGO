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

func Type_Token_To_Struct(Type_Token Token, program Program) (Type, error) {
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
		return Type{Is_Dict: true, Raw_Type: Type_Token.Str_Children[0], Child: &new_Type}, nil
	}
	if str_index_in_str_arr(Type_Token.Value, []string{"string", "bytes", "int", "int64", "float", "float64"})!=-1 || program.Structs[Type_Token.Value]!=nil {
		return Type{Raw_Type: Type_Token.Value}, nil
	} else {
		return Type{}, errors.New("Invalid Type")
	}
}