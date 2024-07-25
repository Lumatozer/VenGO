package venc

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func Definition_Parser(code []Token) (Definitions, error) {
	definitions:=Definitions{Imports: make(map[string]string), Variables: make(map[string]Token), Functions: make([]Function_Definition, 0), Structs: make(map[string]map[string]Token)}
	if len(code)<2 || code[0].Type!="sys" || code[0].Value!="package" || code[1].Type!="variable" || !Is_Valid_Var_Name(code[1].Value) {
		return definitions, errors.New("a valid package name is required")
	}
	definitions.Package_Name=code[1].Value
	for i:=2; i<len(code); i++ {
		if code[i].Type=="sys" && code[i].Value=="struct" {
			if !(len(code)-1>4) {
				return definitions, errors.New("struct definition is incomplete")
			}
			if code[i+1].Type!="variable" || code[i+2].Type!="bracket_open" || code[i+2].Value!="{" {
				return definitions, errors.New("invalid struct declaration during file parsing")
			}
			if !Is_Valid_Var_Name(code[i+1].Value) {
				return definitions, errors.New("invalid struct name '"+code[i+1].Value+"'")
			}
			for Struct_Name:=range definitions.Structs {
				if Struct_Name==code[i+1].Value {
					return definitions, errors.New("struct '"+code[i+1].Value+"' has already been defined")
				}
			}
			definitions.Structs[code[i+1].Value]=make(map[string]Token)
			j:=i+2
			for {
				j++
				if j>=len(code) {
					return definitions, errors.New("unexpected EOF during struct parsing 1")
				}
				if code[j].Type=="bracket_close" && code[j].Value=="}" {
					break
				}
				if code[j].Type!="variable" {
					return definitions, errors.New("invalid struct declaration during file parsing")
				}
				if !Is_Valid_Var_Name(code[j].Value) {
					return definitions, errors.New("invalid struct field name '"+code[i+1].Value+"'")
				}
				field_Name:=code[j].Value
				_,ok:=definitions.Structs[code[i+1].Value][field_Name]
				if ok {
					fmt.Println("field '"+field_Name+"' has already been defined")
				}
				j++
				if j+1>=len(code) {
					return definitions, errors.New("unexpected EOF during struct parsing")
				}
				if code[j].Type!="type" {
					return definitions, errors.New("invalid struct declaration during file parsing")
				}
				definitions.Structs[code[i+1].Value][field_Name]=code[j]
			}
			i=j
			continue
		}
		if code[i].Type=="sys" && code[i].Value=="function" {
			if !(len(code)-i>=2) || code[i+1].Type!="funcall" {
				return definitions, errors.New("invalid function declaration during file parsing")
			}
			if !Is_Valid_Var_Name(code[i+1].Children[0].Value) {
				return definitions, errors.New("function name \""+code[i+1].Children[0].Value+"\" is invalid")
			}
			ArgumentTokens:=code[i+1].Children[1].Children
			if len(ArgumentTokens)!=0 && len(ArgumentTokens)!=2 && (len(ArgumentTokens))%2!=1 {
				return definitions, errors.New("Improper declaration of function arguments inside function '"+code[i+1].Children[0].Value+"'")
			}
			FunctionArguments:=make(map[string]Token)
			if len(ArgumentTokens)!=0 {
				for j:=0; j<len(ArgumentTokens); j+=3 {
					if ArgumentTokens[j].Type!="variable" {
						return definitions, errors.New("Improper declaration of function arguments inside function '"+code[i+1].Children[0].Value+"'")
					}
					_,ArgumentFound:=FunctionArguments[ArgumentTokens[j].Value]
					if ArgumentFound {
						return definitions, errors.New("Improper declaration of function arguments inside function '"+code[i+1].Children[0].Value+"'")
					}
					if ArgumentTokens[j+1].Type!="type" {
						return definitions, errors.New("Improper declaration of function arguments inside function '"+code[i+1].Children[0].Value+"'")
					}
					FunctionArguments[ArgumentTokens[j].Value]=ArgumentTokens[j+1]
					if j+2<len(ArgumentTokens) {
						if ArgumentTokens[j+2].Type!="comma" {
							return definitions, errors.New("Improper declaration of function arguments inside function '"+code[i+1].Children[0].Value+"'")
						}
					}
				}
			}
			function_Name:=code[i+1].Children[0].Value
			for _,function:=range definitions.Functions {
				if function.Name==function_Name {
					return definitions, errors.New("function \""+function_Name+"\" has already been declared")
				}
			}
			if !(len(code)-(i+1)>1) || (code[i+2].Type!="type" && code[i+2].Type!="bracket_open" && code[i+2].Value!="{") {
				definitions.Functions = append(definitions.Functions, Function_Definition{Name: function_Name, Out_Type: Token{Type: "type", Children: []Token{Token{Type: "raw", Value: "void"}}}, Arguments: FunctionArguments})
				i+=1
				continue
			}
			function_TypeToken:=Token{}
			if code[i+2].Type=="type" {
				function_TypeToken=code[i+2]
				i+=1
			} else {
				function_TypeToken=Token{Type: "type", Children: []Token{Token{Type: "raw", Value: "void"}}}
			}
			if !(len(code)-(i+2)>1) || (code[i+2].Type!="bracket_open" && code[i+2].Value!="{") {
				definitions.Functions = append(definitions.Functions, Function_Definition{Name: function_Name, Out_Type: function_TypeToken, Arguments: FunctionArguments})
				i+=1
				continue
			}
			function_Tokens:=make([]Token, 0)
			j:=i+2
			count:=1
			for {
				j++
				if j>=len(code) {
					return definitions, errors.New("unexpected EOF while parsing function '"+function_Name+"'")
				}
				if code[j].Type=="bracket_open" && code[j].Value=="{" {
					count+=1
				}
				if code[j].Type=="bracket_close" && code[j].Value=="}" {
					count-=1
					if count==0 {
						break
					}
				}
				function_Tokens = append(function_Tokens, code[j])
			}
			definitions.Functions = append(definitions.Functions, Function_Definition{Name: function_Name, Out_Type: function_TypeToken, Arguments: FunctionArguments, Internal_Tokens: function_Tokens})
			i=j
			continue
		}
		if code[i].Type=="sys" && code[i].Value=="var" {
			if !(len(code)-i>=3) {
				return definitions, errors.New("invalid variable declartion during file parsing")
			}
			variables:=make([]string, 0)
			j:=i
			variablesTypeToken:=Token{}
			for {
				j++
				if j>=len(code) {
					return definitions, errors.New("unexpected EOF during variable parsing")
				}
				if code[j].Type=="type" {
					if (len(code)-j>1) && code[j+1].Type=="EOS" {
						variablesTypeToken=code[j]
						break
					} else {
						return definitions, errors.New("type should be the last token during variable declarations proceeded by an EOS")
					}
				}
				if code[j].Type!="variable" || !Is_Valid_Var_Name(code[j].Value) {
					return definitions, errors.New("invalid variable declaration statement")
				}
				if str_index_in_arr(code[j].Value, variables)!=-1 {
					return definitions, errors.New("variable '"+code[j].Value+"' has already been declared")
				}
				for variable:=range definitions.Variables {
					if variable==code[j].Value {
						return definitions, errors.New("variable '"+variable+"' has already been declared")
					}
				}
				variables = append(variables, code[j].Value)
			}
			i=j+1
			for _,variable:=range variables {
				definitions.Variables[variable]=variablesTypeToken
			}
			continue
		}
		if code[i].Type=="sys" && code[i].Value=="import" {
			if code[i+1].Type!="expression" {
				fmt.Println(code[i+1])
				return definitions, errors.New("invalid import statement declaration")
			}
			importTokens:=code[i+1].Children
			if len(importTokens)%3!=0 {
				return definitions, errors.New("invalid import statement declaration")
			}
			packagesImported:=make([]string, 0)
			for j:=0; len(importTokens)>j; j+=3 {
				if importTokens[j].Type!="string" || importTokens[j+1].Type!="sys" || importTokens[j+1].Value!="as" || importTokens[j+2].Type!="variable" || !Is_Valid_Var_Name(importTokens[j+2].Value) {
					return definitions, errors.New("import statement declaration is invalid")
				}
				_,AlreadyImported:=definitions.Imports[importTokens[j].Value]
				if AlreadyImported {
					return definitions, errors.New("file '"+importTokens[j].Value+"' has already been imported")
				}
				if str_index_in_arr(importTokens[j+2].Value, packagesImported)!=-1 {
					return definitions, errors.New("module '"+importTokens[j+2].Value+"' has already been imported")
				}
				definitions.Imports[importTokens[j].Value]=importTokens[j+2].Value
				packagesImported = append(packagesImported, importTokens[j+2].Value)
			}
			i+=1
			continue
		}
		return definitions, errors.New("unexpected token of type '"+code[i].Type+"'")
	}
	return definitions, nil
}

