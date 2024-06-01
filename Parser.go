package main

import (
	"errors"
	"strconv"
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
	Names                  string
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

func Definition_Parser(code []Token) Definitions {
	definitions:=Definitions{
		
	}
	return definitions
} 

func Parser(code []Token) (Program, error) {
	line:=0
	program:=Program{
		Structs: make(map[string]map[string]Type),
	}
	global_Variable_names:=make([]string, 0)
	for i:=0; i<len(code); i++ {
		if code[i].Type=="sys" && code[i].Value=="var" {
			if !(len(code)-i>=4) {
				return program, errors.New("incomplete variable declaration at line "+strconv.FormatInt(int64(line), 10))
			}
			variable_declaration_tokens:=make([]Token, 0)
			j:=i
			for {
				j++
				if j>=len(code) {
					return program, errors.New("unexpected EOF during variable declaration at line "+strconv.FormatInt(int64(line), 10))
				}
				if code[j].Type=="semicolon" {
					break
				}
				variable_declaration_tokens = append(variable_declaration_tokens, code[j])
			}
			if !(len(variable_declaration_tokens)>=2) {
				return program, errors.New("incomplete variable declaration at line "+strconv.FormatInt(int64(line), 10))
			}
			if variable_declaration_tokens[len(variable_declaration_tokens)-1].Type!="type" {
				return program, errors.New("expected a type token got "+variable_declaration_tokens[len(variable_declaration_tokens)-1].Type+" at line "+strconv.FormatInt(int64(line), 10))
			}
			variable_Type,err:=Type_Token_To_Struct(variable_declaration_tokens[len(variable_declaration_tokens)-1], &program)
			if err!=nil {
				return program, errors.New(err.Error()+" at line "+strconv.FormatInt(int64(line), 10))
			}
			for _,token:=range variable_declaration_tokens[:len(variable_declaration_tokens)-1] {
				if token.Type!="variable" {
					return program, errors.New("expected token of type variable got '"+token.Type+"' at line "+strconv.FormatInt(int64(line), 10))
				}
				if !Is_Valid_Variable_Name(token.Value) {
					return program, errors.New("invalid variable name '"+token.Value+"' at line "+strconv.FormatInt(int64(line), 10))
				}
				if str_index_in_str_arr(token.Value, global_Variable_names)!=-1 {
					return program, errors.New("variable '"+token.Value+"' has already been defined at line "+strconv.FormatInt(int64(line), 10))
				}
				global_Variable_names = append(global_Variable_names, token.Value)
				program.Rendered_Scope = append(program.Rendered_Scope, nil)
				program.Globally_Available = append(program.Globally_Available, len(program.Rendered_Scope)-1)
				program.Object_References = append(program.Object_References, Object_Reference{Aliases: []string{token.Value}, Object_Type: variable_Type})
			}
			i+=len(variable_declaration_tokens)+1
			continue
		}
		return program, errors.New("unexpected token of type '"+code[i].Type+"' at line "+strconv.FormatInt(int64(line), 10))
	}
	return program, nil
}