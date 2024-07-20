package venc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	INT_TYPE               int8 = iota
	INT64_TYPE             int8 = iota
	STRING_TYPE            int8 = iota
	FLOAT_TYPE             int8 = iota
	FLOAT64_TYPE           int8 = iota
	POINTER_TYPE           int8 = iota
	VOID_TYPE              int8 = iota
)

type Token struct {
	Type                 string
	Num_Value            float64
	Value         string
	String_Children      []string
	Children             []Token
}

type Type struct {
	Is_Array             bool
	Is_Dict              bool
	Is_Raw               bool
	Raw_Type             int8
	Is_Struct            bool
	Is_Pointer           bool
	Struct_Details       map[string]*Type
	Child                *Type
}

type Function struct {
	Name                 string
	Out_Type             Type
	Arguments            map[string]Type
	Scope                map[string]Type
	Instructions         [][]string
}

type Program struct {
	Path                 string
	Structs              map[string]*Type
	Functions            []Function
	Global_Variables     map[string]Type
	Imported_Libraries   map[string]*Program
}

type Function_Definition struct {
	Name                 string
	Arguments            map[string]Token
	Out_Type             Token
	Internal_Tokens      []Token
}

type Definitions struct {
	Imports              map[string]string
	Variables            map[string]Token
	Functions            []Function_Definition
	Structs              map[string]map[string]Token
}

var reserved_tokens = []string{"var", "fn", "if", "while", "continue", "break", "struct", "return", "function", "as", "import", "package"}
var types = []string{"int", "int64", "string", "float", "float64", "void"}
var operators = []string{"+", "-", "*", "/", "^", ">", "<", "=", "&", "!", "|", "&", "%", ":=", "**", "&&", "||", "//"}
var end_of_statements = []string{";"}
var brackets = []string{"(", ")", "[", "]", "{", "}"}
var string_quotes = []string{"\"", "'"}
var comma = ","
var allowed_variable_character = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
var comments = []string{"#"}

func str_index_in_arr(a string, arr []string) int {
	for i:=0; i<len(arr); i++ {
		if a==arr[i] {
			return i
		}
	}
	return -1
}

func Tokensier(code string, debug bool) []Token {
	tokens := make([]Token, 0)
	cache := ""
	for i := 0; i < len(code); i++ {
		char := string(code[i])
		if strings.Contains("1234567890.", char) && cache=="" {
			for {
				char := string(code[i])
				if strings.Contains(".1234567890", char) {
					cache += char
				} else {
					if cache == "." {
						tokens = append(tokens, Token{Type: "dot"})
						cache = ""
						i--
						break
					}
					number, _ := strconv.ParseFloat(cache, 64)
					tokens = append(tokens, Token{Type: "num", Num_Value: number})
					i--
					cache = ""
					break
				}
				i++
				if i == len(code) {
					if cache == "." {
						tokens = append(tokens, Token{Type: "dot"})
						cache = ""
						i--
						break
					}
					number, _ := strconv.ParseFloat(cache, 64)
					tokens = append(tokens, Token{Type: "num", Num_Value: number})
					i--
					cache = ""
					break
				}
			}
			continue
		}
		if str_index_in_arr(char, brackets) != -1 {
			open_type := ""
			if (str_index_in_arr(char, brackets) % 2) != 1 {
				open_type = "open"
			} else {
				open_type = "close"
			}
			tokens = append(tokens, Token{Type: "bracket_" + open_type, Value: char})
			continue
		}
		if str_index_in_arr(char, operators) != -1 {
			tokens = append(tokens, Token{Type: "operator", Value: char})
			continue
		}
		if str_index_in_arr(char, end_of_statements) != -1 {
			tokens = append(tokens, Token{Type: "EOS"})
			continue
		}
		if str_index_in_arr(char, string_quotes) != -1 {
			string_init := char
			for {
				i++
				if i == len(code) {
					if debug {
						fmt.Println("Unexpected EOF")
					}
					return make([]Token, 0)
				}
				char := string(code[i])
				if char != string_init {
					cache += char
				} else {
					tokens = append(tokens, Token{Type: "string", Value: cache})
					cache = ""
					break
				}
			}
			continue
		}
		if char == comma {
			tokens = append(tokens, Token{Type: "comma"})
			continue
		}
		if char == ":" {
			tokens = append(tokens, Token{Type: "colon"})
			continue
		}
		if char == "." {
			tokens = append(tokens, Token{Type: "dot"})
			continue
		}
		if strings.Contains(allowed_variable_character, char) {
			for {
				char := string(code[i])
				if strings.Contains(allowed_variable_character, char) || strings.Contains("1234567890", char) {
					cache += char
				} else {
					if str_index_in_arr(cache, reserved_tokens) != -1 {
						tokens = append(tokens, Token{Type: "sys", Value: cache})
					} else {
						tokens = append(tokens, Token{Type: "variable", Value: cache})
					}
					cache = ""
					i--
					break
				}
				i++
				if i == len(code) {
					tokens = append(tokens, Token{Type: "variable", Value: cache})
					cache = ""
					i--
					break
				}
			}
			continue
		}
		if str_index_in_arr(char, comments) != -1 {
			for {
				i++
				if i == len(code) {
					break
				}
				char := string(code[i])
				if str_index_in_arr(char, end_of_statements) == -1 && char != "\n" {
					continue
				} else {
					break
				}
			}
			continue
		}
	}
	return tokens
}

