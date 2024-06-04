package main

import (
	"errors"
	"fmt"
)

// "errors"
// "fmt"
// "os"
// "path/filepath"

const (
	INT_TYPE               int8 = iota
	INT64_TYPE             int8 = iota
	STRING_TYPE            int8 = iota
	FLOAT_TYPE             int8 = iota
	FLOAT64_TYPE           int8 = iota
	POINTER_TYPE           int8 = iota
)

type Object struct {
	Value                  interface{}
}

type Object_Reference struct {
	Aliases                []string
	Object_Type            Type
}

type Object_Abstract struct {
	Is_Mapping             bool
	Is_Array               bool
	Is_Raw                 bool
	Raw_Type               int8
}

type Function struct {
	Name                   string
	Stack_Spec             map[int]Object_Abstract   // initialize these objects with default properties with new pointers for this scope
	Instructions           [][]int // Instruction set for this function
	Arguments              map[string]Type
	Out_Type               Type
	Base_Program           *Program
	Variable_Scope         map[string]int
}

type Type struct {
	Is_Array               bool
	Is_Dict                bool
	Raw_Type               int8
	Is_Struct              bool
	Is_Pointer             bool
	Struct_Details         map[string]*Type
	Child                  *Type
}

type Program struct {
	Functions              []*Function
	Structs                map[string]map[string]*Type
	Rendered_Scope         []*Object // This Scope will be used for initalizing functions of this file + will retain all the final global states of the variables
	Object_References      []Object_Reference
	Globally_Available     []int

	Int64_Constants        []int64
	String_Constants       []string
	Float_Constants        []float32
	Float64_Constants      []float64
}

type Function_Definition struct {
	Name                   string
	Arguments_Variables    map[string]Token
	Out_Token              Token
	Instruction_Tokens     []Token
}

type Variable_Definition struct {
	Names                  []string
	Type_Token             Token
}

type Struct_Definition struct {
	Name                   string
	Fields_Token           map[string]Token
}

type Definitions struct {
	Imports                [][]string
	Structs                []Struct_Definition
	Functions              []Function_Definition
	Variables              []Variable_Definition
}

