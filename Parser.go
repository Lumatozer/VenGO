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
	Base_Scope             *Scope // for initializing a new Function scope under the program this function belongs to. Will always be the Program's Rendered Scope as only the arguments change
	Int_Constants          []int
	Int64_Constants        []int64
	String_Constants       []string
	Float_Constants        []float32
	Float64_Constants      []float64
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

func Parse_Program(code []Token, importing []string) (Program, error) {
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
			files_to_read:=make(map[string]*Program)
			module_names:=make([]string, 0)
			for j:=0; j<len(import_tokens); j+=3 {
				if import_tokens[j].Type!="string" || import_tokens[j+1].Type!="sys" || import_tokens[j+1].Value!="as" || import_tokens[j+2].Type!="variable" || !Is_Valid_Variable_Name(import_tokens[j+2].Value) {
					return program, errors.New("invalid syntax for importing")
				}
				if str_index_in_str_arr(filepath.Clean(import_tokens[j+2].Value), importing)!=-1 {
					return program, errors.New("circular imports detected")
				}
				import_tokens[j].Value=filepath.Clean(import_tokens[j].Value)
				module_names = append(module_names, import_tokens[j+2].Value)
				file_data,err:=os.ReadFile(import_tokens[j].Value)
				if err!=nil {
					return program, err
				}
				file_tokens,err:=Tokenizer(string(file_data))
				if err!=nil {
					return program, err
				}
				file_program, err:=Parse_Program(file_tokens, append(importing, import_tokens[j+2].Value))
				if err!=nil {
					return program, err
				}
				files_to_read[import_tokens[j+2].Value]=&file_program
			}
			for _,module:=range module_names {
				for module_struct:=range files_to_read[module].Structs {
					program.Structs[module+"."+module_struct]=files_to_read[module].Structs[module_struct]
				}
			}
			for _,module:=range module_names {
				for _,function:=range files_to_read[module].Functions {
					function.Name=module+"."+function.Name
					program.Functions = append(program.Functions, function)
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
				to_initialise:=true
				new_index:=0
				for index,Obj:=range program.Rendered_Scope.Objects {
					if Obj.Name==variable_token.Value {
						if Compare_Type(variable_Type, Obj.Type) {
							to_initialise=false
							new_index=index
							break
						}
						return program, errors.New("Variable \""+Obj.Name+"\" has already been initialised")
					}
				}
				if to_initialise {
					program.Rendered_Scope.Objects = append(program.Rendered_Scope.Objects, Initialise_Object(variable_token.Value, variable_Type, &program))
					new_index=len(program.Rendered_Scope.Objects)-1
				}
				program.State_Variables = append(program.State_Variables, new_index)
			}
			i+=len(variable_tokens)+1
			continue
		}
		if code[i].Type=="sys" && code[i].Value=="function" && len(code)-i>=7 {
			// to add bounds check later on
			if code[i+1].Type!="variable" || !Is_Valid_Variable_Name(code[i+1].Value) {
				return program, errors.New("Invalid function name")
			}
			for j:=0; j<len(program.Functions); j++ {
				if program.Functions[j].Name==code[i+1].Value {
					return program, errors.New("Function has already been declared")
				}
			}
			if code[i+2].Type!="bracket" || code[i+2].Value!="(" {
				return program, errors.New("Invalid function declaration syntax")
			}
			argument_tokens:=make([]Token, 0)
			normal_exit:=false
			for j:=i+3; j<len(code); j++ {
				if code[j].Type=="bracket" && code[j].Value==")" {
					normal_exit=true
					break
				}
				argument_tokens = append(argument_tokens, code[j])
			}
			if !normal_exit || len(argument_tokens)%2!=0 {
				return program, errors.New("Invalid function declaration")
			}
			argument_variables:=make(map[string]Type)
			for j:=0; j<len(argument_tokens); j+=2 {
				if argument_tokens[j].Type!="variable" || !Is_Valid_Variable_Name(argument_tokens[j].Value) {
					return program, errors.New("Invalid argument name")
				}
				if argument_tokens[j+1].Type!="type" {
					return program, errors.New("Invalid argument name")
				}
				variable_Type,err:=Type_Token_To_Struct(argument_tokens[j+1], &program)
				if err!=nil {
					return program, err
				}
				for key:=range argument_variables {
					if key==argument_tokens[j].Value {
						return program, errors.New("Cannot initialize multiple variables with the same name")
					}
				}
				to_initialise:=true
				for _,Obj:=range program.Rendered_Scope.Objects {
					if Obj.Name==argument_tokens[j].Value {
						if Compare_Type(variable_Type, Obj.Type) {
							to_initialise=false
							break
						}
						return program, errors.New("Variable \""+Obj.Name+"\" has already been initialised with another Type")
					}
				}
				argument_variables[argument_tokens[j].Value]=variable_Type
				if to_initialise {
					program.Rendered_Scope.Objects = append(program.Rendered_Scope.Objects, Initialise_Object(argument_tokens[j].Value, variable_Type, &program))
				}
			}
			if i+len(argument_tokens)+6>=len(code) {
				return program, errors.New("Unexpected EOF")
			}
			if code[i+len(argument_tokens)+4].Type!="type" {
				return program, errors.New("Invalid function declaration")
			}
			function_Type,err:=Type_Token_To_Struct(code[i+len(argument_tokens)+4], &program)
			if err!=nil {
				return program, err
			}
			function_tokens:=make([]Token, 0)
			normal_exit=false
			br_count:=0
			if code[i+len(argument_tokens)+5].Type!="bracket" || code[i+len(argument_tokens)+5].Value!="{" {
				return program, errors.New("Invalid function declaration")
			}
			for j:=i+len(argument_tokens)+5; j<len(code); j++ {
				if code[j].Type=="bracket" && code[j].Value=="{" {
					br_count+=1
				}
				if code[j].Type=="bracket" && code[j].Value=="}" {
					br_count-=1
				}
				if br_count==0 {
					normal_exit=true
					break
				}
				function_tokens = append(function_tokens, code[j])
			}
			if !normal_exit {
				return program, errors.New("Unexpected EOF")
			}
			this_Function:=Function{
				Name: code[i+1].Value,
				Arguments: argument_variables,
				Out_Type: function_Type,
				Base_Scope: &program.Rendered_Scope,
			}
			if len(function_tokens)>1 {
				err=Parse_Instructions_For_Function(function_tokens[1:len(function_tokens)-1], &this_Function, &program)
				if err!=nil {
					return program, err
				}
			}
			program.Rendered_Scope.Objects = append(program.Rendered_Scope.Objects, Initialise_Object(this_Function.Name+"."+"return", this_Function.Out_Type, &program))
			program.Functions = append(program.Functions, &this_Function)
			i+=len(argument_tokens)+len(function_tokens)+5
			continue
		}
		fmt.Println("Unexpected Token", code[i])
		return program, errors.New("Unexpected Token")
	}
	fmt.Println(program)
	return program, nil
}