func Is_Valid_Var_Name(var_name string) bool {
	if var_name == "" {
		return false
	}
	if str_index_in_arr(var_name, reserved_tokens) != -1 {
		return false
	}
	for _, char := range "1234567890" {
		var_name = strings.ReplaceAll(var_name, string(char), "")
	}
	if var_name == "" {
		return false
	}
	for _, char := range allowed_variable_character {
		var_name = strings.ReplaceAll(var_name, string(char), "")
	}
	return var_name == ""
}

func Type_Tokens_Parser(tokens []Token, depth int) (Token, error) {
	if len(tokens)==0 {
		return Token{}, errors.New("unexpected EOF during type declaration")
	}
	if len(tokens)==1 {
		if tokens[0].Type!="variable" {
			return Token{}, errors.New("type declaration is invalid "+tokens[0].Type)
		}
		if tokens[0].Value=="void" && depth!=0 {
			return Token{}, errors.New("void type can only be used as is")
		}
		return Token{Type: "raw", Value: tokens[0].Value}, nil
	}
	if tokens[0].Type=="bracket_open" && tokens[0].Value=="[" && tokens[len(tokens)-1].Type=="bracket_close" && tokens[len(tokens)-1].Value=="]" {
		arr_Token,err:=Type_Tokens_Parser(tokens[1:len(tokens)-1], depth+1)
		if err!=nil {
			return Token{}, err
		}
		return Token{Type: "array", Children: []Token{arr_Token}}, nil
	}
	if tokens[0].Type=="operator" && tokens[0].Value=="*" {
		pointer_Token,err:=Type_Tokens_Parser(tokens[1:], depth+1)
		if err!=nil {
			return Token{}, err
		}
		return Token{Type: "pointer", Children: []Token{pointer_Token}}, nil
	}
	if len(tokens)>=6 && tokens[0].Type=="bracket_open" && tokens[0].Value=="{" && tokens[1].Type=="variable" && str_index_in_arr(tokens[1].Value, types)!=-1 && tokens[1].Value!="void" && tokens[2].Type=="operator" && tokens[2].Value=="-" && tokens[3].Type=="operator" && tokens[3].Value==">" && tokens[len(tokens)-1].Type=="bracket_close" && tokens[len(tokens)-1].Value=="}" {
		dict_Token,err:=Type_Tokens_Parser(tokens[4:len(tokens)-1], depth+1)
		if err!=nil {
			return Token{}, err
		}
		return Token{Type: "dict", Children: []Token{Token{Type: "raw", Children: []Token{tokens[1]}}, dict_Token}}, nil
	}
	return Token{}, errors.New("type declaration is invalid")
}

