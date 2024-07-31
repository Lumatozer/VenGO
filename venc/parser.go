package venc

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
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
			if !(len(code)-i>=2) || !(code[i+1].Type=="expression") {
				return definitions, errors.New("invalid import statement during file parsing")
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
	program:=Program{Path: path, Package_Name: definitions.Package_Name, Vitality: true, Structs: make(map[string]*Type), Global_Variables: make(map[string]*Type), Functions: make(map[string]*Function), Imported_Libraries: make(map[string]*Program)}
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
	for _,Function_Definition:=range definitions.Functions {
		Out_Type,err:=Parse_Type(Function_Definition.Out_Type, &program)
		if err!=nil {
			return program, err
		}
		defined_Function:=Function{Out_Type: Out_Type, Arguments: make([]struct{Name string; Type *Type}, 0), Scope: make(map[string]*Type), Instructions: make([][]string, 0)}
		for Variable_Name, Variable_Type:=range program.Global_Variables {
			defined_Function.Scope[Variable_Name]=Variable_Type
		}
		program.Functions[Function_Definition.Name]=&defined_Function
		for Argument, Argument_Token:=range Function_Definition.Arguments {
			Argument_Type,err:=Parse_Type(Argument_Token, &program)
			if err!=nil {
				return program, err
			}
			program.Functions[Function_Definition.Name].Arguments = append(program.Functions[Function_Definition.Name].Arguments, struct{Name string; Type *Type}{Name: Argument, Type: Argument_Type})
			defined_Function.Scope[Argument]=Argument_Type
		}
		temp_Variables:=Temp_Variables{Signature_Lookup: make(map[string]int), Variable_Lookup: make(map[int][]struct{Free bool; Allocated bool})}
		err=Function_Parser(Function_Definition.Internal_Tokens, Function_Definition, &defined_Function, &program, temp_Variables, struct{In_Loop bool; Loop_Details Loop_Details}{})
		if err!=nil {
			return program, err
		}
	}
	return program, nil
}

func Is_Expression_Valid(code []Token) bool {
	if len(code)%2==0 {
		return false
	}
	for i:=0; len(code)>i; i+=2 {
		if code[i].Type=="expression" {
			if !Is_Expression_Valid(code[i].Children) {
				return false
			}
		}
		if code[i].Type=="operator" {
			return false
		}
		if len(code)>i+1 && code[i+1].Type!="operator" {
			return false
		}
	}
	return true
}

func int_index_in_array(a int, arr []int) int {
	for i:=0; len(arr)>i; i++ {
		if arr[i]==a {
			return a
		}
	}
	return -1
}

func Generate_Unique_Temporary_Variable(variable_Type *Type, temp_Variables Temp_Variables, function *Function) string {
	signature:=Type_Signature(variable_Type, make([]*Type, 0))
	Signature_Id,ok:=temp_Variables.Signature_Lookup[signature]
	if !ok {
		temp_Variables.Signature_Lookup[signature]=len(temp_Variables.Signature_Lookup)
		temp_Variables.Variable_Lookup[len(temp_Variables.Signature_Lookup)-1]=make([]struct{Free bool; Allocated bool}, 0)
		Signature_Id=len(temp_Variables.Signature_Lookup)-1
	}
	Signature_Struct:=temp_Variables.Variable_Lookup[Signature_Id]
	for i:=0; len(Signature_Struct)>i; i++ {
		if Signature_Struct[i].Free {
			temp_Variables.Variable_Lookup[Signature_Id][i].Free=false
			variable_Name:="temp."+strconv.FormatInt(int64(Signature_Id), 10)+"_"+strconv.FormatInt(int64(i), 10)
			function.Scope[variable_Name]=variable_Type
			return variable_Name
		}
	}
	temp_Variables.Variable_Lookup[Signature_Id] = append(temp_Variables.Variable_Lookup[Signature_Id], struct{Free bool; Allocated bool}{Free: false, Allocated: false})
	variable_Name:="temp."+strconv.FormatInt(int64(Signature_Id), 10)+"_"+strconv.FormatInt(int64(len(temp_Variables.Variable_Lookup[Signature_Id])-1), 10)
	function.Scope[variable_Name]=variable_Type
	return variable_Name
}

