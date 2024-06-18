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
var operators = []string{"+", "-", "*", "/", "^", ">", "<", "=", "&", "!", "|", "%", ":="}
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

func Tokens_Parser(code []Token, debug bool) ([]Token, error) {
	parsed_tokens := make([]Token, 0)
	for i := 0; i < len(code); i++ {
		current_token := code[i]
		if len(code) > i+1 && code[i+1].Type == "dot" {
			nested_variables := make([]string, 0)
			first := true
			for {
				if len(code) > i+1 && (!first || code[i+1].Type == "dot") {
					if !first {
						if code[i-1].Type != "dot" {
							break
						}
					}
					nested_variables = append(nested_variables, code[i].Value)
					if code[i+1].Type == "dot" {
						i++
					}
					i++
					first = false
				} else {
					if !first && current_token.Type == "variable" && Is_Valid_Var_Name(code[i].Value) && code[i-1].Type == "dot" {
						nested_variables = append(nested_variables, code[i].Value)
						i++
					}
					break
				}
			}
			parsed_tokens = append(parsed_tokens, Token{Type: "nested_tokens", String_Children: nested_variables})
			i--
			continue
		}
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
		if len(parsed_tokens) > 0 && parsed_tokens[len(parsed_tokens)-1].Type == "operator" && parsed_tokens[len(parsed_tokens)-1].Value == "->" && ( str_index_in_arr(current_token.Value, types)!=-1 || (len(code) > i+1 && len(parsed_tokens) != 0 && current_token.Type == "bracket_open" && (current_token.Value == "[" || current_token.Value == "{"))) {
			type_tokens := make([]string, 0)
			brackets := 0
			for {
				if len(code) < i+1 {
					return make([]Token, 0), errors.New("Unexpected EOF")
				}
				if code[i].Type == "bracket_open" && code[i].Value == current_token.Value {
					brackets += 1
				}
				if code[i].Type == "bracket_close" && ((code[i].Value == "]" && current_token.Value == "[") || (code[i].Value == "}" && current_token.Value == "{")) {
					brackets -= 1
				}
				if str_index_in_arr(code[i].Type, []string{"bracket_open", "bracket_close", "variable"}) == -1 {
					return make([]Token, 0), errors.New("Illegal type definition")
				}
				type_tokens = append(type_tokens, code[i].Value)
				if brackets == 0 {
					break
				}
				i++
			}
			if str_index_in_arr("void", type_tokens)!=-1 && len(type_tokens)!=1 {
				return make([]Token, 0), errors.New("void type can only be used as is")
			}
			parsed_tokens[len(parsed_tokens)-1] = Token{Type: "type", String_Children: type_tokens}
			continue
		}
		if code[i].Type == "bracket_open" && code[i].Value == "[" {
			bracket_count := 1
			childrentokens := make([]Token, 0)
			for {
				i++
				if len(code) < i+1 {
					return make([]Token, 0), errors.New("Unexpected EOF")
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
					return make([]Token, 0), errors.New("Unexpected EOF")
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
			if len(grouped_tokens) > 0 && (grouped_tokens[len(grouped_tokens)-1].Type == "variable" || grouped_tokens[len(grouped_tokens)-1].Type == "nested_tokens" || grouped_tokens[len(grouped_tokens)-1].Type == "lookup" || grouped_tokens[len(grouped_tokens)-1].Type == "expression" || grouped_tokens[len(grouped_tokens)-1].Type == "funcall" || grouped_tokens[len(grouped_tokens)-1].Type == "array") {
				grouped_tokens[len(grouped_tokens)-1] = Token{Type: "lookup", Children: []Token{Token{Type: "parent", Children: []Token{grouped_tokens[len(grouped_tokens)-1]}}, Token{Type: "tokens", Children: code[i].Children}}}
				continue
			}
		}
		if len(grouped_tokens) > 0 && code[i].Type == "type" && grouped_tokens[len(grouped_tokens)-1].Type == "expression_wrapper_[]" {
			grouped_tokens[len(grouped_tokens)-1] = Token{Type: "array", Children: []Token{code[i], grouped_tokens[len(grouped_tokens)-1]}}
			continue
		}
		if len(grouped_tokens) > 1 && grouped_tokens[len(grouped_tokens)-1].Type == "dot" && (grouped_tokens[len(grouped_tokens)-2].Type == "nested_tokens" || grouped_tokens[len(grouped_tokens)-2].Type == "lookup" || grouped_tokens[len(grouped_tokens)-2].Type == "variable" || grouped_tokens[len(grouped_tokens)-2].Type == "expression" || grouped_tokens[len(grouped_tokens)-2].Type == "funcall") {
			grouped_tokens[len(grouped_tokens)-2] = Token{Type: "nested_tokens", Children: []Token{grouped_tokens[len(grouped_tokens)-2], code[i]}}
			grouped_tokens = grouped_tokens[:len(grouped_tokens)-1]
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
		return definitions, errors.New("unexpected token of type '"+code[i].Type+"'")
	}
	return definitions, nil
}

func Compile() {}