func Parse_Type(type_token Token, program *Program) (*Type, error) {
	if type_token.Type=="type" {
		return Parse_Type(type_token.Children[0], program)
	}
	if type_token.Type=="raw" {
		Raw_Value,ok:=TYPE_MAP[type_token.Value]
		if ok {
			return &Type{Is_Raw: true, Raw_Type: Raw_Value}, nil
		}
		Struct,ok:=program.Structs[type_token.Value]
		if ok {
			return Struct, nil
		}
		return &Type{}, errors.New("type "+"'"+type_token.Value+"' was not found during type parsing")
	}
	if type_token.Type=="array" {
		Array_Type,err:=Parse_Type(type_token.Children[0], program)
		if err!=nil {
			return &Type{}, err
		}
		return &Type{Is_Array: true, Child: Array_Type}, nil
	}
	if type_token.Type=="dict" {
		Key_Type:=TYPE_MAP[type_token.Children[0].Value]
		Value_Type,err:=Parse_Type(type_token.Children[1], program)
		if err!=nil {
			return &Type{}, err
		}
		return &Type{Is_Dict: true, Raw_Type: Key_Type, Child: Value_Type}, nil
	}
	if type_token.Type=="pointer" {
		Child_Type,err:=Parse_Type(type_token.Children[0], program)
		if err!=nil {
			return &Type{}, err
		}
		return &Type{Is_Pointer: true, Child: Child_Type}, nil
	}
	return &Type{}, errors.New("invalid type")
}