func Free_Temporary_Unique_Variable(variable_Name string, temp_Variables Temp_Variables, function *Function) {
	delete(function.Scope, variable_Name)
	Temp_Id,_:=strconv.ParseInt(strings.Split(strings.Split(variable_Name, "_")[0], ".")[1], 10, 64)
	Temp_Int:=int(Temp_Id)
	Used_Id,_:=strconv.ParseInt(strings.Split(variable_Name, "_")[1], 10, 64)
	Int_Used:=int(Used_Id)
	temp_Variables.Variable_Lookup[Temp_Int][Int_Used].Free=true
}

func Evaluate_Type(code []Token, function *Function, program *Program) (*Type, error) {
	if len(code)==1 {
		if code[0].Type=="num" {
			if math.Round(code[0].Num_Value)==code[0].Num_Value {
				return &Type{Is_Raw: true, Raw_Type: INT_TYPE}, nil
			} else {
				return &Type{Is_Raw: true, Raw_Type: FLOAT64_TYPE}, nil
			}
		}
		if code[0].Type=="string" {
			return &Type{Is_Raw: true, Raw_Type: STRING_TYPE}, nil
		}
		if code[0].Type=="variable" {
			Var_Type,ok:=function.Scope[code[0].Value]
			if !ok {
				return &Type{}, errors.New("variable "+"'"+code[0].Value+"'"+" not in scope of expression")
			}
			return Var_Type, nil
		}
		if code[0].Type=="expression" {
			return Evaluate_Type(code[0].Children, function, program)
		}
		if code[0].Type=="funcall" {
			for Fn_Name, Fn:=range program.Functions {
				if Fn_Name==code[0].Children[0].Value {
					return Fn.Out_Type, nil
				}
			}
		}
	}
	if len(code)>=3 {
		TypeA,err:=Evaluate_Type([]Token{code[0]}, function, program)
		if err!=nil {
			return &Type{}, err
		}
		for i:=0; len(code)>i+1; i+=2 {
			TypeB,err:=Evaluate_Type([]Token{code[i+2]}, function, program)
			if err!=nil {
				return &Type{}, err
			}
			if Type_Signature(TypeA, make([]*Type, 0))==Type_Signature(TypeB, make([]*Type, 0)) {
				if Type_Signature(TypeA, make([]*Type, 0))==Type_Signature(&Type{Is_Raw: true, Raw_Type: INT_TYPE}, make([]*Type, 0)) {
					TypeA=&Type{Is_Raw: true, Raw_Type: INT_TYPE}
					continue
				}
			}
		}
		return TypeA, nil
	}
	return &Type{}, errors.New("could not determine type of the given expression")
}

func Initialise_Temporary_Unique_Variable(variable_Name string, variable_Type *Type, function *Function, program *Program, temp_Variables Temp_Variables) {
	Temp_Id,_:=strconv.ParseInt(strings.Split(variable_Name, "_")[0], 10, 64)
	Temp_Int:=int(Temp_Id)
	Used_Id,_:=strconv.ParseInt(strings.Split(variable_Name, "_")[1], 10, 64)
	Int_Used:=int(Used_Id)
	if !temp_Variables.Variable_Lookup[Temp_Int][Int_Used].Allocated {
		function.Instructions = append(function.Instructions, []string{"var", variable_Name+"->"+Type_Object_To_String(variable_Type, program)+";"})
		temp_Variables.Variable_Lookup[Temp_Int][Int_Used].Allocated=true
	}
}

