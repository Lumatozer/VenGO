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
	if str_index_in_str_arr(Type_Token.Value, []string{"string", "bytes", "int", "int64", "float", "float64", "void"})!=-1 || program.Structs[Type_Token.Value]!=nil {
		return Type{Raw_Type: String_Type_To_Int8(Type_Token.Value)}, nil
	} else {
		return Type{}, errors.New("Invalid Type")
	}
}

func Initialise_Object_Mapping(obj *Object) {
	if obj.Type.Is_Dict {
		if obj.Type.Raw_Type==STRING_TYPE {
			mapping:=make(map[string]*Object)
			obj.String_Mapping=&mapping
		}
		if obj.Type.Raw_Type==INT_TYPE {
			mapping:=make(map[int32]*Object)
			obj.Int_Mapping=&mapping
		}
		if obj.Type.Raw_Type==INT64_TYPE {
			mapping:=make(map[int64]*Object)
			obj.Int64_Mapping=&mapping
		}
		if obj.Type.Raw_Type==FLOAT_TYPE {
			mapping:=make(map[float32]*Object)
			obj.Float_Mapping=&mapping
		}
		if obj.Type.Raw_Type==FLOAT64_TYPE {
			mapping:=make(map[float64]*Object)
			obj.Float64_Mapping=&mapping
		}
	}
}

func Initialise_Object_Values(obj *Object) {
	if !obj.Type.Is_Dict && !obj.Type.Is_Array {
		if obj.Type.Raw_Type==STRING_TYPE {
			Value:=""
			obj.String_Value=&Value
		}
		if obj.Type.Raw_Type==INT_TYPE {
			Value:=int32(0)
			obj.Int_Value=&Value
		}
		if obj.Type.Raw_Type==INT64_TYPE {
			Value:=int64(0)
			obj.Int64_Value=&Value
		}
		if obj.Type.Raw_Type==FLOAT_TYPE{
			Value:=float32(0)
			obj.Float_Value=&Value
		}
		if obj.Type.Raw_Type==FLOAT64_TYPE {
			Value:=float64(0)
			obj.Float64_Value=&Value
		}
	}
	if obj.Type.Is_Array {
		Object_Array:=make([]*Object, 0)
		obj.Children=&Object_Array
	}
}

func Initialise_Object(Object_Name string, Object_Type Type, program *Program) *Object {
	new_Object:=Object{Name: &Object_Name, Type: &Object_Type}
	Initialise_Object_Mapping(&new_Object)
	Initialise_Object_Values(&new_Object)
	return &new_Object
}

func Compare_Type(a Type, b Type) bool {
	if a.Is_Array==a.Is_Dict && a.Is_Dict==b.Is_Dict && b.Raw_Type==a.Raw_Type {
		if a.Child==b.Child && a.Child==nil {
			return true
		}
		return Compare_Type(*a.Child, *b.Child)
	}
	return false
}

func Shallow_Copy(object *Object) *Object {
	copied_Object:=Object{}
	copied_Object.Name=object.Name
	copied_Object.Type=object.Type
	copied_Object.Int_Mapping=object.Int_Mapping
	copied_Object.Int64_Mapping=object.Int64_Mapping
	copied_Object.String_Mapping=object.String_Mapping
	copied_Object.Float_Mapping=object.Float_Mapping
	copied_Object.Float64_Mapping=object.Float64_Mapping
	copied_Object.Field_Children=object.Field_Children
	copied_Object.Children=object.Children
	copied_Object.Int_Value=object.Int_Value
	copied_Object.Int64_Value=object.Int64_Value
	copied_Object.String_Value=object.String_Value
	copied_Object.Float_Value=object.Float_Value
	copied_Object.Float64_Value=object.Float64_Value
	return &copied_Object
}

func Deep_Copy(object *Object) *Object {
	deep_Copied_Object:=Object{Name: object.Name, Type: object.Type}

	if object.Int_Mapping!=nil {
		Int_Mapping:=make(map[int32]*Object)
		for key,obj:=range *object.Int_Mapping {
			Int_Mapping[key]=obj
		}
		deep_Copied_Object.Int_Mapping=&Int_Mapping
	}

	if object.Int64_Mapping!=nil {
		Int64_Mapping:=make(map[int64]*Object)
		for key,obj:=range *object.Int64_Mapping {
			Int64_Mapping[key]=obj
		}
		deep_Copied_Object.Int64_Mapping=&Int64_Mapping
	}

	if object.String_Mapping!=nil {
		String_Mapping:=make(map[string]*Object)
		for key,obj:=range *object.String_Mapping {
			String_Mapping[key]=obj
		}
		deep_Copied_Object.String_Mapping=&String_Mapping
	}

	if object.Float_Mapping!=nil {
		Float_Mapping:=make(map[float32]*Object)
		for key,obj:=range *object.Float_Mapping {
			Float_Mapping[key]=obj
		}
		deep_Copied_Object.Float_Mapping=&Float_Mapping
	}

	if object.Float64_Mapping!=nil {
		Float64_Mapping:=make(map[float64]*Object)
		for key,obj:=range *object.Float64_Mapping {
			Float64_Mapping[key]=obj
		}
		deep_Copied_Object.Float64_Mapping=&Float64_Mapping
	}
	
	if object.Int_Value!=nil {
		Int_Value:=*object.Int_Value
		deep_Copied_Object.Int_Value=&Int_Value
	}

	if object.Int64_Value!=nil {
		Int64_Value:=*object.Int64_Value
		deep_Copied_Object.Int64_Value=&Int64_Value
	}

	if object.String_Value!=nil {
		String_Value:=*object.String_Value
		deep_Copied_Object.String_Value=&String_Value
	}

	if object.Float_Value!=nil {
		Float_Value:=*object.Float_Value
		deep_Copied_Object.Float_Value=&Float_Value
	}

	if object.Float64_Value!=nil {
		Float64_Value:=*object.Float64_Value
		deep_Copied_Object.Float64_Value=&Float64_Value
	}
	return &deep_Copied_Object
}

func Copy_Function(function *Function) *Function {
	copied_Function:=*function
	return &copied_Function
}