func Definition_Parser(code []Token) (Definitions, error) {
	definitions:=Definitions{
		
	}
	Structs:=make([]string, 0)
	global_Variables:=make([]string, 0)
	Functions:=make([]string, 0)
	for i:=0; i<len(code); i++ {
		if code[i].Type=="sys" && code[i].Value=="var" {
			if !(len(code)-i>=4) {
				return definitions, errors.New("invalid variable declaration statement")
			}
			variable_Definition:=Variable_Definition{}
			j:=i
			for {
				j++
				if j>=len(code) {
					return definitions, errors.New("unexpected EOF while parsing variable declaration statement")
				}
				if code[j].Type=="type" {
					variable_Definition.Type_Token=code[j]
					break
				}
				if code[j].Type!="variable" {
					return definitions, errors.New("expected token of type 'variable' during variable definition got '"+code[j].Type+"'")
				}
				if !Is_Valid_Variable_Name(code[j].Value) {
					return definitions, errors.New("invalid variable name '"+code[j].Value+"'")
				}
				if str_index_in_str_arr(code[j].Value, global_Variables)!=-1 {
					return definitions, errors.New("Variable '"+code[j].Value+"' has already been initialized")
				}
				variable_Definition.Names = append(variable_Definition.Names, code[j].Value)
				global_Variables = append(global_Variables, code[j].Value)
			}
			if len(variable_Definition.Names)==0 {
				return definitions, errors.New("invalid variable definition structure")
			}
			definitions.Variables = append(definitions.Variables, variable_Definition)
			i=j+1
			continue
		}
		if code[i].Type=="sys" && code[i].Value=="struct" {
			if !(len(code)-i>=5) {
				return definitions, errors.New("invalid struct declaration statement")
			}
			struct_Definition:=Struct_Definition{
				Fields_Token: make(map[string]Token),
			}
			if code[i+1].Type!="variable" {
				return definitions, errors.New("expected token of type 'variable' during struct definition got '"+code[i+1].Type+"'")
			}
			if !Is_Valid_Variable_Name(code[i+1].Value) {
				return definitions, errors.New("invalid struct name '"+code[i+1].Value+"'")
			}
			if str_index_in_str_arr(code[i+1].Value, Structs)!=-1 {
				return definitions, errors.New("struct '"+code[i+1].Value+"' has already been declared")
			}
			struct_Definition.Name=code[i+1].Value
			Structs = append(Structs, code[i+1].Value)
			j:=i+1
			brackets:=0
			field_Names:=make([]string, 0)
			for {
				j+=1
				if j>=len(code) {
					return definitions, errors.New("unexpected EOF while parsing struct declaration statement")
				}
				if code[j].Type=="bracket" {
					if code[j].Value=="{" {
						brackets+=1
						continue
					}
					if code[j].Value=="}" {
						brackets-=1
						if brackets==0 {
							break
						}
						continue
					}
				}
				if code[j].Type=="variable" {
					if !Is_Valid_Variable_Name(code[j].Value) {
						return definitions, errors.New("invalid field name '"+code[j].Value+"'")
					}
					if str_index_in_str_arr(code[j].Value, field_Names)!=-1 {
						return definitions, errors.New("field '"+code[j].Value+"' has already been declared")
					}
					field_Names = append(field_Names, code[j].Value)
					j+=1
					if j>=len(code) {
						return definitions, errors.New("unexpected EOF while parsing struct declaration statement")
					}
					if code[j].Type!="type" {
						return definitions, errors.New("expected token of type 'type' during struct's field defintion got '"+code[j].Type+"'")
					}
					struct_Definition.Fields_Token[field_Names[len(field_Names)-1]]=code[j]
				}
				if brackets==0 {
					break
				}
			}
			i=j
			definitions.Structs = append(definitions.Structs, struct_Definition)
			continue
		}
		if code[i].Type=="sys" && (code[i].Value=="function" || code[i].Value=="fn") {
			function_Definition:=Function_Definition{
				Arguments_Variables: make(map[string]Token),
			}
			if !(len(code)-i>=6) {
				return definitions, errors.New("invalid function declaration statement")
			}
			if code[i+1].Type!="variable" {
				return definitions, errors.New("expected token of type 'variable' during function declaration got type '"+code[i+1].Type+"'")
			}
			if !Is_Valid_Variable_Name(code[i+1].Value) {
				return definitions, errors.New("invalid function name '"+code[i+1].Value+"'")
			}
			if str_index_in_str_arr(code[i+1].Value, Functions)!=-1 {
				return definitions, errors.New("function '"+code[i+1].Value+"' has already been declared")
			}
			Functions = append(Functions, code[i+1].Value)
			function_Definition.Name=code[i+1].Value
			if code[i+2].Type!="bracket" || code[i+2].Value!="(" {
				return definitions, errors.New("invalid function declaration statement")
			}
			j:=i+1
			brackets:=0
			argument_Names:=make([]string, 0)
			last_comma:=false
			for {
				j++
				if j>=len(code) {
					return definitions, errors.New("Unexpected EOF while parsing function declaration")
				}
				if code[j].Type=="bracket" {
					if code[j].Value=="(" {
						brackets+=1
					}
					if code[j].Value==")" {
						brackets-=1
						if brackets==0 {
							break
						}
					}
					last_comma=false
					continue
				}
				if code[j].Type=="variable" {
					if !Is_Valid_Variable_Name(code[j].Value) {
						return definitions, errors.New("invalid field name '"+code[j].Value+"'")
					}
					if str_index_in_str_arr(code[j].Value, argument_Names)!=-1 {
						return definitions, errors.New("field '"+code[j].Value+"' has already been declared")
					}
					argument_Names = append(argument_Names, code[j].Value)
					j+=1
					if j>=len(code) {
						return definitions, errors.New("unexpected EOF while parsing struct declaration statement")
					}
					if code[j].Type!="type" {
						return definitions, errors.New("expected token of type 'type' during struct's field defintion got '"+code[j].Type+"'")
					}
					function_Definition.Arguments_Variables[argument_Names[len(argument_Names)-1]]=code[j]
					last_comma=false
					continue
				}
				if code[j].Type=="comma" {
					if last_comma {
						return definitions, errors.New("invalid function declaration statement")
					}
					last_comma=true
					continue
				}
			}
			if code[j+1].Type!="type" {
				return definitions, errors.New("invalid function declaration statement")
			}
			function_Definition.Out_Token=code[j+1]
			if code[j+2].Type!="bracket" || code[j+2].Value!="{" {
				return definitions, errors.New("invalid function declaration statement")
			}
			j+=1
			brackets=0
			for {
				j++
				if j>=len(code) {
					return definitions, errors.New("Unexpected EOF while parsing function declaration")
				}
				if code[j].Type=="bracket" {
					if code[j].Value=="{" {
						brackets+=1
					}
					if code[j].Value=="}" {
						brackets-=1
					}
					if brackets==0 {
						break
					}
					continue
				}
				function_Definition.Instruction_Tokens = append(function_Definition.Instruction_Tokens, code[j])
			}
			i=j
			definitions.Functions = append(definitions.Functions, function_Definition)
			continue
		}
		return definitions, errors.New("unexpected token of type '"+code[i].Type+"'")
	}
	return definitions, nil
} 

func Parser(code []Token) (Program, error) {
	program:=Program{
		Structs: make(map[string]map[string]*Type),
	}
	definitions,err:=Definition_Parser(code)
	if err!=nil {
		return program, err
	}
	fmt.Println(definitions)
	return program, nil
}