func Parse_Instructions_For_Function(code []Token, function *Function, program *Program) error {
	function.Stack_Spec = append(function.Stack_Spec, len(program.Rendered_Scope.Objects)-1)
	lines:=make([][]Token, 0)
	this_line:=make([]Token, 0)
	for i:=0; i<len(code); i++ {
		if code[i].Type=="semicolon" {
			if len(this_line)>0 {
				lines = append(lines, this_line)
			}
			this_line=make([]Token, 0)
			continue
		}
		this_line = append(this_line, code[i])
	}
	if len(this_line)>0 {
		lines = append(lines, this_line)
	}
	for _,line:=range lines {
		if len(line)>=3 && line[0].Type=="sys" && line[0].Value=="var" {
			object_Type,err:=Type_Token_To_Struct(line[len(line)-1], program)
			if err!=nil {
				return err
			}
			for _,variable:=range line[1:len(line)-1] {
				if variable.Type!="variable" {
					return errors.New("Invalid variable initialisation syntax")
				}
				to_initialise:=true
				new_index:=0
				for index,Obj:=range program.Rendered_Scope.Objects {
					if Obj.Name==variable.Value {
						if Compare_Type(object_Type, Obj.Type) {
							to_initialise=false
							new_index=index
							break
						}
						return errors.New("Variable \""+Obj.Name+"\" has already been initialised")
					}
				}
				if to_initialise {
					program.Rendered_Scope.Objects = append(program.Rendered_Scope.Objects, Initialise_Object(variable.Value, object_Type, program))
					new_index=len(program.Rendered_Scope.Objects)-1
				}
				function.Stack_Spec = append(function.Stack_Spec, new_index)
			}
			continue
		}
		
		if len(line)==3 && line[0].Type=="sys" && line[0].Value=="set" {
			if line[1].Type!="variable" {
				return errors.New("Invalid instruction")
			}
			if line[2].Type!="number" {
				return errors.New("Invalid instruction")
			}
			to_use_index:=-1
			for index,Object:=range program.Rendered_Scope.Objects {
				if Object.Name==line[1].Value {
					
					if int_index_in_int_arr(index, program.State_Variables)!=-1 || int_index_in_int_arr(index, function.Stack_Spec)!=-1 {
						to_use_index=index
					} else {
						return errors.New("Variable not in scope")
					}
				}
			}
			if to_use_index==-1 {
				return errors.New("Variable not found")
			}
			function.Int_Constants = append(function.Int_Constants, int(line[2].Float64_Constant))
			function.Instructions = append(function.Instructions, []int{SET_INSTRUCTION, to_use_index, len(function.Int_Constants)-1})
			continue
		}
	}
	return nil
}