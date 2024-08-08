package Vengine

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"github.com/lumatozer/VenGO/structs"
	"github.com/lumatozer/VenGO/venc"
	"os"
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
	if len(strings.Trim(name, "."))==0 {
		return false
	}
	for _,char:=range "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890_." {
		name=strings.ReplaceAll(name, string(char), "")
	}
	return len(name)==0
}

func Is_Package_Name_Valid(name string) bool {
	for _,char:=range "abcdefghijklmnopqrstuvwxyz_ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		name=strings.ReplaceAll(name, string(char), "")
	}
	return len(name)==0
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
	if string_Type=="void" {
		return VOID_TYPE
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
		return &Type{Child: rendered_Type, Raw_Type: POINTER_TYPE}, nil
	}
	if str_index_in_str_arr(Type_Token.Value, []string{"string", "bytes", "int", "int64", "float", "float64", "void"})!=-1 {
		return &Type{Raw_Type: String_Type_To_Int8(Type_Token.Value)}, nil
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
	return int(0)
}

func Type_Struct_To_Object_Abstract(Type_Object Type) Object_Abstract {
	if Type_Object.Is_Array {
		return Object_Abstract{Is_Array: true, Raw_Type: Type_Object.Raw_Type}
	}
	if Type_Object.Is_Dict {
		return Object_Abstract{Is_Mapping: true, Raw_Type: Type_Object.Raw_Type}
	}
	if Type_Object.Raw_Type!=0 {
		return Object_Abstract{Raw_Type: Type_Object.Raw_Type}
	}
	return Object_Abstract{}
}

func Hash(a string) string {
	hashed_Bytes:=make([]byte, 0)
	for _,b:=range sha256.Sum256([]byte(a)) {
		hashed_Bytes = append(hashed_Bytes, b)
	}
	return hex.EncodeToString(hashed_Bytes)
}

func Type_Signature(a *Type, traversed []*Type) string {
	if a.Is_Array {
		return Hash("array"+Type_Signature(a.Child, traversed))
	}
	if a.Is_Dict {
		return Hash("dict:key->"+strconv.FormatInt(int64(a.Raw_Type), 10)+":value->"+Type_Signature(a.Child, traversed))
	}
	if a.Raw_Type!=0 {
		return Hash("raw"+strconv.FormatInt(int64(a.Raw_Type), 10))
	}
	if a.Raw_Type==POINTER_TYPE {
		return Hash("pointer"+Type_Signature(a.Child, traversed))
	}
	if a.Is_Struct {
		for _,traveresed_Struct:=range traversed {
			if traveresed_Struct==a {
				return Hash("struct-recursed-at"+strconv.FormatInt(int64(len(traversed)), 10))
			}
		}
		traversed = append(traversed, a)
		struct_Keys:=make([]string, 0)
		for Field:=range a.Struct_Details {
			struct_Keys = append(struct_Keys, Field)
		}
		slices.Sort(struct_Keys)
		out:="struct{"
		for _,Field:=range struct_Keys {
			out+=Field+"->"+Type_Signature(a.Struct_Details[Field], traversed)
		}
		out+="}"
		return Hash(out)
	}
	return ""
}

func Equal_Type(a *Type, b *Type) bool {
	return Type_Signature(a, make([]*Type, 0))==Type_Signature(b, make([]*Type, 0))
}

func Default_Object_By_Object_Abstract(object_Abstract Object_Abstract) Object {
	if object_Abstract.Is_Array {
		fmt.Println("\n\n\n\n", "BRO WHO TF?\n\n\n\n")
		return Object{Value: make([]*Object, 0)}
	}
	if object_Abstract.Is_Mapping {
		if object_Abstract.Raw_Type==INT_TYPE {
			return Object{Value: make(map[int]*Object)}
		}
		if object_Abstract.Raw_Type==INT64_TYPE {
			return Object{Value: make(map[int64]*Object)}
		}
		if object_Abstract.Raw_Type==STRING_TYPE {
			return Object{Value: make(map[string]*Object)}
		}
		if object_Abstract.Raw_Type==FLOAT_TYPE {
			return Object{Value: make(map[float32]*Object)}
		}
		if object_Abstract.Raw_Type==FLOAT64_TYPE {
			return Object{Value: make(map[float64]*Object)}
		}
	}
	if object_Abstract.Raw_Type!=0 {
		if object_Abstract.Raw_Type==INT_TYPE {
			return Object{Value: int(0)}
		}
		if object_Abstract.Raw_Type==INT64_TYPE {
			return Object{Value: int64(0)}
		}
		if object_Abstract.Raw_Type==STRING_TYPE {
			return Object{Value: ""}
		}
		if object_Abstract.Raw_Type==FLOAT_TYPE {
			return Object{Value: float32(0)}
		}
		if object_Abstract.Raw_Type==FLOAT64_TYPE {
			return Object{Value: float64(0)}
		}
	}
	return Object{}
}

