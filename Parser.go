package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Object struct {
	Name                   string
	Type                   Type
	Location               int
	Int_Mapping            map[int]*Object
	Int64_Mapping          map[int64]*Object
	String_Mapping         map[string]*Object
	Float_Mapping          map[float32]*Object
	Float64_Mapping        map[float64]*Object
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
	Int_Objects            []int
	Int64_Objects          []int64
	String_Objects         []string
	Float_Objects          []float32
	Float64_Objects        []float64
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

func Parse_Program(code []Token, imported []string) (Program, error) {
	program:=Program{
		Structs: make(map[string]map[string]Type),
		Rendered_Scope: Scope{},
	}
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
				return program, errors.New("Unexpected EOF while parsing struct")
			}
			struct_tokens=struct_tokens[1:len(struct_tokens)-1]
			this_struct:=make(map[string]Type)
			for j:=0; j<len(struct_tokens); j+=2 {
				if struct_tokens[j].Type!="variable" || !Is_Valid_Variable_Name(struct_tokens[j].Value) {
					return program, errors.New("Invalid field name\""+struct_tokens[j].Value+"\"")
				}
				if struct_tokens[j+1].Type!="type" {
					return program, errors.New("Invalid token for type")
				}
				out_Type_struct,err:=Type_Token_To_Struct(struct_tokens[j+1], &program)
				if err!=nil {
					return program, err
				}
				this_struct[struct_tokens[j].Value]=out_Type_struct
			}
			program.Structs[code[i+1].Value]=this_struct
			i+=len(struct_tokens)+2+1
			continue
		}
		if code[i].Type=="sys" && code[i].Value=="import" && len(code)-i>=6 && code[i+1].Type=="bracket" && code[i+1].Value=="(" {
			import_tokens:=make([]Token, 0)
			normal_exit:=false
			for j:=i+2; j<len(code); j++ {
				if code[j].Type=="bracket" && code[j].Value==")" {
					normal_exit=true
					break
				}
				import_tokens = append(import_tokens, code[j])
			}
			if !normal_exit || len(import_tokens)<3 || len(import_tokens)%3!=0 {
				return program, errors.New("Unexpected EOF while parsing import statement")
			}
			files_to_read:=make(map[string]Program)
			module_names:=make([]string, 0)
			for j:=0; j<len(import_tokens); j+=3 {
				if import_tokens[j].Type!="string" || import_tokens[j+1].Type!="sys" || import_tokens[j+1].Value!="as" || import_tokens[j+2].Type!="variable" || !Is_Valid_Variable_Name(import_tokens[j+2].Value) {
					return program, errors.New("invalid syntax for importing")
				}
				if str_index_in_str_arr(filepath.Clean(import_tokens[j+2].Value), imported)!=-1 {
					return program, errors.New("circular imports detected")
				}
				import_tokens[j].Value=filepath.Clean(import_tokens[j].Value)
				module_names = append(module_names, import_tokens[j+2].Value)
				imported = append(imported, import_tokens[j+2].Value)
				file_data,err:=os.ReadFile(import_tokens[j].Value)
				if err!=nil {
					return program, err
				}
				file_tokens,err:=Tokenizer(string(file_data))
				if err!=nil {
					return program, err
				}
				file_program, err:=Parse_Program(file_tokens, imported)
				if err!=nil {
					return program, err
				}
				files_to_read[import_tokens[j+2].Value]=file_program
			}
			for _,module:=range module_names {
				for module_struct:=range files_to_read[module].Structs {
					program.Structs[module+"."+module_struct]=files_to_read[module].Structs[module_struct]
				}
			}
			i+=2+len(import_tokens)
			continue
		}
		if code[i].Type=="sys" && code[i].Value=="var" && len(code)-1>=4 {
			variable_tokens:=make([]Token, 0)
			normal_exit:=false
			for j:=i+1; j<len(code); j++ {
				if code[j].Type=="semicolon" {
					normal_exit=true
					break
				}
				variable_tokens = append(variable_tokens, code[j])
			}
			if !normal_exit || len(variable_tokens)<2 || variable_tokens[len(variable_tokens)-1].Type!="type" {
				return program, errors.New("Unexpected EOF while parsing import statement")
			}
			variable_Type,err:=Type_Token_To_Struct(variable_tokens[len(variable_tokens)-1], &program)
			if err!=nil {
				return program, err
			}
			for _,variable_token:=range variable_tokens[:len(variable_tokens)-1] {
				if variable_token.Type!="variable" || !Is_Valid_Variable_Name(variable_token.Value) {
					return program, errors.New("Invalid variable name")
				}
				for _,Obj:=range program.Rendered_Scope.Objects {
					if Obj.Name==variable_token.Value {
						return program, errors.New("Variable \""+Obj.Name+"\" has already been initialised")
					}
				}
				program.Rendered_Scope.Objects = append(program.Rendered_Scope.Objects, Initialise_Object(variable_token.Value, variable_Type, &program))
				program.State_Variables = append(program.State_Variables, len(program.Rendered_Scope.Objects)-1)
			}
			i+=len(variable_tokens)+1
			continue
		}
		fmt.Println("Unexpected Token", code[i])
		return program, errors.New("Unexpected Token")
	}
	fmt.Println(program)
	return program, nil
}