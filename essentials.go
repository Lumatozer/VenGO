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
		return Type{Is_Dict: true, Raw_Type: Type_Token.Str_Children[0], Child: &new_Type}, nil
	}
	if str_index_in_str_arr(Type_Token.Value, []string{"string", "bytes", "int", "int64", "float", "float64"})!=-1 || program.Structs[Type_Token.Value]!=nil {
		return Type{Raw_Type: Type_Token.Value}, nil
	} else {
		return Type{}, errors.New("Invalid Type")
	}
}

func Get_Object_New_Location(Object_Type Type, program *Program) int {
	if !Object_Type.Is_Array && !Object_Type.Is_Dict {
		if Object_Type.Raw_Type=="string" {
			program.Rendered_Scope.String_Objects = append(program.Rendered_Scope.String_Objects, "")
			return len(program.Rendered_Scope.String_Objects)
		}
		if Object_Type.Raw_Type=="int" {
			program.Rendered_Scope.Int_Objects = append(program.Rendered_Scope.Int_Objects, 0)
			return len(program.Rendered_Scope.String_Objects)
		}
		if Object_Type.Raw_Type=="int64" {
			program.Rendered_Scope.Int64_Objects = append(program.Rendered_Scope.Int64_Objects, 0)
			return len(program.Rendered_Scope.String_Objects)
		}
		if Object_Type.Raw_Type=="float" {
			program.Rendered_Scope.Float_Objects = append(program.Rendered_Scope.Float_Objects, 0)
			return len(program.Rendered_Scope.String_Objects)
		}
		if Object_Type.Raw_Type=="float64" {
			program.Rendered_Scope.Float64_Objects = append(program.Rendered_Scope.Float64_Objects, 0)
			return len(program.Rendered_Scope.String_Objects)
		}
	}
	return -1
}

func Initialise_Object_Mapping(obj *Object) {
	if obj.Type.Is_Dict {
		if obj.Type.Raw_Type=="string" {
			obj.String_Mapping=make(map[string]*Object)
		}
		if obj.Type.Raw_Type=="int" {
			obj.Int_Mapping=make(map[int]*Object)
		}
		if obj.Type.Raw_Type=="int64" {
			obj.Int64_Mapping=make(map[int64]*Object)
		}
		if obj.Type.Raw_Type=="float" {
			obj.Float_Mapping=make(map[float32]*Object)
		}
		if obj.Type.Raw_Type=="float64" {
			obj.Float64_Mapping=make(map[float64]*Object)
		}
	}
}

func Initialise_Object(Object_Name string, Object_Type Type, program *Program) *Object {
	new_Object:=Object{Name: Object_Name, Type: Object_Type, Location: Get_Object_New_Location(Object_Type, program)}
	Initialise_Object_Mapping(&new_Object)
	return &new_Object
}