func Copy_Object(object *Object) Object  {
	a:=object.Value
	_,ok:=a.(int)
	if ok {
		return Object{Value: a}
	}
	_,ok=a.(int64)
	if ok {
		return Object{Value: a}
	}
	_,ok=a.(string)
	if ok {
		return Object{Value: a}
	}
	_,ok=a.(float32)
	if ok {
		return Object{Value: a}
	}
	_,ok=a.(float64)
	if ok {
		return Object{Value: a}
	}
	_,ok=a.(map[int]*Object)
	if ok {
		Object_Value:=make(map[int]*Object)
		for Map_Key,Map_Item:=range a.(map[int]*Object) {
			Object_Value[Map_Key]=Map_Item
		}
		return Object{Value: Object_Value}
	}
	_,ok=a.(map[int64]*Object)
	if ok {
		Object_Value:=make(map[int64]*Object)
		for Map_Key,Map_Item:=range a.(map[int64]*Object) {
			Object_Value[Map_Key]=Map_Item
		}
		return Object{Value: Object_Value}
	}
	_,ok=a.(map[string]*Object)
	if ok {
		Object_Value:=make(map[string]*Object)
		for Map_Key,Map_Item:=range a.(map[string]*Object) {
			Object_Value[Map_Key]=Map_Item
		}
		return Object{Value: Object_Value}
	}
	_,ok=a.(map[float32]*Object)
	if ok {
		Object_Value:=make(map[float32]*Object)
		for Map_Key,Map_Item:=range a.(map[float32]*Object) {
			Object_Value[Map_Key]=Map_Item
		}
		return Object{Value: Object_Value}
	}
	_,ok=a.(map[float64]*Object)
	if ok {
		Object_Value:=make(map[float64]*Object)
		for Map_Key,Map_Item:=range a.(map[float64]*Object) {
			Object_Value[Map_Key]=Map_Item
		}
		return Object{Value: Object_Value}
	}
	_,ok=a.(*Object)
	if ok {
		return Object{Value: a}
	}
	_,ok=a.([]*Object)
	if ok {
		Children:=make([]*Object, 0)
		for _,Child:=range a.([]*Object) {
			Children = append(Children, Child)
		}
		return Object{Value: Children}
	}
	return Object{Value: nil}
}

func Print_Mapping(Keys []interface{}, Values []*Object) string {
	out:="{"
	for i:=range Keys {
		out+=fmt.Sprint(Keys[i])+"->"+Object_PrintS(Values[i])+", "
	}
	out=strings.Trim(out, ", ")
	return out+"}"
}


func Object_PrintS(o *Object) string {
	_,ok:=o.Value.(int)
	if ok {
		return fmt.Sprint(o.Value)
	}
	_,ok=o.Value.(int64)
	if ok {
		return fmt.Sprint(o.Value)
	}
	_,ok=o.Value.(string)
	if ok {
		return fmt.Sprint(o.Value)
	}
	_,ok=o.Value.(float32)
	if ok {
		return fmt.Sprint(o.Value)
	}
	_,ok=o.Value.(float64)
	if ok {
		return fmt.Sprint(o.Value)
	}
	_,ok=o.Value.(*Object)
	if ok {
		return "&{"+Object_PrintS(o.Value.(*Object))+"}"
	}
	_,ok=o.Value.(map[int]*Object)
	if ok {
		keys:=make([]interface{}, 0)
		values:=make([]*Object, 0)
		for key,value:=range o.Value.(map[int]*Object) {
			keys = append(keys, key)
			values = append(values, value)
		}
		return Print_Mapping(keys, values)
	}
	_,ok=o.Value.(map[int64]*Object)
	if ok {
		keys:=make([]interface{}, 0)
		values:=make([]*Object, 0)
		for key,value:=range o.Value.(map[int64]*Object) {
			keys = append(keys, key)
			values = append(values, value)
		}
		return Print_Mapping(keys, values)
	}
	_,ok=o.Value.(map[string]*Object)
	if ok {
		keys:=make([]interface{}, 0)
		values:=make([]*Object, 0)
		for key,value:=range o.Value.(map[string]*Object) {
			keys = append(keys, key)
			values = append(values, value)
		}
		return Print_Mapping(keys, values)
	}
	_,ok=o.Value.(map[float32]*Object)
	if ok {
		keys:=make([]interface{}, 0)
		values:=make([]*Object, 0)
		for key,value:=range o.Value.(map[float32]*Object) {
			keys = append(keys, key)
			values = append(values, value)
		}
		return Print_Mapping(keys, values)
	}
	_,ok=o.Value.(map[float64]*Object)
	if ok {
		keys:=make([]interface{}, 0)
		values:=make([]*Object, 0)
		for key,value:=range o.Value.(map[float64]*Object) {
			keys = append(keys, key)
			values = append(values, value)
		}
		return Print_Mapping(keys, values)
	}
	_,ok=o.Value.([]*Object)
	if ok {
		out:="[ "
		for _,Child:=range o.Value.([]*Object) {
			out+=Object_PrintS(Child)+", "
		}
		out=strings.Trim(out, ", ")
		return out+" ]"
	}
	return ""
}

