package Vengine

import (
	"github.com/lumatozer/VenGO/structs"
)

const (
	_                      int8 = iota
	INT_TYPE               int8 = iota
	INT64_TYPE             int8 = iota
	STRING_TYPE            int8 = iota
	FLOAT_TYPE             int8 = iota
	FLOAT64_TYPE           int8 = iota
	POINTER_TYPE           int8 = iota
	VOID_TYPE              int8 = iota
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
	Argument_Names         []string
	Argument_Indexes       []int
	External_Function      *func([]*interface{})structs.Execution_Result
	Constants              Constants
	Calls                  []int
	Databases              []int
}

type Constants struct {
	INT_64                 []int64
	STRING                 []string
	FLOAT                  []float32
	FLOAT64                []float64
}

type Type struct {
	Is_Array               bool
	Is_Dict                bool
	Raw_Type               int8
	Is_Struct              bool
	Struct_Details         map[string]*Type
	Child                  *Type
}

type Program struct {
	Is_Dynamic             bool
	Package_Name           string
	Functions              []Function
	Structs                map[string]*Type
	Rendered_Scope         []*Object // This Scope will be used for initalizing functions of this file + will retain all the final global states of the variables
	Object_References      []Object_Reference
	Globally_Available     []int
	Int64_Constants        []int64
	String_Constants       []string
	Float_Constants        []float32
	Float64_Constants      []float64
	Dependencies           []string
}

type Function_Definition struct {
	Name                   string
	Arguments_Variables    map[string]Token
	Out_Token              Token
	Instruction_Tokens     []Token
	Arguments              []string
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
	PackageName            string
	Imports                [][]string
	Structs                []Struct_Definition
	Functions              []Function_Definition
	Variables              []Variable_Definition
}