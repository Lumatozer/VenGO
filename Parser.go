package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"os"
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
	Is_Raw                 bool
	Raw_Type               int8
	Is_Struct              bool
	Is_Pointer             bool
	Struct_Details         map[string]*Type
	Child                  *Type
}

type Program struct {
	Functions              []Function
	Structs                map[string]*Type
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
	imported_Aliases:=make([]string, 0)
	imported_Files:=make([]string, 0)
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
					relative_Path,err:=filepath.Abs(code[j].Value)
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

func Parser(code []Token) (Program, error) {
	program:=Program{
		Structs: make(map[string]*Type),
	}
	definitions,err:=Definition_Parser(code)
	if err!=nil {
		return program, err
	}
	for _,Import_Declaration:=range definitions.Imports {
		file_Path:=Import_Declaration[0]
		Alias:=Import_Declaration[1]
		data,err:=os.ReadFile(file_Path)
		if err!=nil {
			return program, err
		}
		Imported_File,err:=Tokenizer(string(data))
		if err!=nil {
			return program, err
		}
		Imported_Program,err:=Parser(Imported_File)
		if err!=nil {
			return program, err
		}
		for Imported_Struct:=range Imported_Program.Structs {
			program.Structs[Alias+"."+Imported_Struct]=Imported_Program.Structs[Imported_Struct]
		}
		for _,Imported_Function:=range Imported_Program.Functions {
			copied_Imported_Function:=Imported_Function
			copied_Imported_Function.Name=Alias+"."+copied_Imported_Function.Name
			program.Functions = append(program.Functions, copied_Imported_Function)
			for argument_Name, argument_type:=range copied_Imported_Function.Arguments {
				program.Object_References = append(program.Object_References, Object_Reference{Aliases: []string{Alias+"."+Imported_Function.Name+"."+argument_Name}, Object_Type: argument_type})
				program.Globally_Available = append(program.Globally_Available, len(program.Object_References)-1)
				program.Rendered_Scope = append(program.Rendered_Scope, &Object{})
			}
			program.Object_References = append(program.Object_References, Object_Reference{Aliases: []string{Alias+"."+Imported_Function.Name+"."+"return"}, Object_Type: Imported_Function.Out_Type})
			program.Globally_Available = append(program.Globally_Available, len(program.Object_References)-1)
			program.Rendered_Scope = append(program.Rendered_Scope, &Object{})
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
			program.Rendered_Scope = append(program.Rendered_Scope, &Object{})
			base_Function_Variable_Scope[variable_Name]=len(program.Object_References)-1
		}
	}
	for _,Function_Definition:=range definitions.Functions {
		copy_base_Function_Variable_Scope:=base_Function_Variable_Scope
		function_Out_Type,err:=Type_Token_To_Struct(Function_Definition.Out_Token, &program)
		if err!=nil {
			return program, err
		}
		function_Declaration:=Function{Name: Function_Definition.Name, Stack_Spec: make(map[int]Object_Abstract), Arguments: make(map[string]Type), Variable_Scope: copy_base_Function_Variable_Scope, Out_Type: *function_Out_Type}
		for argument_Name, argument_Type_Token:=range Function_Definition.Arguments_Variables {
			argument_Type,err:=Type_Token_To_Struct(argument_Type_Token, &program)
			if err!=nil {
				return program, err
			}
			function_Declaration.Arguments[argument_Name]=*argument_Type
			argument_Reference:=Object_Reference{Aliases: []string{argument_Name, function_Declaration.Name+"."+argument_Name}, Object_Type: *argument_Type}
			program.Object_References = append(program.Object_References, argument_Reference)
			function_Declaration.Stack_Spec[len(program.Object_References)-1]=Type_Struct_To_Object_Abstract(*argument_Type)
			function_Declaration.Variable_Scope[argument_Name]=len(program.Object_References)-1
			program.Rendered_Scope = append(program.Rendered_Scope, &Object{})
		}
		program.Object_References = append(program.Object_References, Object_Reference{Aliases: []string{function_Declaration.Name+"."+"result"}, Object_Type: function_Declaration.Out_Type})
		program.Globally_Available = append(program.Globally_Available, len(program.Object_References)-1)
		program.Rendered_Scope = append(program.Rendered_Scope, &Object{})
		program.Functions = append(program.Functions, function_Declaration)
	}
	fmt.Println(definitions)
	fmt.Println(program)
	return program, nil
}