func Tokens_Parser(code []Token, debug bool) ([]Token, error) {
	parsed_tokens := make([]Token, 0)
	for i := 0; i < len(code); i++ {
		current_token := code[i]
		if len(code) > i+1 && current_token.Type == "operator" && code[i+1].Type == "operator" {
			combined_operator := current_token.Value + code[i+1].Value
			accepted_combined_operator_array := []string{"+=", "-=", "/=", "*=", "%=", "//", "!=", "==", "->", ":=", "||", "&&"}
			if str_index_in_arr(combined_operator, accepted_combined_operator_array) != -1 {
				parsed_tokens = append(parsed_tokens, Token{Type: "operator", Value: combined_operator})
				i++
				continue
			}
		}
		if len(code) > i+1 && current_token.Type == "colon" && code[i+1].Type == "operator" {
			combined_operator := ":" + code[i+1].Value
			if combined_operator == ":=" {
				parsed_tokens = append(parsed_tokens, Token{Type: "operator", Value: combined_operator})
				i++
				continue
			}
		}
		if len(parsed_tokens) > 0 && parsed_tokens[len(parsed_tokens)-1].Type == "operator" && parsed_tokens[len(parsed_tokens)-1].Value == "->" {
			if str_index_in_arr(current_token.Type, []string{"variable", "bracket_open", "operator"})==-1 {
				return make([]Token, 0), errors.New("type declaration is invalid")
			}
			Type_Tokens := make([]Token, 0)
			brackets := 0
			for {
				if len(code) < i+1 {
					return make([]Token, 0), errors.New("unexpected EOF")
				}
				if code[i].Type == "bracket_open" && (code[i].Value=="[" || code[i].Value=="{") {
					brackets += 1
				}
				if code[i].Type == "bracket_close" && (code[i].Value=="]" || code[i].Value=="}") {
					brackets -= 1
				}
				Type_Tokens = append(Type_Tokens, code[i])
				if code[i].Type=="operator" && code[i].Value=="*" {
					i++
					continue
				}
				if brackets == 0 {
					break
				}
				i++
			}
			Type_Token,err:=Type_Tokens_Parser(Type_Tokens, 0)
			if err!=nil {
				return make([]Token, 0), err
			}
			parsed_tokens[len(parsed_tokens)-1] = Token{Type: "type", Children: []Token{Type_Token}}
			continue
		}
		if code[i].Type == "bracket_open" && code[i].Value == "[" {
			bracket_count := 1
			childrentokens := make([]Token, 0)
			for {
				i++
				if len(code) < i+1 {
					return make([]Token, 0), errors.New("unexpected EOF")
				}
				if code[i].Type == "bracket_open" && code[i].Value == "[" {
					bracket_count += 1
				}
				if code[i].Type == "bracket_close" && code[i].Value == "]" {
					bracket_count -= 1
				}
				if bracket_count == 0 {
					break
				}
				childrentokens = append(childrentokens, code[i])
			}
			tokens, err := Tokens_Parser(childrentokens, debug)
			if err != nil {
				return make([]Token, 0), err
			}
			parsed_tokens = append(parsed_tokens, Token{Type: "expression_wrapper_[]", Children: tokens})
			continue
		}
		if code[i].Type == "bracket_open" && code[i].Value == "(" {
			bracket_count := 1
			children_Tokens := make([]Token, 0)
			for {
				i++
				if len(code) < i+1 {
					return make([]Token, 0), errors.New("unexpected EOF")
				}
				if code[i].Type == "bracket_open" && code[i].Value == "(" {
					bracket_count += 1
				}
				if code[i].Type == "bracket_close" && code[i].Value == ")" {
					bracket_count -= 1
				}
				if bracket_count == 0 {
					break
				}
				children_Tokens = append(children_Tokens, code[i])
			}
			tokens, err := Tokens_Parser(children_Tokens, debug)
			if err != nil {
				return make([]Token, 0), err
			}
			parsed_tokens = append(parsed_tokens, Token{Type: "expression", Children: tokens})
			continue
		}
		parsed_tokens = append(parsed_tokens, current_token)
	}
	return parsed_tokens, nil
}

