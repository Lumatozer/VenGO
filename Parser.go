package main

import (
	"errors"
	"fmt"
)

type Object struct {
	Type                   Type
	Location               int
	Callable_Functions     []int
	Int_Mapping            map[int]*Object
	Int64_Mapping          map[int64]*Object
	Float64_Mapping        map[float32]*Object
	Float_Mapping          map[float64]*Object
	Field_Children         map[int]*Object
	Children               []*Object
}

type Function struct {
	Id                     int
	Name                   string
	Stack_Spec             []int   // initialize these objects with default properties with new pointers for this scope
	Instructions           [][]int // Instruction set for this function
	Arguments              map[string]Type
	Out_Type               Type
	Base_Scope             Type // for initializing a new Function scope under the program this function belongs to. Will always be the Program's Rendered Scope as only the arguments change
}

type Scope struct {
	Ip                     int
	Objects                []*Object
}

type Type struct {
	Is_Array               bool
	Is_Dict                bool
	Raw_Type               string
	Child                  *Type
}

type Program struct {
	Functions              []*Function
	Structs                map[string]map[string]Type
	Rendered_Scope         Scope // This Scope will be used for initalizing functions of this file + will retain all the final global states of the variables
	State_Variables        []int // Indices of variables to be stored on the blockchain for this program
}

type Execution struct {
	Gas_Limit              int64
	Entry_Program          *Program
	Entry_Function         *Function
	Programs               []Program
}

func Parse_Program(code []Token) error {
	for i := 0; i < len(code); i++ {
		if code[i].Type == "sys" && code[i].Value == "struct" && len(code)-i >= 6 && code[i+1].Type=="variable" {
			struct_tokens:=make([]Token, 0)
			normal_exit:=false
			br_count:=0
			for j:=i+2; j<len(code); j++ {
				if code[j].Type=="bracket" && code[j].Value=="{" {
					br_count+=1
				}
				if code[j].Type=="bracket" && code[j].Value=="}" {
					br_count-=1
				}
				struct_tokens = append(struct_tokens, code[j])
				if br_count==0 {
					normal_exit=true
					break
				}
			}
			if !normal_exit || len(struct_tokens)<2 || len(struct_tokens)%2!=0 {
				return errors.New("Unexpected EOF while parsing struct")
			}
			struct_tokens=struct_tokens[1:len(struct_tokens)-1]
			this_struct:=make(map[string]Type)
			for j:=0; j<len(struct_tokens); j+=2 {
				if struct_tokens[j].Type!="variable" || !Is_Valid_Variable_Name(struct_tokens[j].Value) {
					return errors.New("Invalid field name\""+struct_tokens[j].Value+"\"")
				}
				if struct_tokens[j+1].Type!="type" {
					return errors.New("Invalid token for type")
				}
				fmt.Println(this_struct, Type_Token_To_Struct(struct_tokens[j+1]))
			}
		}
	}
	return nil
}