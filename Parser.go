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
	Struct_Details         map[string]Type
	Child                  *Type
}

type Program struct {
	Functions              []*Function
	Structs                map[string]map[string]Type
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
	Argument_Tokens        map[string][]Token
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
	global_Variables:=make([]string, 0)
	for i:=0; i<len(code); i++ {
		if code[i].Type=="sys" && code[i].Value=="var" && len(code)-i>=4 {
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
					return definitions, errors.New("expected token of type 'variable' got '"+code[j].Type+"'")
				}
				if !Is_Valid_Variable_Name(code[j].Value) {
					return definitions, errors.New("invalid variable name")
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
		return definitions, errors.New("Unexpected token of type '"+code[i].Type+"'")
	}
	return definitions, nil
} 

func Parser(code []Token) (Program, error) {
	program:=Program{
		Structs: make(map[string]map[string]Type),
	}
	definitions,err:=Definition_Parser(code)
	if err!=nil {
		return program, err
	}
	fmt.Println(definitions)
	return program, nil
}