func Compile_Expression(code []Token, function *Function, program *Program, temp_Variables Temp_Variables) (string, []string, error) {
	out:=""
	used_Variables:=make([]string, 0)
	if len(code)==1 {
		Var_Type,err:=Evaluate_Type([]Token{code[0]}, function, program)
		if err!=nil {
			return out, used_Variables, err
		}
		if code[0].Type=="variable" {
			return code[0].Value, used_Variables, nil
		}
		if code[0].Type=="num" {
			Temp_Var:=Generate_Unique_Temporary_Variable(&Type{Is_Raw: true, Raw_Type: INT_TYPE}, temp_Variables, function)
			Initialise_Temporary_Unique_Variable(Temp_Var, Var_Type, function, program, temp_Variables)
			function.Instructions = append(function.Instructions, []string{"set", Temp_Var, strconv.FormatInt(int64(code[0].Num_Value), 10)+";"})
			return Temp_Var, []string{Temp_Var}, nil
		}
		if code[0].Type=="expression" {
			return Compile_Expression(code[0].Children, function, program, temp_Variables)
		}
		if code[0].Type=="funcall" {
			found_Function:=&Function{}
			function_Name:=""
			for Fn_Name, Fn:=range program.Functions {
				if Fn_Name==code[0].Children[0].Value {
					found_Function=Fn
					function_Name=Fn_Name
					break
				}
			}
			if len(code[0].Children[1].Children)!=len(found_Function.Arguments) {
				return out, make([]string, 0), errors.New("function call arguments do not match length of function call")
			}
			Variables:=make([]string, 0)
			call_String:="("
			for i:=0; len(code[0].Children[1].Children)>i; i+=2 {
				Var, Occupied_Variable, err:=Compile_Expression([]Token{code[0].Children[1].Children[i]}, function, program, temp_Variables)
				if err!=nil {
					return "", used_Variables, err
				}
				for _,variable:=range Occupied_Variable {
					Free_Temporary_Unique_Variable(variable, temp_Variables, function)
				}
				Variables = append(Variables, Var)
				call_String+=Var+", "
			}
			call_String=strings.Trim(call_String, ", ")+")"
			for _,variable:=range Variables {
				if strings.HasPrefix(variable, "temp.") {
					Free_Temporary_Unique_Variable(variable, temp_Variables, function)
				}
			}
			Temp_Var:=Generate_Unique_Temporary_Variable(found_Function.Out_Type, temp_Variables, function)
			Initialise_Temporary_Unique_Variable(Temp_Var, found_Function.Out_Type, function, program, temp_Variables)
			function.Instructions = append(function.Instructions, []string{"call", function_Name+call_String, Temp_Var+";"})
			return Temp_Var, []string{Temp_Var}, nil
		}
		return out, used_Variables, errors.New("could not compile expression")
	}
	if len(code)==3 {
		Type_A,err:=Evaluate_Type([]Token{code[0]}, function, program)
		if err!=nil {
			return out, used_Variables, err
		}
		Type_B,err:=Evaluate_Type([]Token{code[2]}, function, program)
		if err!=nil {
			return out, used_Variables, err
		}
		Var_A, Occupied_Vars, err:=Compile_Expression([]Token{code[0]}, function, program, temp_Variables)
		if err!=nil {
			return out, used_Variables, err
		}
		used_Variables = append(used_Variables, Occupied_Vars...)
		Var_B, Occupied_Vars, err:=Compile_Expression([]Token{code[2]}, function, program, temp_Variables)
		if err!=nil {
			return out, used_Variables, err
		}
		used_Variables = append(used_Variables, Occupied_Vars...)
		if Type_Signature(Type_A, make([]*Type, 0))==Type_Signature(Type_B, make([]*Type, 0)) {
			if Type_Signature(Type_A, make([]*Type, 0))==Type_Signature(&Type{Is_Raw: true, Raw_Type: INT_TYPE}, make([]*Type, 0)) {
				Temp_Var:=Generate_Unique_Temporary_Variable(&Type{Is_Raw: true, Raw_Type: INT_TYPE}, temp_Variables, function)
				Initialise_Temporary_Unique_Variable(Temp_Var, &Type{Is_Raw: true, Raw_Type: INT_TYPE}, function, program, temp_Variables)
				Instructions_Map:=map[string]string{"+":"add", "-":"sub", "*":"mult", "/":"div", "**":"pow", "//":"floor", "%":"mod", "==":"equals", "!=":"nequals", ">":"greater", "<":"smaller", "&&":"and", "||":"or", "^":"xor"}
				Instruction, Instruction_Found:=Instructions_Map[code[1].Value]
				if Instruction_Found {
					function.Instructions = append(function.Instructions, []string{Instruction, Var_A, Var_B, Temp_Var+";"})
					if strings.HasPrefix(Var_A, "temp.") {
						Free_Temporary_Unique_Variable(Var_A, temp_Variables, function)
					}
					if strings.HasPrefix(Var_B, "temp.") {
						Free_Temporary_Unique_Variable(Var_B, temp_Variables, function)
					}
					for _,variable:=range used_Variables {
						Free_Temporary_Unique_Variable(variable, temp_Variables, function)
					}
					return Temp_Var, make([]string, 0), nil
				}
				return "", used_Variables, errors.New("invalid operation between integers")
			}
		}
		return out, used_Variables, errors.New("could not compile expression")
	}
	if len(code)>3 {
		Token_A:=code[0]
		to_Free:=""
		for i:=0; len(code)>i+1; i+=2 {
			out, Occupied_Vars, err:=Compile_Expression(append(append([]Token{}, Token_A), code[i+1:i+3]...), function, program, temp_Variables)
			if to_Free!="" {
				Free_Temporary_Unique_Variable(to_Free, temp_Variables, function)
			}
			if err!=nil {
				return out, used_Variables, err
			}
			for _,Variable:=range Occupied_Vars {
				Free_Temporary_Unique_Variable(Variable, temp_Variables, function)
			}
			Token_A=Token{Type: "variable", Value: out}
			to_Free=out
		}
		out=Token_A.Value
	}
	return out, used_Variables, nil
}