func Does_Struct_Depend_On(Struct_A string, Struct_B string, Dependency_Map map[string][]string) bool {
	for _,Dependency:=range Dependency_Map[Struct_A] {
		if Dependency==Struct_B || Does_Struct_Depend_On(Dependency, Struct_B, Dependency_Map) {
			return true
		}
	}
	return false
}

func Struct_Dependencies(Struct_Fields map[string]Token, program *Program) []string {
	Dependencies:=make([]string, 0)
	for _,Field_Token:=range Struct_Fields {
		if Field_Token.Children[0].Type=="raw" && !strings.Contains(Field_Token.Children[0].Value, ".") {
			_,ok:=program.Structs[Field_Token.Children[0].Value]
			if ok {
				Dependencies = append(Dependencies, Field_Token.Children[0].Value)
			}
		}
	}
	return Dependencies
}

func Parser(path string, definitions Definitions) (Program, error) {
	program:=Program{Path: path, Package_Name: definitions.Package_Name, Vitality: true, Structs: make(map[string]*Type), Global_Variables: make(map[string]*Type), Functions: make(map[string]Function), Imported_Libraries: make(map[string]*Program)}
	Dependencies:=make(map[string][]string)
	for Import_Path, Import_Alias:=range definitions.Imports {
		data,err:=os.ReadFile(Import_Path)
		if err!=nil {
			return program, err
		}
		tokens:=Tokensier(string(data), false)
		tokens,err=Tokens_Parser(tokens, false)
		if err!=nil {
			return program, err
		}
		tokens,err=Token_Grouper(tokens, false)
		if err!=nil {
			return program, err
		}
		imported_Definition,err:=Definition_Parser(tokens)
		if err!=nil {
			return program, err
		}
		imported_Program,err:=Parser(Import_Path, imported_Definition)
		if err!=nil {
			return program, err
		}
		for Imported_Struct:=range imported_Program.Structs {
			program.Structs[Import_Alias+"."+Imported_Struct]=imported_Program.Structs[Imported_Struct]
		}
		for Imported_Function:=range imported_Program.Functions {
			program.Functions[Import_Alias+"."+Imported_Function]=imported_Program.Functions[Imported_Function]
		}
		program.Imported_Libraries[Import_Alias]=&imported_Program
	}
	for Struct_Name:=range definitions.Structs {
		program.Structs[Struct_Name]=&Type{}
		Dependencies[Struct_Name]=Struct_Dependencies(definitions.Structs[Struct_Name], &program)
	}
	for Struct_Name:=range Dependencies {
		for _,Struct_B:=range Dependencies[Struct_Name] {
			if Does_Struct_Depend_On(Struct_B, Struct_Name, Dependencies) {
				return program, errors.New("struct '"+Struct_Name+"' declaration is recursive")
			}
		}
	}
	for Struct_Name:=range definitions.Structs {
		Struct:=map[string]*Type{}
		for Field, Field_Type:=range definitions.Structs[Struct_Name] {
			Struct_Type,err:=Parse_Type(Field_Type, &program)
			if err!=nil {
				return program, err
			}
			Struct[Field]=Struct_Type
		}
		*program.Structs[Struct_Name]=Type{Is_Struct: true, Struct_Details: Struct}
	}
	for Variable_Name, Var_Type:=range definitions.Variables {
		Variable_Type,err:=Parse_Type(Var_Type, &program)
		if err!=nil {
			return program, err
		}
		program.Global_Variables[Variable_Name]=Variable_Type
	}
	return program, nil
}