func Load_Packages(program *Program, packages []structs.Package) {
	if program.Is_Dynamic {
		for _,Package:=range packages {
			if Package.Name==program.Package_Name {
				for i,function:=range program.Functions {
					external_Function,ok:=Package.Functions[function.Name]
					if ok {
						*program.Functions[i].External_Function=external_Function
					}
				}
			}
		}
	}
	for _,Function:=range program.Functions {
		if Function.Base_Program!=program {
			Load_Packages(Function.Base_Program, packages)
		}
	}
}

func Convert_Raw_Types_For_Vitality(a int8) int8 {
	if a==INT_TYPE {
		return Venc.INT_TYPE
	}
	if a==INT64_TYPE {
		return Venc.INT64_TYPE
	}
	if a==STRING_TYPE {
		return Venc.STRING_TYPE
	}
	if a==FLOAT_TYPE {
		return Venc.FLOAT_TYPE
	}
	if a==FLOAT64_TYPE {
		return Venc.FLOAT64_TYPE
	}
	if a==POINTER_TYPE {
		return Venc.POINTER_TYPE
	}
	if a==VOID_TYPE {
		return Venc.VOID_TYPE
	}
	return 0
}

func VASM_Type_To_Vitality_Type(a *Type) *Venc.Type {
	if a==nil {
		return &Venc.Type{}
	}
	out:=&Venc.Type{Is_Array: a.Is_Array, Is_Dict: a.Is_Dict, Is_Raw: a.Raw_Type!=0, Raw_Type: Convert_Raw_Types_For_Vitality(a.Raw_Type), Is_Struct: a.Is_Struct, Is_Pointer: a.Raw_Type==POINTER_TYPE, Child: VASM_Type_To_Vitality_Type(a.Child), Struct_Details: make(map[string]*Venc.Type)}
	for Field,Field_Type:=range a.Struct_Details {
		out.Struct_Details[Field]=VASM_Type_To_Vitality_Type(Field_Type)
	}
	return out
}

func VASM_Program_To_Vitality_Program(program Program, path string) Venc.Program {
	venc_Program:=Venc.Program{Vitality: false, Path: path, Package_Name: program.Package_Name, Structs: make(map[string]*Venc.Type), Functions: make(map[string]*Venc.Function), Global_Variables: make(map[string]*Venc.Type), Imported_Libraries: make(map[string]*Venc.Program)}
	for Struct:=range program.Structs {
		venc_Program.Structs[Struct]=VASM_Type_To_Vitality_Type(program.Structs[Struct])
	}
	for _,Function:=range program.Functions {
		venc_Program.Functions[Function.Name]=&Venc.Function{Out_Type: VASM_Type_To_Vitality_Type(&Function.Out_Type), Arguments: make([]struct{Name string; Type *Venc.Type}, 0), Scope: make(map[string]*Venc.Type), Instructions: make([][]string, 0)}
		for Function_Argument:=range Function.Arguments {
			Argument_Type:=Function.Arguments[Function_Argument]
			venc_Program.Functions[Function.Name].Arguments = append(venc_Program.Functions[Function.Name].Arguments, struct{Name string; Type *Venc.Type}{Name: Function_Argument, Type: VASM_Type_To_Vitality_Type(&Argument_Type)})
		}
	}
	for _,Dependency:=range program.Dependencies {
		venc_Program.Imported_Libraries[Dependency]=&Venc.Program{Path: Dependency}
	}
	return venc_Program
}

func VASM_Translator(path string) (Venc.Program, error) {
	data,err:=os.ReadFile(path)
	if err!=nil {
		fmt.Println(err)
		return Venc.Program{}, err
	}
	tokens,err:=Tokenizer(string(data))
	if err!=nil {
		fmt.Println(err)
		return Venc.Program{}, err
	}
	VASM_Program, err:=Parser(tokens, path, make(map[string]Program))
	if err!=nil {
		return Venc.Program{}, err
	}
	return VASM_Program_To_Vitality_Program(VASM_Program, path), nil
}