func Function_Parser(code []Token, function_definition Function_Definition, function *Function, program *Program, temp_Variables Temp_Variables, loop_Details struct{In_Loop bool; Loop_Details Loop_Details}) error {
	for i:=0; len(code)>i; i++ {
		if code[i].Type=="sys" && code[i].Value=="var" {
			if !(len(code)-i>=3) {
				return  errors.New("invalid variable declartion during file parsing")
			}
			variables:=make([]string, 0)
			j:=i
			variablesType:=&Type{}
			for {
				j++
				if j>=len(code) {
					return errors.New("unexpected EOF during variable parsing")
				}
				if code[j].Type=="type" {
					if (len(code)-j>1) && code[j+1].Type=="EOS" {
						varType,err:=Parse_Type(code[j], program)
						if err!=nil {
							return err
						}
						variablesType=varType
						break
					} else {
						return errors.New("type should be the last token during variable declarations proceeded by an EOS")
					}
				}
				if code[j].Type!="variable" || !Is_Valid_Var_Name(code[j].Value) {
					return errors.New("invalid variable declaration statement")
				}
				if str_index_in_arr(code[j].Value, variables)!=-1 {
					return errors.New("variable '"+code[j].Value+"' has already been declared")
				}
				for variable:=range function.Scope {
					if variable==code[j].Value {
						return errors.New("variable '"+variable+"' has already been declared")
					}
				}
				variables = append(variables, code[j].Value)
			}
			i=j+1
			Instructions:=make([]string, 0)
			Instructions = append(Instructions, "var")
			for _,variable:=range variables {
				function.Scope[variable]=variablesType
				Instructions = append(Instructions, variable)
			}
			Instructions[len(Instructions)-1]+="->"+Type_Object_To_String(variablesType, program)+";"
			function.Instructions = append(function.Instructions, Instructions)
			continue
		}
		if len(code)-i>=4 && code[i+1].Type=="operator" && code[i+1].Value=="=" {
			LHS_Token:=code[i]
			expression_Tokens:=make([]Token, 0)
			i+=1
			for {
				i++
				if i>=len(code) {
					return errors.New("unexpected EOS during function '"+function_definition.Name+"' parsing")
				}
				if code[i].Type=="EOS" {
					break
				}
				expression_Tokens = append(expression_Tokens, code[i])
			}
			if !Is_Expression_Valid(expression_Tokens) {
				return errors.New("invalid expression")
			}
			RHS_Type,err:=Evaluate_Type(expression_Tokens, function, program)
			if err!=nil {
				return err
			}
			LHS_Type,err:=Evaluate_Type([]Token{LHS_Token}, function, program)
			if err!=nil {
				return err
			}
			if Type_Signature(RHS_Type, make([]*Type, 0))!=Type_Signature(LHS_Type, make([]*Type, 0)) {
				return errors.New("LHS and RHS Types do not match")
			}
			RHS,used_Variables,err:=Compile_Expression(expression_Tokens, function, program, temp_Variables)
			if err!=nil {
				return err
			}
			for _,variable:=range used_Variables {
				Free_Temporary_Unique_Variable(variable, temp_Variables, function)
			}
			if LHS_Token.Type=="variable" {
				function.Instructions = append(function.Instructions, []string{"copy", LHS_Token.Value, RHS+";"})
			}
			if strings.HasPrefix(RHS, "temp.") {
				Free_Temporary_Unique_Variable(RHS, temp_Variables, function)
			}
			continue
		}
		if len(code)-i>=2 && code[i].Type=="sys" && code[i].Value=="return" {
			return_Tokens:=make([]Token, 0)
			for {
				i++
				if i>=len(code) {
					return errors.New("unexpected EOS during function '"+function_definition.Name+"' parsing")
				}
				if code[i].Type=="EOS" {
					break
				}
				return_Tokens = append(return_Tokens, code[i])
			}
			if !(len(return_Tokens)==0 && function.Out_Type.Is_Raw && function.Out_Type.Raw_Type==VOID_TYPE) {
				RHS_Type,err:=Evaluate_Type(return_Tokens, function, program)
				if err!=nil {
					return err
				}
				if Type_Signature(RHS_Type, make([]*Type, 0))!=Type_Signature(function.Out_Type, make([]*Type, 0)) {
					return errors.New("function return statement type does not match function return type")
				}
				RHS,used_Variables,err:=Compile_Expression(return_Tokens, function, program, temp_Variables)
				if err!=nil {
					return err
				}
				for _,variable:=range used_Variables {
					Free_Temporary_Unique_Variable(variable, temp_Variables, function)
				}
				if strings.HasPrefix(RHS, "temp.") {
					Free_Temporary_Unique_Variable(RHS, temp_Variables, function)
				}
				function.Instructions = append(function.Instructions, []string{"return", RHS+";"})
			} else {
				void_Variable:=Generate_Unique_Temporary_Variable(&Type{Is_Raw: true, Raw_Type: VOID_TYPE}, temp_Variables, function)
				function.Instructions = append(function.Instructions, []string{"return", void_Variable+";"})
				Free_Temporary_Unique_Variable(void_Variable, temp_Variables, function)
			}
			continue
		}
		if code[i].Type=="sys" && code[i].Value=="if" {
			if !(len(code)-i>=4) {
				return errors.New("incomplete if condition declaration found")
			}
			Condition_Type,err:=Evaluate_Type([]Token{code[i+1]}, function, program)
			if err!=nil {
				return err
			}
			if Type_Signature(Condition_Type, make([]*Type, 0))!=Type_Signature(&Type{Is_Raw: true, Raw_Type: INT_TYPE}, make([]*Type, 0)) {
				return errors.New("if condition must be of type int")
			}
			condition_Variable,used_Variables,err:=Compile_Expression([]Token{code[i+1]}, function, program, temp_Variables)
			if err!=nil {
				return err
			}
			if code[i+2].Type!="bracket_open" && code[i+2].Value!="{" {
				return errors.New("curly bracket expression expected after if condition")
			}
			conditional_Tokens:=make([]Token, 0)
			i+=1
			brackets:=0
			for {
				i++
				if i>=len(code) {
					return errors.New("unexpected EOS during function '"+function_definition.Name+"' parsing")
				}
				if code[i].Type=="bracket_open" && code[i].Value=="{" {
					brackets+=1
				}
				if code[i].Type=="bracket_close" && code[i].Value=="}" {
					brackets-=1
				}
				conditional_Tokens = append(conditional_Tokens, code[i])
				if brackets==0 {
					break
				}
			}
			conditional_Tokens = conditional_Tokens[1:len(conditional_Tokens)-1]
			Jump_Line_Count_Var:=Generate_Unique_Temporary_Variable(&Type{Is_Raw: true, Raw_Type: INT_TYPE}, temp_Variables, function)
			Initialise_Temporary_Unique_Variable(Jump_Line_Count_Var, &Type{Is_Raw: true, Raw_Type: INT_TYPE}, function, program, temp_Variables)
			Jump_Line_Instruction_Index:=len(function.Instructions)
			function.Instructions = append(function.Instructions, []string{"set", Jump_Line_Count_Var})
			function.Instructions = append(function.Instructions, []string{"not", condition_Variable+";"})
			function.Instructions = append(function.Instructions, []string{"jump", Jump_Line_Count_Var, condition_Variable+";"})
			if strings.HasPrefix("temp.", condition_Variable) {
				Free_Temporary_Unique_Variable(condition_Variable, temp_Variables, function)
			}
			instruction_Count:=len(function.Instructions)
			Free_Temporary_Unique_Variable(Jump_Line_Count_Var, temp_Variables, function)
			for _,variable:=range used_Variables {
				Free_Temporary_Unique_Variable(variable, temp_Variables, function)
			}
			err=Function_Parser(conditional_Tokens, function_definition, function, program, temp_Variables, loop_Details)
			if err!=nil {
				return err
			}
			function.Instructions[Jump_Line_Instruction_Index]=[]string{"set", Jump_Line_Count_Var, strconv.FormatInt(int64(len(function.Instructions)-instruction_Count), 10)+";"}
			continue
		}
		if code[i].Type=="sys" && code[i].Value=="for" {
			if !(len(code)-i>=3) {
				return errors.New("incomplete if condition declaration found")
			}
			if code[i+1].Type!="bracket_open" && code[i+1].Value!="{" {
				return errors.New("curly bracket expression expected after for keyword")
			}
			brackets:=0
			loop_Tokens:=make([]Token, 0)
			for {
				i++
				if i>=len(code) {
					return errors.New("unexpected EOS during function '"+function_definition.Name+"' parsing")
				}
				if code[i].Type=="bracket_open" && code[i].Value=="{" {
					brackets+=1
				}
				if code[i].Type=="bracket_close" && code[i].Value=="}" {
					brackets-=1
				}
				loop_Tokens = append(loop_Tokens, code[i])
				if brackets==0 {
					break
				}
			}
			loop_Tokens=loop_Tokens[1:len(loop_Tokens)-1]
			Loop_Start:=Generate_Unique_Temporary_Variable(&Type{Is_Raw: true, Raw_Type: INT_TYPE}, temp_Variables, function)
			Initialise_Temporary_Unique_Variable(Loop_Start, &Type{Is_Raw: true, Raw_Type: INT_TYPE}, function, program, temp_Variables)
			Loop_Break:=Generate_Unique_Temporary_Variable(&Type{Is_Raw: true, Raw_Type: INT_TYPE}, temp_Variables, function)
			Initialise_Temporary_Unique_Variable(Loop_Break, &Type{Is_Raw: true, Raw_Type: INT_TYPE}, function, program, temp_Variables)
			Variable_Index:=len(function.Instructions)
			function.Instructions = append(function.Instructions, []string{"set", Loop_Break, "0;"})
			function.Instructions = append(function.Instructions, []string{"set", Loop_Start, strconv.FormatInt(int64(len(function.Instructions)+1), 10)+";"})
			err:=Function_Parser(loop_Tokens, function_definition, function, program, temp_Variables, struct{In_Loop bool; Loop_Details Loop_Details}{In_Loop: true, Loop_Details: Loop_Details{Continue_Variable: Loop_Start, Break_Variable: Loop_Break}})
			function.Instructions = append(function.Instructions, []string{"jumpto", Loop_Start+";"})
			if err!=nil {
				return err
			}
			function.Instructions[Variable_Index]=[]string{"set", Loop_Break, strconv.FormatInt(int64(len(function.Instructions))+1, 10)+";"}
			Free_Temporary_Unique_Variable(Loop_Break, temp_Variables, function)
			Free_Temporary_Unique_Variable(Loop_Start, temp_Variables, function)
			continue
		}
		if code[i].Type=="sys" && code[i].Value=="break" {
			if !(len(code)-i>=2) {
				return errors.New("semicolon not found after break keyword")
			}
			if !loop_Details.In_Loop {
				return errors.New("break keyword used outside loop")
			}
			function.Instructions = append(function.Instructions, []string{"jumpto", loop_Details.Loop_Details.Break_Variable+";"})
			i+=1
			continue
		}
		if code[i].Type=="sys" && code[i].Value=="continue" {
			if !(len(code)-i>=2) {
				return errors.New("semicolon not found after continue keyword")
			}
			if !loop_Details.In_Loop {
				return errors.New("continue keyword used outside loop")
			}
			function.Instructions = append(function.Instructions, []string{"jumpto", loop_Details.Loop_Details.Continue_Variable+";"})
			i+=1
			continue
		}
		return errors.New("unexpected token of type '"+code[i].Type+"' inside function '"+function_definition.Name+"'")
	}
	return nil
}