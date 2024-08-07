package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/lumatozer/VenGO/structs"
)

// "errors"
// "fmt"
// "os"
// "path/filepath"

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

func Definition_Parser(code []Token, codePath string) (Definitions, error) {
	definitions:=Definitions{
		
	}
	Structs:=make([]string, 0)
	global_Variables:=make([]string, 0)
	Functions:=make([]string, 0)
	imported_Aliases:=make([]string, 0)
	imported_Files:=make([]string, 0)
	if len(code)<2 {
		return definitions, errors.New("missing package declaration during file parsing")
	}
	if code[0].Type!="sys" || code[0].Value!="package" || code[1].Type!="variable" || !Is_Valid_Variable_Name(code[1].Value) {
		return definitions, errors.New("invalid package declaration during file parsing")
	}
	definitions.PackageName=code[1].Value
	for i:=2; i<len(code); i++ {
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
				if str_index_in_str_arr(code[j].Value, global_Variables)!=-1 {
					return definitions, errors.New("Variable '"+code[j].Value+"' has already been initialized")
				}
				if !Is_Valid_Variable_Name(code[j].Value) {
					return definitions, errors.New("variable name is invalid")
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
		if code[i].Type=="sys" && code[i].Value=="import" {
			if !(len(code)-i>=6) {
				return definitions, errors.New("invalid variable declaration statement")
			}
			brackets:=0
			j:=i
			last_Token:=Token{}
			file_Path:=""
			for {
				j++
				if j>=len(code) {
					return definitions, errors.New("unexpected EOF while parsing variable declaration statement")
				}
				if code[j].Type=="bracket" {
					if code[j].Value=="(" {
						if j!=i+1 {
							return definitions, errors.New("invalid import declaration statement")
						}
						brackets+=1
						last_Token=code[j]
						continue
					}
					if code[j].Value==")" {
						if last_Token.Type!="variable" {
							return definitions, errors.New("invalid import declaration statement")
						}
						brackets-=1
						if brackets==0 {
							break
						}
						last_Token=code[j]
						continue
					}
					return definitions, errors.New("invalid import declaration statement")
				}
				if code[j].Type=="string" {
					if j!=i+2 && last_Token.Type!="variable" {
						return definitions, errors.New("invalid import declaration statement")
					}
					relative_Path,err:=filepath.Abs(filepath.Dir(codePath)+"/"+code[j].Value)
					if err!=nil {
						return definitions, err
					}
					if str_index_in_str_arr(relative_Path, imported_Files)!=-1 {
						return definitions, errors.New("same file '"+relative_Path+"' being imported multiple times")
					}
					imported_Files = append(imported_Files, relative_Path)
					file_Path=relative_Path
					last_Token=code[j]
					continue
				}
				if code[j].Type=="sys" {
					if last_Token.Type!="string" {
						return definitions, errors.New("invalid import declaration statementx")
					}
					if code[j].Value!="as" {
						return definitions, errors.New("invalid import declaration statement")
					}
					last_Token=code[j]
					continue
				}
				if code[j].Type=="variable" {
					if last_Token.Type!="sys" {
						return definitions, errors.New("invalid import declaration statement")
					}
					if str_index_in_str_arr(code[j].Value, imported_Aliases)!=-1 {
						return definitions, errors.New("same module alias '"+code[j].Value+"' used twice")
					}
					imported_Aliases = append(imported_Aliases, code[j].Value)
					if !Is_Package_Name_Valid(code[j].Value) {
						return definitions, errors.New("package import alias is invalid")
					}
					definitions.Imports = append(definitions.Imports, []string{file_Path, code[j].Value})
					last_Token=code[j]
					continue
				}
			}
			i=j
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
						return definitions, errors.New("invalid variable name '"+code[j].Value+"'")
					}
					if str_index_in_str_arr(code[j].Value, argument_Names)!=-1 {
						return definitions, errors.New("field '"+code[j].Value+"' has already been declared")
					}
					argument_Names = append(argument_Names, code[j].Value)
					j+=1
					if j>=len(code) {
						return definitions, errors.New("unexpected EOF while parsing function declaration statement")
					}
					if code[j].Type!="type" {
						return definitions, errors.New("expected token of type 'type' during function defintion got '"+code[j].Type+"'")
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
					if code[j].Value!="{" && code[j].Value!="}" {
						function_Definition.Instruction_Tokens = append(function_Definition.Instruction_Tokens, code[j])
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

func Get_Token_Struct_Dependence(token Token) (Token, error) {
	if token.Type=="type" {
		return Get_Token_Struct_Dependence(token.Tok_Children[0])
	}
	if token.Type=="pointer" {
		return token, nil
	}
	if token.Type=="array" {
		return Get_Token_Struct_Dependence(token.Tok_Children[0])
	}
	if token.Type=="dict" {
		return Get_Token_Struct_Dependence(token.Tok_Children[0])
	}
	if token.Type=="raw" {
		return token, nil
	}
	return Token{}, errors.New("Unable to resolve token dependence")
}

func Get_Struct_Definition_Refers(struct_Definition Struct_Definition) ([]string, error) {
	out_Dependencies:=make([]string, 0)
	for field:=range struct_Definition.Fields_Token {
		dependence_Token,err:=Get_Token_Struct_Dependence(struct_Definition.Fields_Token[field])
		if err!=nil {
			return out_Dependencies, err
		}
		if dependence_Token.Type!="pointer" && str_index_in_str_arr(dependence_Token.Value, out_Dependencies)==-1 && str_index_in_str_arr(dependence_Token.Value, Primitive_Types)==-1 {
			out_Dependencies = append(out_Dependencies, dependence_Token.Value)
		}
	}
	return out_Dependencies, nil
}

func Is_Struct_Declaration_Recursive(struct_name string, nested_Inside []string, struct_Dependencies map[string][]string) bool {
	for _,Struct:=range struct_Dependencies[struct_name] {
		if str_index_in_str_arr(Struct, nested_Inside)!=-1 {
			return true
		} else {
			return Is_Struct_Declaration_Recursive(Struct, append(nested_Inside, Struct), struct_Dependencies)
		}
	}
	return false
}

func Parser(code []Token, filePath string, imported_Programs map[string]Program) (Program, error) {
	filePath,err:=filepath.Abs(filePath)
	if err!=nil {
		return Program{}, err
	}
	is_Header:=strings.HasSuffix(filePath, ".vh")
	program:=Program{
		Structs: make(map[string]*Type),
		Functions: make([]Function, 0),
		Rendered_Scope: make([]*Object, 0),
		Is_Dynamic: is_Header,
		Dependencies: make([]string, 0),
	}
	definitions,err:=Definition_Parser(code, filePath)
	if err!=nil {
		return program, err
	}
	program.Package_Name=definitions.PackageName
	for _,Import_Declaration:=range definitions.Imports {
		file_Path:=Import_Declaration[0]
		file_Path,err=filepath.Abs(file_Path)
		program.Dependencies = append(program.Dependencies, file_Path)
		if err!=nil {
			return program, err
		}
		Alias:=Import_Declaration[1]
		Imported_Program,recycled_Import:=imported_Programs[file_Path]
		old_dir,err:=os.Getwd()
		if err!=nil {
			return program, err
		}
		if !recycled_Import {
			data,err:=os.ReadFile(file_Path)
			if err!=nil {
				return program, err
			}
			Imported_File,err:=Tokenizer(string(data))
			if err!=nil {
				return program, err
			}
			os.Chdir(filepath.Dir(file_Path))
			Imported_Program,err=Parser(Imported_File, file_Path, imported_Programs)
			program.Dependencies = append(program.Dependencies, Imported_Program.Dependencies...)
			if err!=nil {
				return program, err
			}
		}
		imported_Programs[file_Path]=Imported_Program
		Imported_Program.Package_Name=file_Path
		os.Chdir(old_dir)
		for Imported_Struct:=range Imported_Program.Structs {
			program.Structs[Alias+"."+Imported_Struct]=Imported_Program.Structs[Imported_Struct]
		}
		for _,Imported_Function:=range Imported_Program.Functions {
			copied_Imported_Function:=Imported_Function
			copied_Imported_Function.Name=Alias+"."+copied_Imported_Function.Name
			program.Functions = append(program.Functions, copied_Imported_Function)
		}
	}
	base_Function_Variable_Scope:=make(map[string]int)
	struct_Keys:=make([]string, 0)
	struct_Dependencies:=make(map[string][]string)
	for _,Struct:=range definitions.Structs {
		dependence,err:=Get_Struct_Definition_Refers(Struct)
		if err!=nil {
			return program, err
		}
		struct_Dependencies[Struct.Name]=dependence
		struct_Keys = append(struct_Keys, Struct.Name)
	}
	for _,Struct:=range struct_Keys {
		if Is_Struct_Declaration_Recursive(Struct, []string{Struct}, struct_Dependencies) {
			return program, errors.New("definition for struct '"+Struct+"' is recursive")
		}
		program.Structs[Struct]=&Type{Struct_Details: make(map[string]*Type)}
	}
	for _,Struct:=range definitions.Structs {
		struct_Type:=Type{Is_Struct: true, Struct_Details: make(map[string]*Type)}
		for field:=range Struct.Fields_Token {
			rendered_Type,err:=Type_Token_To_Struct(Struct.Fields_Token[field], &program)
			if err!=nil {
				return program, err
			}
			struct_Type.Struct_Details[field]=rendered_Type
		}
		*program.Structs[Struct.Name]=struct_Type
	}
	for _,variable_Definition:=range definitions.Variables {
		for _,variable_Name:=range variable_Definition.Names {
			variable_Type,err:=Type_Token_To_Struct(variable_Definition.Type_Token, &program)
			if err!=nil {
				return program, err
			}
			program.Object_References = append(program.Object_References, Object_Reference{Aliases: []string{variable_Name}, Object_Type: *variable_Type})
			program.Globally_Available = append(program.Globally_Available, len(program.Object_References)-1)
			program.Rendered_Scope = append(program.Rendered_Scope, &Object{Value: Default_Object_By_Type(*variable_Type)})
			base_Function_Variable_Scope[variable_Name]=len(program.Object_References)-1
		}
	}
	function_Count_Before_Processing:=len(program.Functions)
	for _,Function_Definition:=range definitions.Functions {
		copy_base_Function_Variable_Scope:=make(map[string]int)
		for variable_Name,index:=range base_Function_Variable_Scope {
			copy_base_Function_Variable_Scope[variable_Name]=index
		}
		function_Out_Type,err:=Type_Token_To_Struct(Function_Definition.Out_Token, &program)
		if err!=nil {
			return program, err
		}
		sample_Package_Function_Definition:=func(objects []*interface{})structs.Execution_Result{
			return structs.Execution_Result{}
		}
		function_Declaration:=Function{Name: Function_Definition.Name, Stack_Spec: make(map[int]Object_Abstract), Arguments: make(map[string]Type), Variable_Scope: copy_base_Function_Variable_Scope, Out_Type: *function_Out_Type, Base_Program: &program, Argument_Indexes: make([]int, 0), External_Function: &sample_Package_Function_Definition}
		for argument_Name, argument_Type_Token:=range Function_Definition.Arguments_Variables {
			argument_Type,err:=Type_Token_To_Struct(argument_Type_Token, &program)
			if err!=nil {
				return program, err
			}
			function_Declaration.Arguments[argument_Name]=*argument_Type
			function_Declaration.Argument_Names = append(function_Declaration.Argument_Names, argument_Name)
			argument_Reference:=Object_Reference{Aliases: []string{argument_Name}, Object_Type: *argument_Type}
			program.Rendered_Scope = append(program.Rendered_Scope, &Object{Value: Default_Object_By_Type(*argument_Type)})
			program.Object_References = append(program.Object_References, argument_Reference)
			function_Declaration.Stack_Spec[len(program.Object_References)-1]=Type_Struct_To_Object_Abstract(*argument_Type)
			function_Declaration.Variable_Scope[argument_Name]=len(program.Object_References)-1
			function_Declaration.Argument_Indexes = append(function_Declaration.Argument_Indexes, len(program.Rendered_Scope)-1)
		}
		program.Functions = append(program.Functions, function_Declaration)
	}
	if is_Header {
		return program, nil
	}
	for i,Function_Definition:=range definitions.Functions {
		function_Declaration:=program.Functions[i+function_Count_Before_Processing]
		err=Function_Parser(&Function_Definition, &function_Declaration, &program)
		program.Functions[i+function_Count_Before_Processing]=function_Declaration
		if err!=nil {
			return program, err
		}
	}
	fmt.Println(program)
	return program, nil
}

func Function_Parser(function_Definition *Function_Definition, function *Function, program *Program) error {
	code:=function_Definition.Instruction_Tokens
	global_Variables:=make([]string, 0)
	for i:=range program.Globally_Available {
		global_Variables = append(global_Variables, program.Object_References[i].Aliases[0])
	}
	local_Variables:=make([]string, 0)
	for local_Variable:=range function.Variable_Scope {
		local_Variables = append(local_Variables, local_Variable)
	}
	for i:=0; i<len(code); i++ {
		if code[i].Type=="sys" && code[i].Value=="var" {
			starting_length:=len(global_Variables)+len(local_Variables)
			if !(len(code)-i>=4) {
				return errors.New("invalid variable declaration statement")
			}
			variable_Type:=&Type{}
			j:=i
			variables_Added:=make([]string, 0)
			for {
				j++
				if j>=len(code) {
					return errors.New("unexpected EOF while parsing variable declaration statement")
				}
				if code[j].Type=="type" {
					err:=errors.New("")
					variable_Type,err=Type_Token_To_Struct(code[j], program)
					if err!=nil {
						return err
					}
					break
				}
				if code[j].Type!="variable" {
					return errors.New("expected token of type 'variable' during variable definition got '"+code[j].Type+"'")
				}
				if !Is_Valid_Variable_Name(code[j].Value) {
					return errors.New("variable name is invalid")
				}
				local_Index:=str_index_in_str_arr(code[j].Value, local_Variables)
				global_Index:=str_index_in_str_arr(code[j].Value, global_Variables)
				if global_Index!=-1 || local_Index!=-1 {
					return errors.New("Variable '"+code[j].Value+"' has already been initialized "+function.Name)
				}
				local_Variables = append(local_Variables, code[j].Value)
				variables_Added = append(variables_Added, code[j].Value)
			}
			if len(global_Variables)+len(local_Variables)-starting_length==0 {
				return errors.New("invalid variable definition structure")
			}
			for _,variable:=range variables_Added {
				program.Object_References = append(program.Object_References, Object_Reference{Aliases: []string{variable}, Object_Type: *variable_Type})
				program.Rendered_Scope = append(program.Rendered_Scope, &Object{Value: Default_Object_By_Type(*variable_Type)})
				function.Stack_Spec[len(program.Rendered_Scope)-1]=Type_Struct_To_Object_Abstract(*variable_Type)
				function.Variable_Scope[variable]=len(program.Rendered_Scope)-1
				function.Instructions = append(function.Instructions, []int{USE_DEFAULT_OBJECT_INSTRUCTION, len(program.Rendered_Scope)-1, 0})
			}
			i=j+1
			continue
		}
		if code[i].Type=="variable" && code[i].Value=="set" {
			if len(code)-i<4 {
				return errors.New("invalid set instruction definition structure 1")
			}
			if code[i+1].Type!="variable" {
				return errors.New("invalid set instruction definition structure 2")
			}
			variable_Index, found:=function.Variable_Scope[code[i+1].Value]
			if !found {
				return errors.New("variable '"+code[i+1].Value+"' not found")
			}
			if !Equal_Type(&program.Object_References[variable_Index].Object_Type, &Type{Raw_Type: INT_TYPE}) {
				return errors.New("expected type a variable integer while parsing set instruction")
			}
			if code[i+2].Type!="number" || code[i+3].Type!="semicolon" {
				return errors.New("invalid set instruction definition structure 3")
			}
			function.Instructions = append(function.Instructions, []int{SET_INSTRUCTION, variable_Index, int(code[i+2].Float64_Constant)})
			i+=3
			continue
		}
		if Instruction_Index:=str_index_in_str_arr(code[i].Value, []string{"add", "sub", "div", "mult", "pow", "floor", "mod", "greater", "smaller", "and", "or", "xor", "equals", "nequals"}); code[i].Type=="variable" && Instruction_Index!=-1 {
			Instruction:=[]int{ADD_INSTRUCTION, SUB_INSTRUCTION, DIV_INSTRUCTION, MULT_INSTRUCTION, POWER_INSTRUCTION, FLOOR_INSTRUCTION, MOD_INSTRUCTION, GREATER_INSTRUCTION, SMALLER_INSTRUCTION, AND_INSTRUCTION, OR_INSTRUCTION, XOR_INSTRUCTION, EQUALS_INSTRUCTION, NEQUALS_INSTRUCTION}[Instruction_Index]
			if len(code)-i<5 {
				return errors.New("invalid instruction definition structure")
			}
			if code[i+1].Type!="variable" {
				return errors.New("invalid instruction definition structure")
			}
			variable1_Index, found:=function.Variable_Scope[code[i+1].Value]
			if !found {
				return errors.New("variable '"+code[i+1].Value+"' not found")
			}
			if !Equal_Type(&program.Object_References[variable1_Index].Object_Type, &Type{Raw_Type: INT_TYPE}) {
				return errors.New("expected type a variable integer while parsing instruction")
			}
			if code[i+2].Type!="variable" {
				return errors.New("invalid instruction definition structure")
			}
			variable2_Index, found:=function.Variable_Scope[code[i+2].Value]
			if !found {
				return errors.New("variable '"+code[i+2].Value+"' not found")
			}
			if !Equal_Type(&program.Object_References[variable2_Index].Object_Type, &Type{Raw_Type: INT_TYPE}) {
				return errors.New("expected type a variable integer while parsing len instruction")
			}
			if code[i+3].Type!="variable" {
				return errors.New("invalid instruction definition structure")
			}
			variable3_Index, found:=function.Variable_Scope[code[i+3].Value]
			if !found {
				return errors.New("variable '"+code[i+3].Value+"' not found")
			}
			if !Equal_Type(&program.Object_References[variable3_Index].Object_Type, &Type{Raw_Type: INT_TYPE}) {
				return errors.New("expected type a variable integer while parsing instruction")
			}
			if code[i+4].Type!="semicolon" {
				return errors.New("invalid instruction definition structure")
			}
			function.Instructions = append(function.Instructions, []int{Instruction, variable1_Index, variable2_Index, variable3_Index})
			i+=4
			continue
		}
		if code[i].Type=="variable" && code[i].Value=="not" {
			if len(code)-i<3 {
				return errors.New("invalid instruction definition structure")
			}
			if code[i+1].Type!="variable" {
				return errors.New("invalid instruction definition structure")
			}
			variable_Index, found:=function.Variable_Scope[code[i+1].Value]
			if !found {
				return errors.New("variable '"+code[i+1].Value+"' not found")
			}
			if !Equal_Type(&program.Object_References[variable_Index].Object_Type, &Type{Raw_Type: INT_TYPE}) {
				return errors.New("expected type a variable integer while parsing instruction")
			}
			function.Instructions = append(function.Instructions, []int{NOT_INSTRUCTION, variable_Index})
			i+=2
			continue
		}
		if code[i].Type=="variable" && code[i].Value=="return" {
			if len(code)-i<3 {
				return errors.New("unexpected EOF while parsing return statement")
			}
			if code[i+1].Type!="variable" {
				return errors.New("invalid return instruction definition structure")
			}
			variable_Index, found:=function.Variable_Scope[code[i+1].Value]
			if !found {
				return errors.New("variable '"+code[i+1].Value+"' not found")
			}
			if !Equal_Type(&program.Object_References[variable_Index].Object_Type, &function.Out_Type) {
				return errors.New("return type of function does not match the type being returned")
			}
			if code[i+2].Type!="semicolon" {
				return errors.New("invalid return instruction definition structure")
			}
			function.Instructions = append(function.Instructions, []int{RETURN_INSTRUCTION, variable_Index})
			i+=2
			continue
		}
		if code[i].Type=="variable" && code[i].Value=="copy" {
			if len(code)-i<4 {
				return errors.New("unexpected EOF while parsing copy statement")
			}
			if code[i+1].Type!="variable" {
				return errors.New("invalid return instruction definition structure")
			}
			variableA_Index, found:=function.Variable_Scope[code[i+1].Value]
			if !found {
				return errors.New("variable '"+code[i+1].Value+"' not found")
			}
			if code[i+2].Type!="variable" {
				return errors.New("invalid copy instruction definition structure")
			}
			variableB_Index, found:=function.Variable_Scope[code[i+2].Value]
			if !found {
				return errors.New("variable '"+code[i+2].Value+"' not found")
			}
			if !Equal_Type(&program.Object_References[variableA_Index].Object_Type, &program.Object_References[variableB_Index].Object_Type) {
				return errors.New("return type of function does not match the type being returned")
			}
			if code[i+3].Type!="semicolon" {
				return errors.New("invalid return instruction definition structure")
			}
			function.Instructions = append(function.Instructions, []int{DEEP_COPY_OBJECT_INSTRUCTION, variableA_Index, variableB_Index})
			i+=3
			continue
		}
		if code[i].Type=="variable" && code[i].Value=="soft_copy" {
			if len(code)-i<4 {
				return errors.New("unexpected EOF while parsing copy statement")
			}
			if code[i+1].Type!="variable" {
				return errors.New("invalid return instruction definition structure")
			}
			variableA_Index, found:=function.Variable_Scope[code[i+1].Value]
			if !found {
				return errors.New("variable '"+code[i+1].Value+"' not found")
			}
			if code[i+2].Type!="variable" {
				return errors.New("invalid copy instruction definition structure")
			}
			variableB_Index, found:=function.Variable_Scope[code[i+2].Value]
			if !found {
				return errors.New("variable '"+code[i+2].Value+"' not found")
			}
			if !Equal_Type(&program.Object_References[variableA_Index].Object_Type, &program.Object_References[variableB_Index].Object_Type) {
				return errors.New("return type of function does not match the type being returned")
			}
			if code[i+3].Type!="semicolon" {
				return errors.New("invalid return instruction definition structure")
			}
			function.Instructions = append(function.Instructions, []int{SOFT_COPY_OBJECT_INSTRUCTION, variableA_Index, variableB_Index})
			i+=3
			continue
		}
		if code[i].Type=="variable" && code[i].Value=="append" {
			if len(code)-i<5 {
				return errors.New("invalid instruction definition structure")
			}
			if code[i+1].Type!="variable" {
				return errors.New("invalid instruction definition structure")
			}
			variable1_Index, found:=function.Variable_Scope[code[i+1].Value]
			if !found {
				return errors.New("variable '"+code[i+1].Value+"' not found")
			}
			if !program.Object_References[variable1_Index].Object_Type.Is_Array {
				return errors.New("expected type a array while parsing instruction")
			}
			if code[i+2].Type!="variable" {
				return errors.New("invalid instruction definition structure")
			}
			variable2_Index, found:=function.Variable_Scope[code[i+2].Value]
			if !found {
				return errors.New("variable '"+code[i+2].Value+"' not found")
			}
			if !Equal_Type(&program.Object_References[variable2_Index].Object_Type, program.Object_References[variable1_Index].Object_Type.Child) {
				return errors.New("expected type a array.child while parsing append instruction")
			}
			if code[i+3].Type!="variable" {
				return errors.New("invalid instruction definition structure")
			}
			variable3_Index, found:=function.Variable_Scope[code[i+3].Value]
			if !found {
				return errors.New("variable '"+code[i+3].Value+"' not found")
			}
			if !Equal_Type(&program.Object_References[variable3_Index].Object_Type, &program.Object_References[variable1_Index].Object_Type) {
				return errors.New("expected type a array while parsing instruction")
			}
			if code[i+4].Type!="semicolon" {
				return errors.New("invalid instruction definition structure")
			}
			function.Instructions = append(function.Instructions, []int{APPEND_INSTRUCTION, variable1_Index, variable2_Index, variable3_Index})
			i+=4
			continue
		}
		if code[i].Type=="variable" && code[i].Value=="array_lookup" {
			if len(code)-i<5 {
				return errors.New("invalid instruction definition structure")
			}
			if code[i+1].Type!="variable" {
				return errors.New("invalid instruction definition structure")
			}
			variable1_Index, found:=function.Variable_Scope[code[i+1].Value]
			if !found {
				return errors.New("variable '"+code[i+1].Value+"' not found")
			}
			if !program.Object_References[variable1_Index].Object_Type.Is_Array {
				return errors.New("expected type a array while parsing instruction")
			}
			if code[i+2].Type!="variable" {
				return errors.New("invalid instruction definition structure")
			}
			variable2_Index, found:=function.Variable_Scope[code[i+2].Value]
			if !found {
				return errors.New("variable '"+code[i+2].Value+"' not found")
			}
			if !Equal_Type(&program.Object_References[variable2_Index].Object_Type, &Type{Raw_Type: INT_TYPE}) {
				return errors.New("expected type to be an integer while parsing array_lookup instruction")
			}
			if code[i+3].Type!="variable" {
				return errors.New("invalid instruction definition structure")
			}
			variable3_Index, found:=function.Variable_Scope[code[i+3].Value]
			if !found {
				return errors.New("variable '"+code[i+3].Value+"' not found")
			}
			if !Equal_Type(&program.Object_References[variable3_Index].Object_Type, program.Object_References[variable1_Index].Object_Type.Child) {
				return errors.New("expected type a array.child while parsing instruction")
			}
			if code[i+4].Type!="semicolon" {
				return errors.New("invalid instruction definition structure")
			}
			function.Instructions = append(function.Instructions, []int{ARRAY_TYPE_LOOKUP_INSTRUCTION, variable1_Index, variable2_Index, variable3_Index})
			i+=4
			continue
		}
		if code[i].Type=="variable" && code[i].Value=="dict_lookup" {
			if len(code)-i<5 {
				return errors.New("invalid instruction definition structure")
			}
			if code[i+1].Type!="variable" {
				return errors.New("invalid instruction definition structure")
			}
			variable1_Index, found:=function.Variable_Scope[code[i+1].Value]
			if !found {
				return errors.New("variable '"+code[i+1].Value+"' not found")
			}
			if !program.Object_References[variable1_Index].Object_Type.Is_Dict {
				return errors.New("expected type a dict while parsing instruction")
			}
			if code[i+2].Type!="variable" {
				return errors.New("invalid instruction definition structure")
			}
			variable2_Index, found:=function.Variable_Scope[code[i+2].Value]
			if !found {
				return errors.New("variable '"+code[i+2].Value+"' not found")
			}
			if !(program.Object_References[variable2_Index].Object_Type.Raw_Type==program.Object_References[variable1_Index].Object_Type.Raw_Type) {
				return errors.New("unexpected type of lookup key while parsing dict_lookup instruction")
			}
			if code[i+3].Type!="variable" {
				return errors.New("invalid instruction definition structure")
			}
			variable3_Index, found:=function.Variable_Scope[code[i+3].Value]
			if !found {
				return errors.New("variable '"+code[i+3].Value+"' not found")
			}
			if !Equal_Type(&program.Object_References[variable3_Index].Object_Type, program.Object_References[variable1_Index].Object_Type.Child) {
				return errors.New("expected type a array.child while parsing instruction")
			}
			if code[i+4].Type!="semicolon" {
				return errors.New("invalid instruction definition structure")
			}
			function.Instructions = append(function.Instructions, []int{DICT_TYPE_LOOKUP_INSTRUCTION, variable1_Index, variable2_Index, variable3_Index, int(program.Object_References[variable1_Index].Object_Type.Raw_Type)})
			i+=4
			continue
		}
		if code[i].Type=="variable" && code[i].Value=="len" {
			if len(code)-i<4 {
				return errors.New("invalid instruction definition structure")
			}
			if code[i+1].Type!="variable" {
				return errors.New("invalid instruction definition structure")
			}
			variable1_Index, found:=function.Variable_Scope[code[i+1].Value]
			if !found {
				return errors.New("variable '"+code[i+1].Value+"' not found")
			}
			if !program.Object_References[variable1_Index].Object_Type.Is_Array {
				return errors.New("expected type a array while parsing instruction")
			}
			if code[i+2].Type!="variable" {
				return errors.New("invalid instruction definition structure")
			}
			variable2_Index, found:=function.Variable_Scope[code[i+2].Value]
			if !found {
				return errors.New("variable '"+code[i+2].Value+"' not found")
			}
			if !Equal_Type(&program.Object_References[variable2_Index].Object_Type, &Type{Raw_Type: INT_TYPE}) {
				return errors.New("expected type an integer variable while parsing instruction")
			}
			if code[i+3].Type!="semicolon" {
				return errors.New("invalid instruction definition structure")
			}
			function.Instructions = append(function.Instructions, []int{LEN_INSTRUCTION, variable1_Index, variable2_Index})
			i+=3
			continue
		}
		if code[i].Type=="variable" && code[i].Value=="jump" {
			if len(code)-i<4 {
				return errors.New("unexpected EOF while parsing copy statement")
			}
			if code[i+1].Type!="variable" {
				return errors.New("invalid jump instruction definition structure")
			}
			variableA_Index, found:=function.Variable_Scope[code[i+1].Value]
			if !found {
				return errors.New("variable '"+code[i+1].Value+"' not found")
			}
			if code[i+2].Type!="variable" {
				return errors.New("invalid jump instruction definition structure")
			}
			variableB_Index, found:=function.Variable_Scope[code[i+2].Value]
			if !found {
				return errors.New("variable '"+code[i+2].Value+"' not found")
			}
			if !(Equal_Type(&program.Object_References[variableA_Index].Object_Type, &program.Object_References[variableB_Index].Object_Type) && Equal_Type(&program.Object_References[variableA_Index].Object_Type, &Type{Raw_Type: INT_TYPE})) {
				return errors.New("jump instruction has invalid variable types")
			}
			if code[i+3].Type!="semicolon" {
				return errors.New("invalid jump instruction definition structure")
			}
			function.Instructions = append(function.Instructions, []int{JUMP_INSTRUCTION, variableA_Index, variableB_Index})
			i+=3
			continue
		}
		if code[i].Type=="variable" && code[i].Value=="jumpto" {
			if len(code)-i<3 {
				return errors.New("unexpected EOF while parsing copy statement")
			}
			if code[i+1].Type!="variable" {
				return errors.New("invalid jumpto instruction definition structure")
			}
			variable_Index, found:=function.Variable_Scope[code[i+1].Value]
			if !found {
				return errors.New("variable '"+code[i+1].Value+"' not found")
			}
			if !(Equal_Type(&program.Object_References[variable_Index].Object_Type, &Type{Raw_Type: INT_TYPE})) {
				return errors.New("jumpto instruction has invalid variable types")
			}
			if code[i+2].Type!="semicolon" {
				return errors.New("invalid jumpto instruction definition structure")
			}
			function.Instructions = append(function.Instructions, []int{JUMPTO_INSTRUCTION, variable_Index})
			i+=2
			continue
		}
		if code[i].Type=="variable" && code[i].Value=="call" {
			if code[i+1].Type!="variable" {
				return errors.New("invalid call instruction definition")
			}
			function_Index:=-1
			for index,function:=range program.Functions {
				if function.Name==code[i+1].Value {
					function_Index=index
				}
			}
			if function_Index==-1 {
				return errors.New("function '"+code[i+1].Value+"' was not found")
			}
			if code[i+2].Type!="bracket" || code[i+2].Value!="(" {
				return errors.New("invalid call instruction definition")
			}
			j:=i+2
			variable_Indexes:=make([]int, 0)
			last_Comma:=true
			for {
				j++
				if j>=len(code) {
					return errors.New("unexpected EOF during parsing of call instruction")
				}
				if code[j].Type=="bracket" && code[j].Value==")" {
					break
				}
				if code[j].Type=="comma" {
					if last_Comma {
						return errors.New("invalid call instruction definition")
					}
					last_Comma=true
					continue
				}
				if code[j].Type=="variable" {
					variable_Index,found:=function.Variable_Scope[code[j].Value]
					if !found {
						return errors.New("variable '"+code[j].Value+"' was not found")
					}
					variable_Indexes = append(variable_Indexes, variable_Index)
					last_Comma=false
				} else {
					return errors.New("invalid call instruction definition")
				}
			}
			if len(variable_Indexes)!=len(program.Functions[function_Index].Arguments) {
				return errors.New("function call arguments number do not match")
			}
			for index,argument_Name:=range program.Functions[function_Index].Argument_Names {
				argument_Type:=program.Functions[function_Index].Arguments[argument_Name]
				if !Equal_Type(&program.Object_References[variable_Indexes[index]].Object_Type, &argument_Type) {
					return errors.New("function argument types do not match")
				}
			}
			if j+2>=len(code) {
				return errors.New("unexpected EOF during parsing of call instruction")
			}
			if code[j+1].Type!="variable" || code[j+2].Type!="semicolon" {
				return errors.New("invalid call instruction definition")
			}
			variable_Index,found:=function.Variable_Scope[code[j+1].Value]
			if !found {
				return errors.New("variable '"+code[j+1].Value+"' was not found")
			}
			if !Equal_Type(&program.Functions[function_Index].Out_Type, &program.Object_References[variable_Index].Object_Type) {
				return errors.New("function call return variable's type does not match the return type of the function being called")
			}
			instructions:=[]int{CALL_INSTRUCTION, function_Index, variable_Index}
			instructions = append(instructions, variable_Indexes...)
			function.Instructions = append(function.Instructions, instructions)
			i=j+2
			continue
		}
		return errors.New("unrecognised instruction token of type '"+code[i].Type+"'")
	}
	return nil
}