func Token_Grouper(code []Token, debug bool) ([]Token, error) {
	grouped_tokens := make([]Token, 0)
	for i := 0; i < len(code); i++ {
		tokens_children, err := Token_Grouper(code[i].Children, false)
		if err != nil {
			return make([]Token, 0), err
		}
		code[i].Children = tokens_children
		if code[i].Type == "expression_wrapper_[]" {
			if len(grouped_tokens) > 0 && (grouped_tokens[len(grouped_tokens)-1].Type == "variable" || grouped_tokens[len(grouped_tokens)-1].Type == "lookup" || grouped_tokens[len(grouped_tokens)-1].Type == "expression" || grouped_tokens[len(grouped_tokens)-1].Type == "funcall" || grouped_tokens[len(grouped_tokens)-1].Type == "array") {
				grouped_tokens[len(grouped_tokens)-1] = Token{Type: "lookup", Children: []Token{Token{Type: "parent", Children: []Token{grouped_tokens[len(grouped_tokens)-1]}}, Token{Type: "tokens", Children: code[i].Children}}}
				continue
			}
		}
		if len(grouped_tokens) > 0 && code[i].Type == "type" && grouped_tokens[len(grouped_tokens)-1].Type == "expression_wrapper_[]" {
			grouped_tokens[len(grouped_tokens)-1] = Token{Type: "array", Children: []Token{code[i], grouped_tokens[len(grouped_tokens)-1]}}
			continue
		}
		if len(grouped_tokens) > 0 && code[i].Type == "expression" {
			if grouped_tokens[len(grouped_tokens)-1].Type == "variable" || grouped_tokens[len(grouped_tokens)-1].Type == "lookup" || grouped_tokens[len(grouped_tokens)-1].Type == "expression" || grouped_tokens[len(grouped_tokens)-1].Type == "funcall" || grouped_tokens[len(grouped_tokens)-1].Type == "nested_tokens" {
				grouped_tokens[len(grouped_tokens)-1] = Token{Type: "funcall", Children: []Token{grouped_tokens[len(grouped_tokens)-1], code[i]}}
				continue
			}
		}
		grouped_tokens = append(grouped_tokens, code[i])
	}
	i:=-1
	for {
		i+=1
		if i>=len(grouped_tokens) {
			break
		}
		if len(grouped_tokens)-i>1 && grouped_tokens[i].Type=="dot" {
			grouped_tokens[i-1]=Token{Type: "field_access", Children: []Token{grouped_tokens[i-1], grouped_tokens[i+1]}}
			grouped_tokens = append(grouped_tokens[:i], grouped_tokens[i+2:]...)
			i-=2
			continue
		}
		if i>0 && grouped_tokens[i].Type=="operator" && grouped_tokens[i-1].Type=="operator" {
			if str_index_in_arr(grouped_tokens[i].Value+grouped_tokens[i-1].Value, operators)!=-1 {
				grouped_tokens[i-1].Value=grouped_tokens[i].Value+grouped_tokens[i-1].Value
				grouped_tokens=append(grouped_tokens[:i], grouped_tokens[i+1:]...)
				i-=1
				continue
			}
		}
		if i>0 && grouped_tokens[i].Type=="operator" && grouped_tokens[i].Value=="=" && grouped_tokens[i-1].Type=="colon" {
			grouped_tokens[i].Value=":="
			grouped_tokens=append(grouped_tokens[:i-1], grouped_tokens[i:]...)
			i-=1
			continue
		}
	}
	return grouped_tokens, nil
}

func Definition_Parser(code []Token) (Definitions, error) {
	definitions:=Definitions{Imports: make(map[string]string), Variables: make(map[string]Token), Functions: make([]Function_Definition, 0), Structs: make(map[string]map[string]Token)}
	for i:=0; i<len(code); i++ {
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
				if code[j].Type!="variable" || !Is_Valid_Var_Name(code[j].Value) || str_index_in_arr(code[j].Value, variables)!=-1 {
					return definitions, errors.New("invalid variable declaration statement")
				}
				for variable:=range definitions.Variables {
					if variable==code[j].Value {
						return definitions, errors.New("invalid variable declaration statement")
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
		return definitions, errors.New("unexpected token of type '"+code[i].Type+"'")
	}
	return definitions, nil
}

func Compile() {}