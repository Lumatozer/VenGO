package main

import "strings"

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

func Type_Token_To_Struct(Type_Token Token) Type {
	if Type_Token.Type=="type" {
		return Type_Token_To_Struct(Type_Token.Tok_Children[0])
	}
	if Type_Token.Type=="array" {
		new_Type:=Type_Token_To_Struct(Type_Token.Tok_Children[0])
		return Type{Is_Array: true, Child: &new_Type}
	}
	if Type_Token.Type=="dict" {
		new_Type:=Type_Token_To_Struct(Type_Token.Tok_Children[0])
		return Type{Is_Dict: true, Raw_Type: Type_Token.Str_Children[0], Child: &new_Type}
	}
	return Type{Raw_Type: Type_Token.Value}
}