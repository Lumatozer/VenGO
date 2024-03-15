package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

type Token struct {
	Type         string
	num_value    float64
	string_value string
	children     []Token
	keys         []Token
}

type Serializable_Token struct {
	Type         string
	Num_value    float64
	String_value string
	Children     []Serializable_Token
	Keys         []Serializable_Token
}

type Function struct {
	name string
	args map[string][]string
	args_keys []string
	Type []string
	Code []Token
}

type Struct struct {
	name           string
	fields         map[string][]string
	// fields_keys    []string no for
}

type Variable struct {
	name string
	Type []string
}

type Scope struct {
	looping                  	bool
	current_loop_continue_line  int
	current_loop_break_line  	int
	current_return_jump_line 	int64
}

// add scope
type Symbol_Table struct {
	functions           []Function
	structs             []Struct
	variables           []Variable
	data                []string
	operations          map[string][][]string
	used_variables      map[string][]int
	variable_mapping    map[string]string
	current_scope       Scope
	files               map[string]Symbol_Table
	current_file        string
	struct_registration [][]string
	finished_importing  bool
	imported_libraries  map[string]string
	global_variables    [][]string
	struct_mapping      map[string]string
}

var reserved_tokens = []string{"var", "fn", "if", "while", "continue", "break", "struct", "return", "function", "as", "import","print","len"}
var type_tokens = []string{"string", "num"}
var operators = []string{"+", "-", "*", "/", "^", ">", "<", "=", "&", "!", "|", "%"}
var end_of_statements = []string{";"}
var brackets = []string{"(", ")", "[", "]", "{", "}"}
var string_quotes = []string{"\"", "'"}
var comma = ","
var allowed_variable_character = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
var comments = []string{"#"}

func unpack_token_struct(orig Serializable_Token) Token {
	new_children := make([]Token, 0)
	for _, child := range orig.Children {
		new_children = append(new_children, unpack_token_struct(child))
	}
	return Token{Type: orig.Type, string_value: orig.String_value, num_value: orig.Num_value, children: new_children}
}

func CloneMyStruct(orig Token) (Token, Serializable_Token, error) {
	new_children := make([]Serializable_Token, 0)
	for _, child := range orig.children {
		_, new_child, _ := CloneMyStruct(child)
		new_children = append(new_children, new_child)
	}
	new_keys := make([]Serializable_Token, 0)
	for _, key := range orig.children {
		_, new_key, _ := CloneMyStruct(key)
		new_keys = append(new_keys, new_key)
	}
	new_orig := Serializable_Token{Type: (orig).Type, Children: new_children, String_value: (orig).string_value, Num_value: (orig).num_value, Keys: new_keys}
	origJSON, err := json.Marshal(&new_orig)
	if err != nil {
		return Token{Type: "check clone my struct 1"}, Serializable_Token{Type: "check clone my struct 1"}, err
	}

	clone := Serializable_Token{}
	if err = json.Unmarshal(origJSON, &clone); err != nil {
		return Token{Type: "check clone my struct 2"}, Serializable_Token{Type: "check clone my struct 2"}, err
	}
	new_unpacked_children := make([]Token, 0)
	for _, child := range clone.Children {
		new_unpacked_children = append(new_unpacked_children, unpack_token_struct(child))
	}
	return Token{Type: clone.Type, children: new_unpacked_children, string_value: clone.String_value, num_value: clone.Num_value}, clone, nil
}

func remove_token_at_index(i int, tokens []Token) []Token {
	return append(tokens[:i], tokens[i+1:]...)
}

func tokensier(code string, debug bool) []Token {
	tokens := make([]Token, 0)
	cache := ""
	for i := 0; i < len(code); i++ {
		char := string(code[i])
		if strings.Contains("1234567890.", char) {
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
					tokens = append(tokens, Token{Type: "num", num_value: number})
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
					tokens = append(tokens, Token{Type: "num", num_value: number})
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
			tokens = append(tokens, Token{Type: "bracket_" + open_type, string_value: char})
			continue
		}
		if str_index_in_arr(char, operators) != -1 {
			tokens = append(tokens, Token{Type: "operator", string_value: char})
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
					tokens = append(tokens, Token{Type: "string", string_value: str_parser(cache)})
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
				if strings.Contains(allowed_variable_character, char) {
					cache += char
				} else {
					if str_index_in_arr(cache, reserved_tokens) != -1 {
						tokens = append(tokens, Token{Type: "sys", string_value: cache})
					} else {
						tokens = append(tokens, Token{Type: "variable", string_value: cache})
					}
					cache = ""
					i--
					break
				}
				i++
				if i == len(code) {
					tokens = append(tokens, Token{Type: "variable", string_value: cache})
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

func valid_var_name(var_name string) bool {
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

func tokens_parser(code []Token, debug bool) ([]Token, error) {
	parsed_tokens := make([]Token, 0)
	for i := 0; i < len(code); i++ {
		current_token := code[i]
		if len(code) > i+1 && code[i+1].Type == "dot" {
			nested_variables := make([]Token, 0)
			first := true
			for {
				if len(code) > i+1 && (!first || code[i+1].Type == "dot") {
					if !first {
						if code[i-1].Type != "dot" {
							break
						}
					}
					nested_variables = append(nested_variables, code[i])
					if code[i+1].Type == "dot" {
						i++
					}
					i++
					first = false
				} else {
					if !first && current_token.Type == "variable" && valid_var_name(code[i].string_value) && code[i-1].Type == "dot" {
						nested_variables = append(nested_variables, code[i])
						i++
					}
					break
				}
			}
			parsed_tokens = append(parsed_tokens, Token{Type: "nested_tokens", children: nested_variables})
			i--
			continue
		}
		if len(code) > i+1 && current_token.Type == "operator" && code[i+1].Type == "operator" {
			combined_operator := current_token.string_value + code[i+1].string_value
			accepted_combined_operator_array := []string{"+=", "-=", "/=", "*=", "%=", "//", "!=", "==", "->", ":=", "||", "&&"}
			if str_index_in_arr(combined_operator, accepted_combined_operator_array) != -1 {
				parsed_tokens = append(parsed_tokens, Token{Type: "operator", string_value: combined_operator})
				i++
				continue
			}
		}
		if len(code) > i+1 && current_token.Type == "colon" && code[i+1].Type == "operator" {
			combined_operator := ":" + code[i+1].string_value
			if combined_operator == ":=" {
				parsed_tokens = append(parsed_tokens, Token{Type: "operator", string_value: combined_operator})
				i++
				continue
			}
		}
		if len(parsed_tokens) > 0 && parsed_tokens[len(parsed_tokens)-1].Type == "operator" && parsed_tokens[len(parsed_tokens)-1].string_value == "->" && ((current_token.string_value == "string" || current_token.string_value == "num" || current_token.Type == "variable") || (len(code) > i+1 && len(parsed_tokens) != 0 && current_token.Type == "bracket_open" && (current_token.string_value == "[" || current_token.string_value == "{"))) {
			type_define_tokens := make([]Token, 0)
			brackets := 0
			for {
				if len(code) < i+1 {
					return make([]Token, 0), errors.New("Unexpected EOF")
				}
				if code[i].Type == "bracket_open" && code[i].string_value == current_token.string_value {
					brackets += 1
				}
				if code[i].Type == "bracket_close" && ((code[i].string_value == "]" && current_token.string_value == "[") || (code[i].string_value == "}" && current_token.string_value == "{")) {
					brackets -= 1
				}
				if str_index_in_arr(code[i].Type, []string{"bracket_open", "bracket_close", "variable"}) == -1 {
					return make([]Token, 0), errors.New("Illegal type definition")
				}
				type_define_tokens = append(type_define_tokens, code[i])
				if brackets == 0 {
					break
				}
				i++
			}
			parsed_tokens[len(parsed_tokens)-1] = Token{Type: "type", keys: type_define_tokens}
			continue
		}
		if len(code) > i+1 && len(parsed_tokens) != 0 && current_token.Type == "variable" && str_index_in_arr(current_token.string_value, type_tokens) != -1 && parsed_tokens[len(parsed_tokens)-1].Type == "operator" && parsed_tokens[len(parsed_tokens)-1].string_value == "->" {
			parsed_tokens[len(parsed_tokens)-1] = Token{Type: "type", children: []Token{current_token}}
			continue
		}
		if code[i].Type == "bracket_open" && code[i].string_value == "[" {
			bracket_count := 1
			childrentokens := make([]Token, 0)
			for {
				i++
				if len(code) < i+1 {
					return make([]Token, 0), errors.New("Unexpected EOF")
				}
				if code[i].Type == "bracket_open" && code[i].string_value == "[" {
					bracket_count += 1
				}
				if code[i].Type == "bracket_close" && code[i].string_value == "]" {
					bracket_count -= 1
				}
				if bracket_count == 0 {
					break
				}
				childrentokens = append(childrentokens, code[i])
			}
			tokens, err := tokens_parser(childrentokens, debug)
			if err != nil {
				return make([]Token, 0), err
			}
			parsed_tokens = append(parsed_tokens, Token{Type: "expression_wrapper_[]", children: tokens})
			continue
		}
		if code[i].Type == "bracket_open" && code[i].string_value == "(" {
			bracket_count := 1
			childrentokens := make([]Token, 0)
			for {
				i++
				if len(code) < i+1 {
					return make([]Token, 0), errors.New("Unexpected EOF")
				}
				if code[i].Type == "bracket_open" && code[i].string_value == "(" {
					bracket_count += 1
				}
				if code[i].Type == "bracket_close" && code[i].string_value == ")" {
					bracket_count -= 1
				}
				if bracket_count == 0 {
					break
				}
				childrentokens = append(childrentokens, code[i])
			}
			tokens, err := tokens_parser(childrentokens, debug)
			if err != nil {
				return make([]Token, 0), err
			}
			parsed_tokens = append(parsed_tokens, Token{Type: "expression", children: tokens})
			continue
		}
		parsed_tokens = append(parsed_tokens, current_token)
	}
	return parsed_tokens, nil
}

func token_grouper(code []Token, debug bool) ([]Token, error) {
	grouped_tokens := make([]Token, 0)
	for i := 0; i < len(code); i++ {
		tokens_children, err := token_grouper(code[i].children, false)
		if err != nil {
			return make([]Token, 0), err
		}
		code[i].children = tokens_children
		if code[i].Type == "expression_wrapper_[]" {
			if len(grouped_tokens) > 0 && (grouped_tokens[len(grouped_tokens)-1].Type == "variable" || grouped_tokens[len(grouped_tokens)-1].Type == "nested_tokens" || grouped_tokens[len(grouped_tokens)-1].Type == "lookup" || grouped_tokens[len(grouped_tokens)-1].Type == "expression" || grouped_tokens[len(grouped_tokens)-1].Type == "funcall" || grouped_tokens[len(grouped_tokens)-1].Type == "array") {
				grouped_tokens[len(grouped_tokens)-1] = Token{Type: "lookup", children: []Token{Token{Type: "parent", children: []Token{grouped_tokens[len(grouped_tokens)-1]}}, Token{Type: "tokens", children: code[i].children}}}
				continue
			}
		}
		if len(grouped_tokens) > 0 && code[i].Type == "type" && grouped_tokens[len(grouped_tokens)-1].Type == "expression_wrapper_[]" {
			grouped_tokens[len(grouped_tokens)-1] = Token{Type: "array", children: []Token{code[i], grouped_tokens[len(grouped_tokens)-1]}}
			continue
		}
		if len(grouped_tokens) > 1 && grouped_tokens[len(grouped_tokens)-1].Type == "dot" && (grouped_tokens[len(grouped_tokens)-2].Type == "nested_tokens" || grouped_tokens[len(grouped_tokens)-2].Type == "lookup" || grouped_tokens[len(grouped_tokens)-2].Type == "variable" || grouped_tokens[len(grouped_tokens)-2].Type == "expression" || grouped_tokens[len(grouped_tokens)-2].Type == "funcall") {
			grouped_tokens[len(grouped_tokens)-2] = Token{Type: "nested_tokens", children: []Token{grouped_tokens[len(grouped_tokens)-2], code[i]}}
			grouped_tokens = grouped_tokens[:len(grouped_tokens)-1]
			continue
		}
		if len(grouped_tokens) > 0 && code[i].Type == "expression" {
			if grouped_tokens[len(grouped_tokens)-1].Type == "sys" && (grouped_tokens[len(grouped_tokens)-1].string_value == "print" || grouped_tokens[len(grouped_tokens)-1].string_value == "len") {
				grouped_tokens[len(grouped_tokens)-1].Type="variable"
			}
			if grouped_tokens[len(grouped_tokens)-1].Type == "variable" || grouped_tokens[len(grouped_tokens)-1].Type == "lookup" || grouped_tokens[len(grouped_tokens)-1].Type == "expression" || grouped_tokens[len(grouped_tokens)-1].Type == "funcall" || grouped_tokens[len(grouped_tokens)-1].Type == "nested_tokens" {
				grouped_tokens[len(grouped_tokens)-1] = Token{Type: "funcall", children: []Token{grouped_tokens[len(grouped_tokens)-1], code[i]}}
				continue
			}
		}
		grouped_tokens = append(grouped_tokens, code[i])
	}
	return grouped_tokens, nil
}

func deep_check(tokens []Token) bool {
	for _, token := range tokens {
		if token.Type == "nested_tokens" {
			if len(token.children) < 2 {
				return false
			}
		} else {
			if !deep_check(token.children) {
				return false
			}
		}
	}
	return true
}

func bracket_token_getter(tokens []Token, bracket string) ([]Token, int64) {
	res := make([]Token, 0)
	brackets := 0
	close_bracket := ""
	if bracket == "{" {
		close_bracket = "}"
	}
	if bracket == "(" {
		close_bracket = ")"
	}
	if bracket == "[" {
		close_bracket = "]"
	}
	for _, token := range tokens {
		if token.Type == "bracket_open" && token.string_value == bracket {
			brackets += 1
		}
		if token.Type == "bracket_close" && token.string_value == close_bracket {
			brackets -= 1
		}
		res = append(res, token)
		if brackets == 0 {
			break
		}
	}
	return res, int64(brackets)
}

func get_current_statement_tokens(tokens []Token) ([]Token, int64) {
	res := make([]Token, 0)
	for _, token := range tokens {
		if token.Type == "EOS" {
			return res, 0
		}
		res = append(res, token)
	}
	return res, 1
}

func variable_doesnot_exist(symbol_table Symbol_Table, variable string) bool {
	for i := 0; i < len(symbol_table.functions); i++ {
		if symbol_table.functions[i].name == variable {
			return false
		}
	}
	for i := 0; i < len(symbol_table.structs); i++ {
		if symbol_table.structs[i].name == variable {
			return false
		}
	}
	for i := 0; i < len(symbol_table.variables); i++ {
		if symbol_table.variables[i].name == variable {
			return false
		}
	}
	return true
}

func type_token_to_string_array(type_token Token) []string {
	res := make([]string, 0)
	for _, tokenx := range type_token.keys {
		res = append(res, tokenx.string_value)
	}
	return res
}

func valid_type(Type []string, symbol_table Symbol_Table) bool {
	valid_final_types := []string{"num", "string"}
	if len(Type)%2 == 0 {
		return false
	}
	for i := 0; i < ((len(Type)-1)/2)-1; i++ {
		if Type[i] == "{" && Type[len(Type)-1] == "}" {
			continue
		}
		if Type[i] == "[" && Type[len(Type)-1] == "]" {
			continue
		}
		return false
	}
	for _, Struct := range symbol_table.structs {
		valid_final_types = append(valid_final_types, Struct.name)
	}
	if str_index_in_arr(Type[(len(Type)-1)/2], valid_final_types) != -1 {
		return true
	}
	return false
}

func does_function_exist(function_name string, symbol_table Symbol_Table) bool {
	for i := 0; i < len(symbol_table.functions); i++ {
		if symbol_table.functions[i].name == function_name {
			return true
		}
	}
	return false
}

func function_index_in_symbol_table(function_name string, symbol_table Symbol_Table) int {
	for i := 0; i < len(symbol_table.functions); i++ {
		if symbol_table.functions[i].name == function_name {
			return i
		}
	}
	return -1
}

func variable_index_in_symbol_table(variable_name string, symbol_table Symbol_Table) int {
	for i := 0; i < len(symbol_table.variables); i++ {
		if symbol_table.variables[i].name == variable_name {
			return i
		}
	}
	return -1
}

func struct_index_in_symbol_table(struct_name string, symbol_table Symbol_Table) int {
	if symbol_table.struct_mapping[struct_name]!="" {
		struct_name=symbol_table.struct_mapping[struct_name]
	}
	for i := 0; i < len(symbol_table.structs); i++ {
		if symbol_table.structs[i].name == struct_name {
			return i
		}
	}
	return -1
}

func calculate_i_skip(code []Token) int {
	total := 0
	for i := 0; i < len(code); i++ {
		if code[i].Type == "branch" || code[i].Type == "while" {
			total += 4
			total += calculate_i_skip(code[i].children[1].children)
		} else {
			total += 1
		}
	}
	return total
}

func branch_parser(code_original []Token) (string, []Token) {
	code:=make([]Token, 0)
	for _,code_obj:=range code_original {
		code = append(code, clone_token(code_obj))
	}
	parsed_tokens := make([]Token, 0)
	for i := 0; i < len(code); i++ {
		if code[i].Type == "sys" && code[i].string_value == "if" && (len(code)-i) >= 4 {
			if code[i+1].Type != "expression" || code[i+2].Type != "bracket_open" || code[i+2].string_value != "{" {
				return "invalid expression", make([]Token, 0)
			}
			tokens, err := bracket_token_getter(code[i+2:], "{")
			if err != 0 {
				return "tokeniser_bracket_error", make([]Token, 0)
			}
			var errstring string
			errstring, tokens = branch_parser(tokens[1 : len(tokens)-1])
			if errstring != "" {
				return errstring, make([]Token, 0)
			}
			parsed_tokens = append(parsed_tokens, Token{Type: "branch", children: []Token{code[i+1], Token{Type: "code", children: tokens}}})
			i += 3 + calculate_i_skip(tokens)
			continue
		}
		if code[i].Type == "sys" && code[i].string_value == "while" && (len(code)-i) >= 4 {
			if code[i+1].Type != "expression" || code[i+2].Type != "bracket_open" || code[i+2].string_value != "{" {
				return "invalid expression", make([]Token, 0)
			}
			tokens, err := bracket_token_getter(code[i+2:], "{")
			if err != 0 {
				return "tokeniser_bracket_error", make([]Token, 0)
			}
			var errstring string
			errstring, tokens = branch_parser(tokens[1 : len(tokens)-1])
			if errstring != "" {
				return errstring, make([]Token, 0)
			}
			parsed_tokens = append(parsed_tokens, Token{Type: "while", children: []Token{code[i+1], Token{Type: "code", children: tokens}}})
			i += 3 + calculate_i_skip(tokens)
			continue
		}
		parsed_tokens = append(parsed_tokens, code[i])
	}
	return "", parsed_tokens
}

func type_struct_translator(Type []string, symbol_table Symbol_Table) []string {
	for _, struct_ := range symbol_table.structs {
		if struct_.name == Type[len(Type)/2] {
			Type[len(Type)/2] = symbol_table.current_file + "-" + Type[len(Type)/2]
		}
	}
	return Type
}

func pre_parser(symbol_table Symbol_Table, code_original []Token, depth int) (string, Symbol_Table) {
	code:=make([]Token, 0)
	for _,code_obj:=range code_original {
		code = append(code, clone_token(code_obj))
	}
	result := ""
	for i := 0; i < len(code); i++ {
		// fmt.Println(code[i])
		if depth == 0 {
			if code[i].Type == "sys" && code[i].string_value == "struct" && len(code) > i+2 && code[i+1].Type == "variable" && valid_var_name(code[i+1].string_value) && code[i+2].Type == "bracket_open" && code[i+2].string_value == "{" {
				if !variable_doesnot_exist(symbol_table, code[i+1].string_value) {
					return "error", symbol_table
				}
				i++
				i++
				tokens, err := bracket_token_getter(code[i:], code[i].string_value)
				if err != 0 {
					return "error", symbol_table
				}
				tokens = tokens[1 : len(tokens)-1]
				if len(tokens)%2 != 0 {
					return "error", symbol_table
				}
				Struct_Variables := make(map[string][]string)
				for index := 0; int64(index) < int64(len(tokens))-1; index += 2 {
					variable_type := type_token_to_string_array(tokens[index+1])
					if tokens[index].Type == "variable" && valid_var_name(tokens[index].string_value) && valid_type(variable_type, symbol_table) {
						Struct_Variables[tokens[index].string_value] = variable_type
					} else {
						return "struct_error", symbol_table
					}
				}
				data_index := str_index_in_arr(code[i-1].string_value, symbol_table.data)
				if data_index == -1 {
					symbol_table.data = append(symbol_table.data, code[i-1].string_value)
					data_index = len(symbol_table.data) - 1
				}
				fields_converted := make([]string, 0)
				for key, val := range Struct_Variables {
					fields_converted = append(fields_converted, key+"->"+strings.Join(string_array_types_to_vitality_types(val, symbol_table), ","))
				}
				data_index = str_index_in_arr(strings.Join(fields_converted, ";"), symbol_table.data)
				if data_index == -1 {
					symbol_table.data = append(symbol_table.data, strings.Join(fields_converted, ";"))
					data_index = len(symbol_table.data) - 1
				}
				symbol_table.struct_registration = append(symbol_table.struct_registration, []string{"register_struct", symbol_table.current_file + "-" + code[i-1].string_value, strconv.FormatInt(int64(data_index), 10)})
				symbol_table.structs = append(symbol_table.structs, Struct{name: code[i-1].string_value, fields: Struct_Variables})
				symbol_table.struct_mapping[symbol_table.current_file+"-"+code[i-1].string_value]=code[i-1].string_value
				i += len(tokens) + 1
				continue
			}
			if code[i].Type == "sys" && code[i].string_value == "function" && len(code) > i+1 && code[i+1].Type == "funcall" && variable_doesnot_exist(symbol_table, code[i+1].children[0].string_value) {
				function_name := symbol_table.current_file + "-" + code[i+1].children[0].string_value
				function_arguments := make(map[string][]string)
				function_argument_keys := make([]string, 0)
				if len(code[i+1].children[1].children) != 0 && ((len(code[i+1].children[1].children) < 2) || (len(code[i+1].children[1].children) != 2 && len(code[i+1].children[1].children)%3 != 2)) {
					return "function_error", symbol_table
				}
				if !valid_var_name(code[i+1].children[0].string_value) {
					return "function_error_identifier", symbol_table
				}
				i++
				i++
				for index := 0; index < len(code[i-1].children[1].children); index += 3 {
					function_arguments[code[i-1].children[1].children[index].string_value] = type_token_to_string_array(code[i-1].children[1].children[index+1])
					function_argument_keys = append(function_argument_keys,code[i-1].children[1].children[index].string_value)
					if !valid_type(type_token_to_string_array(code[i-1].children[1].children[index+1]), symbol_table) {
						return "invalid function argument definition", symbol_table
					}
				}
				function_type := make([]string, 0)
				if code[i].Type == "type" {
					function_type = type_token_to_string_array(code[i])
					if !valid_type(function_type, symbol_table) {
						return "function_return_type_is_invalid", symbol_table
					}
				}
				if code[i].Type == "type" {
					i++
				}
				tokens, err := bracket_token_getter(code[i:], code[i].string_value)
				if err != 0 {
					return "error", symbol_table
				}
				// this crashes sometimes
				to_ignore := len(tokens[1 : len(tokens)-1])
				err_, tokens := branch_parser(tokens[1 : len(tokens)-1])
				if err_ != "" {
					return "error", symbol_table
				}
				symbol_table.functions = append(symbol_table.functions, Function{args: function_arguments, name: function_name, Type: function_type, Code: tokens, args_keys: function_argument_keys})
				symbol_table.data = append(symbol_table.data, function_name)
				i += to_ignore + 1
				continue
			}
			if code[i].Type == "sys" && code[i].string_value == "var" && (len(code)-i) >= 4 && code[i+1].Type == "variable" && valid_var_name(code[i+1].string_value) && valid_type(type_token_to_string_array(code[i+2]), symbol_table) && code[i+3].Type == "EOS" && variable_doesnot_exist(symbol_table, code[i+1].string_value) {
				symbol_table.variables = append(symbol_table.variables, Variable{name: code[i+1].string_value, Type: type_token_to_string_array(code[i+2])})
				variable_type := type_token_to_string_array(code[i+2])
				if variable_type[0] == "string" {
					string_index := str_index_in_arr("", symbol_table.data)
					if string_index == -1 {
						symbol_table.data = append(symbol_table.data, "")
						string_index = len(symbol_table.data) - 1
					}
					symbol_table.global_variables = append(symbol_table.global_variables, []string{"str.set", symbol_table.current_file + "-" + code[i+1].string_value, strconv.FormatInt(int64(string_index), 10)})
				} else if variable_type[0] == "num" {
					symbol_table.global_variables = append(symbol_table.global_variables, []string{"set", symbol_table.current_file + "-" + code[i+1].string_value, "0"})
				} else if variable_type[0] == "[" {
					new_variable_type := make([]string, 0)
					for _, vb := range variable_type {
						new_variable_type = append(new_variable_type, vb)
					}
					data_index := str_index_in_arr(strings.Join(string_array_types_to_vitality_types(variable_type, symbol_table)[1:], ","), symbol_table.data)
					if data_index == -1 {
						symbol_table.data = append(symbol_table.data, strings.Join(string_array_types_to_vitality_types(variable_type, symbol_table)[1:], ","))
						data_index = len(symbol_table.data) - 1
					}
					symbol_table.global_variables = append(symbol_table.global_variables, []string{"arr.init", symbol_table.current_file + "-" + code[i+1].string_value, strconv.FormatInt(int64(data_index), 10)})
				} else if variable_type[0] == "{" {
					data_index := str_index_in_arr(strings.Join(string_array_types_to_vitality_types(variable_type, symbol_table)[1:], ","), symbol_table.data)
					if data_index == -1 {
						symbol_table.data = append(symbol_table.data, strings.Join(string_array_types_to_vitality_types(variable_type, symbol_table)[1:], ","))
						data_index = len(symbol_table.data) - 1
					}
				} else if len(variable_type) == 1 {
					if !strings.Contains(variable_type[0], "-") {
						variable_type = []string{symbol_table.current_file + "-" + variable_type[0]}
					}
					if struct_index_in_symbol_table(variable_type[0], symbol_table) == -1 {
						return "Invalid variable initialisation", symbol_table
					}
					symbol_table.data = append(symbol_table.data, variable_type[0])
					symbol_table.global_variables = append(symbol_table.global_variables, []string{"struct.init", symbol_table.current_file + "-" + code[i+1].string_value, strconv.FormatInt(int64(str_index_in_arr(variable_type[0], symbol_table.data)), 10)})
				}
				// symbol_table.operations[function_name]=append(symbol_table.operations[function_name], )
				// you need to add the compiler
				i += 3
				continue
			}
			if code[i].Type == "sys" && code[i].string_value == "import" && (len(code)-i) >= 4 && code[i+1].Type == "string" && code[i+2].Type == "sys" && code[i+2].string_value == "as" && code[i+3].Type == "variable" && valid_var_name(code[i+3].string_value) && code[i+4].Type == "EOS" && !strings.Contains(code[i+1].string_value, "-") {
				code_in_bytes, err := os.ReadFile(code[i+1].string_value)
				if err != nil {
					return err.Error(), symbol_table
				}
				tokens_, _ := tokens_parser(tokensier(string(code_in_bytes), true), true)
				tokens_, _ = token_grouper(tokens_, true)
				new_symbol_table := Symbol_Table{
					functions:           symbol_table.functions,
					used_variables:      symbol_table.used_variables,
					structs:             make([]Struct, 0),
					variables:           make([]Variable, 0),
					data:                make([]string, 0),
					operations:          make(map[string][][]string),
					variable_mapping:    make(map[string]string),
					current_scope:       Scope{},
					files:               symbol_table.files,
					struct_registration: make([][]string, 0),
					imported_libraries:  make(map[string]string),
					current_file:        code[i+1].string_value,
					global_variables:    symbol_table.global_variables,
					struct_mapping:      make(map[string]string),
				}
				new_file_name := code[i+3].string_value
				debug_keys := make([]string, 0)
				for name, _ := range symbol_table.files {
					debug_keys = append(debug_keys, name)
				}

				if symbol_table.files[new_file_name].current_file != "" && symbol_table.files[new_file_name].finished_importing == false {
					return "circular_import", symbol_table
				}
				if str_index_in_arr(new_file_name, debug_keys) == -1 {
					new_symbol_table.files[new_file_name] = Symbol_Table{current_file: new_file_name}
					new_symbol_table.current_file = new_file_name
					err_, new_file := build(new_symbol_table, tokens_, 1)
					if len(err_) >= 6 && strings.HasPrefix(err_, "Error") {
						return err_, symbol_table
					}
					symbol_table.files[new_file_name] = new_file
					new_symbol_table_data := new_file.data
					new_symbol_table_data = append(new_symbol_table_data, symbol_table.data...)
					symbol_table.data = new_symbol_table_data
					for _, function := range new_file.functions {
						symbol_table.functions = append(symbol_table.functions, function)
					}
					symbol_table.imported_libraries[code[i+3].string_value] = new_file_name
					for function_name_, function := range new_symbol_table.operations {
						symbol_table.operations[function_name_] = function
					}
					symbol_table.global_variables = append(symbol_table.global_variables, new_file.global_variables...)
					symbol_table.struct_registration = append(symbol_table.struct_registration, new_file.struct_registration...)
					for _, struct_ := range new_file.structs {
						struct_.name = new_file.current_file + "-" + struct_.name
						symbol_table.structs = append(symbol_table.structs, struct_)
					}
					for key,val:=range new_file.struct_mapping {
						symbol_table.struct_mapping[key]=new_file.current_file + "-" + val
					}
				}
				i += 4
				continue
			}
		}
		fmt.Println(code[i])
		return "error_token_not_parsed", symbol_table
	}
	return result, symbol_table
}

func is_valid_statement(symbol_table Symbol_Table, code_original []Token) bool {
	code:=make([]Token, 0)
	for _,code_obj:=range code_original {
		code = append(code, clone_token(code_obj))
	}
	if len(code)%2 != 1 {
		return false
	}
	for i, token := range code {
		if token.Type == "expression" {
			if !is_valid_statement(symbol_table, token.children) {
				return false
			}
		}
		if i%2 == 0 {
			if str_index_in_arr(token.Type, []string{"lookup", "variable", "expression", "string", "num", "funcall", "nested_tokens", "array"}) == -1 {
				return false
			}
		}
		if i%2 == 1 {
			if token.Type != "operator" && str_index_in_arr(token.string_value, operators) == -1 {
				return false
			}
			if str_index_in_arr(token.string_value, []string{"|", "&", "!"}) != -1 {
				return false
			}
		}
	}
	return true
}

func are_function_arguments_valid(argument_expression_original Token, function Function, symbol_table Symbol_Table) bool {
	argument_expression:=clone_token(argument_expression_original)
	if len(argument_expression.children)%2 != 1 && len(argument_expression.children) != 0 {
		return false
	}
	arguments := make([]Token, 0)
	cache := make([]Token, 0)
	for i := 0; i < len(argument_expression.children); i++ {
		if argument_expression.children[i].Type == "comma" {
			if len(cache) == 0 {
				return false
			}
			arguments = append(arguments, Token{Type: "expression", children: cache})
			cache = make([]Token, 0)
		} else {
			cache = append(cache, argument_expression.children[i])
		}
	}
	if len(cache) != 0 {
		arguments = append(arguments, Token{Type: "expression", children: cache})
	}
	if len(arguments) != len(function.args) {
		return false
	}
	function_arguments := make([]string, 0)
	for _,key := range function.args_keys {
		function_arguments = append(function_arguments, key)
	}
	for i := 0; i < len(arguments); i++ {
		if !string_arr_compare(evaluate_type(symbol_table, arguments[i].children, 0), function.args[function_arguments[i]]) {
			return false
		}
	}
	return true
}

func are_array_arguments_valid(argument_expression_original Token, Type []string, symbol_table Symbol_Table) bool {
	argument_expression:=clone_token(argument_expression_original)
	if len(argument_expression.children) == 0 {
		return true
	}
	if len(argument_expression.children)%2 != 1 {
		return false
	}
	arguments := make([]Token, 0)
	cache := make([]Token, 0)
	for i := 0; i < len(argument_expression.children); i++ {
		if argument_expression.children[i].Type == "comma" {
			if len(cache) == 0 {
				return false
			}
			arguments = append(arguments, Token{Type: "expression", children: cache})
			cache = make([]Token, 0)
		} else {
			cache = append(cache, argument_expression.children[i])
		}
	}
	if len(cache) != 0 {
		arguments = append(arguments, Token{Type: "expression", children: cache})
	}
	for i := 0; i < len(arguments); i++ {
		if len(Type) < 3 {
			return false
		}
		if !string_arr_compare(evaluate_type(symbol_table, arguments[i].children, 0), Type[1:len(Type)-1]) {
			return false
		}
	}
	return true
}

func nested_puller(code_original []Token) []Token {
	code:=make([]Token, 0)
	for _,code_obj:=range code_original {
		code = append(code, clone_token(code_obj))
	}
	grouped_tokens := make([]Token, 0)
	for i := 0; i < len(code); i++ {
		if code[i].Type == "nested_tokens" {
			new_children := make([]Token, 0)
			for j := 0; j < len(code[i].children); j++ {
				if code[i].children[j].Type == "nested_tokens" {
					new_children = append(new_children, nested_puller(code[i].children[j].children)...)
					continue
				}
				new_children = append(new_children, code[i].children[j])
			}
			code[i].children = new_children
			grouped_tokens = append(grouped_tokens, code[i])
			continue
		}
		code[i].children = nested_puller(code[i].children)
		grouped_tokens = append(grouped_tokens, code[i])
	}
	return grouped_tokens
}

func clone_token(a Token) Token {
	new_children:=make([]Token, 0)
	for _,child:=range a.children {
		new_children = append(new_children, clone_token(child))
	}
	new_keys:=make([]Token, 0)
	for _,key:=range a.keys {
		new_keys = append(new_keys, clone_token(key))
	}
	return Token{Type: strings.Clone(a.Type), num_value: a.num_value, string_value: strings.Clone(a.string_value), children: new_children, keys: new_keys}
}

func evaluate_type(symbol_table Symbol_Table, code_original []Token, depth int) []string {
	if len(code_original) == 0 {
		return make([]string, 0)
	}
	code:=make([]Token, 0)
	for _,code_obj:=range code_original {
		code = append(code, clone_token(code_obj))
	}
	current_type := make([]string, 0)
	if is_valid_statement(symbol_table, code) || depth != 0 {
		if len(code) == 1 {
			if code[0].Type == "string" {
				return []string{"string"}
			}
			if code[0].Type == "num" {
				return []string{"num"}
			}
			if code[0].Type == "expression" {
				return evaluate_type(symbol_table, code[0].children, 0)
			}
			if code[0].Type == "variable" {
				variable_index := variable_index_in_symbol_table(code[0].string_value, symbol_table)
				if variable_index == -1 && symbol_table.variable_mapping[code[0].string_value]=="" {
					return make([]string, 0)
				}
				if variable_index != -1  {
					return symbol_table.variables[variable_index].Type
				}
				if symbol_table.variable_mapping[code[0].string_value]!="" {
					return symbol_table.variables[variable_index_in_symbol_table(symbol_table.variable_mapping[code[0].string_value], symbol_table)].Type
				}
				return symbol_table.variables[variable_index].Type
			}
			if code[0].Type == "array" {
				if !valid_type(type_token_to_string_array(code[0].children[0]), symbol_table) {
					return make([]string, 0)
				}
				if are_array_arguments_valid(code[0].children[1], type_token_to_string_array(code[0].children[0]), symbol_table) {
					return type_token_to_string_array(code[0].children[0])
				}
				return make([]string, 0)
			}
			if code[0].Type == "lookup" {
				parent_type := evaluate_type(symbol_table, []Token{code[0].children[0].children[0]}, 0)
				if len(parent_type) == 0 {
					return make([]string, 0)
				}
				if str_index_in_arr(parent_type[0], []string{"[", "{"}) == -1 {
					return make([]string, 0)
				}
				lookup_type := evaluate_type(symbol_table, code[0].children[1].children, 0)
				if len(lookup_type) == 0 {
					return make([]string, 0)
				}
				if parent_type[0] == "[" && string_arr_compare(lookup_type, []string{"num"}) {
					return parent_type[1 : len(parent_type)-1]
				}
				if parent_type[0] == "{" && string_arr_compare(lookup_type, []string{"string"}) {
					return parent_type[1 : len(parent_type)-1]
				}
				return make([]string, 0)
			}
			if code[0].Type == "nested_tokens" {
				parent_type := make([]string, 0)
				if depth == 0 {
					parent_type = evaluate_type(symbol_table, []Token{code[0].children[0]}, 0)
					if len(parent_type) == 1 && strings.Contains(parent_type[0], "-struct-") && strings.Contains(parent_type[0], symbol_table.current_file) {
						parent_type = []string{strings.Split(parent_type[0], "-struct-")[1]}
					}
					if len(parent_type) == 0 {
						return make([]string, 0)
					}
				} else {
					parent_type = []string{code[0].children[0].string_value}
				}
				if len(parent_type) == 1 && struct_index_in_symbol_table(parent_type[0], symbol_table) != -1 {
					struct_index := struct_index_in_symbol_table(parent_type[0], symbol_table)
					res := symbol_table.structs[struct_index].fields[code[0].children[1].string_value]
					if len(res) == 0 {
						return make([]string, 0)
					}
					if len(code[0].children) > 2 {
						if struct_index_in_symbol_table(res[0], symbol_table) == -1 {
							return make([]string, 0)
						}
						code[0].children[1] = Token{string_value: res[0]}
						code_ := code
						code_[0].children = code[0].children[1:]
						return evaluate_type(symbol_table, code_, 1)
					} else {
						return res
					}
					return make([]string, 0)
				}
				return make([]string, 0)
			}
			if code[0].Type == "funcall" {
				if code[0].children[0].Type == "variable" {
					function_index := function_index_in_symbol_table(symbol_table.current_file+"-"+code[0].children[0].string_value, symbol_table)
					if function_index == -1 {
						return make([]string, 0)
					}
					if are_function_arguments_valid(code[0].children[1], symbol_table.functions[function_index], symbol_table) {
						return symbol_table.functions[function_index].Type
					}
				}
				if code[0].children[0].Type == "nested_tokens" {
					code_ := code[0].children[0]
					imported_libraries := make([]string, 0)
					for key := range symbol_table.imported_libraries {
						imported_libraries = append(imported_libraries, key)
					}
					if str_index_in_arr(code_.children[0].string_value, imported_libraries) == -1 {
						return make([]string, 0)
					}
					library_symbol_table := symbol_table.files[symbol_table.imported_libraries[code_.children[0].string_value]]
					function_name := library_symbol_table.current_file + "-" + code_.children[1].string_value
					function_index := function_index_in_symbol_table(function_name, library_symbol_table)
					if function_index == -1 {
						return make([]string, 0)
					}
					function := library_symbol_table.functions[function_index]

					if are_function_arguments_valid(code[0].children[1], function, symbol_table) {
						return type_struct_translator(function.Type, library_symbol_table)
					}
					// function_name:=code[0].children[0].children[len(code[0].children[0].children)-1].string_value
					// arguments:=code[0].children[1]
					// add import function calls
				}
				return make([]string, 0)
			}
			return make([]string, 0)
		}
		if len(code) > 1 {
			for i := 0; i < len(code); i++ {
				if i+2 >= len(code) {
					return current_type
				}
				lhs := current_type
				if string_arr_compare(make([]string, 0), current_type) {
					lhs = evaluate_type(symbol_table, []Token{code[i]}, 0)
				}
				rhs := evaluate_type(symbol_table, []Token{code[i+2]}, 0)
				if string_arr_compare(lhs, make([]string, 0)) {
					return make([]string, 0)
				}
				if string_arr_compare(rhs, make([]string, 0)) {
					return make([]string, 0)
				}
				i++
				switch operator := code[i]; operator.string_value {
				case "+", "*":
					if operator.string_value == "*" {
						if (string_arr_compare(rhs, []string{"string"}) && string_arr_compare(lhs, []string{"num"})) || (string_arr_compare(rhs, []string{"num"}) && string_arr_compare(lhs, []string{"string"})) || (string_arr_compare(rhs, []string{"num"}) && string_arr_compare(lhs, []string{"num"})) {
							current_type = lhs
							continue
						}
					}
					if operator.string_value == "+" && string_arr_compare(rhs, lhs) && (string_arr_compare(rhs, []string{"num"}) || string_arr_compare(rhs, []string{"string"})) {
						current_type = rhs
						continue
					}
					return make([]string, 0)
				case "-", "/", "//", "^", "**", "%", "&&", "||":
					if string_arr_compare(lhs, rhs) && string_arr_compare(lhs, []string{"num"}) {
						current_type = rhs
						continue
					}
					return make([]string, 0)
				case "==", "!=", ">", "<":
					if !(string_arr_compare(lhs, rhs) && (string_arr_compare(lhs, []string{"num"}) || string_arr_compare(lhs, []string{"string"}))) {
						return make([]string, 0)
					}
					return []string{"num"}
				}
			}
		}
	} else {
		fmt.Println("invalid statement", code, len(code))
		return make([]string, 0)
	}
	return current_type
}

func string_array_types_to_vitality_types(typesx []string, symbol_table Symbol_Table) []string {
	if len(typesx) == 0 {
		return make([]string, 0)
	}
	new_types := make([]string, 0)
	for _, str := range typesx {
		new_types = append(new_types, str)
	}
	types := type_struct_translator(new_types, symbol_table)
	output := make([]string, 0)
	if types[0] == "[" {
		output = append(output, "arr")
		for _, type_string := range string_array_types_to_vitality_types(types[1:len(types)-1], symbol_table) {
			output = append(output, type_string)
		}
	} else if types[0] == "{" {
		output = append(output, "dict")
		for _, type_string := range string_array_types_to_vitality_types(types[1:len(types)-1], symbol_table) {
			output = append(output, type_string)
		}
	} else {
		output = append(output, types[0])
	}
	return output
}

func get_variable(Type []string, symbol_table Symbol_Table) (string, Symbol_Table) {
	if symbol_table.variable_mapping[strings.Join(Type, "")] == "" {
		unique_combinations := len(symbol_table.variable_mapping)
		power_of_52 := 0
		for true {
			if math.Pow(52, float64(power_of_52)) > float64(unique_combinations) {
				break
			} else {
				power_of_52 += 1
			}
		}
		if power_of_52 != 0 {
			power_of_52 = power_of_52 - 1
		}
		output_variable := strings.Repeat("Z", power_of_52)
		output_variable += string("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"[int(math.Mod(float64(unique_combinations), 52))])
		symbol_table.variable_mapping[strings.Join(Type, "")] = output_variable
	}
	variable_count := 0
	for true {
		found := false
		for _, i := range symbol_table.used_variables[symbol_table.variable_mapping[strings.Join(Type, "")]] {
			if i == variable_count {
				found = true
				break
			}
		}
		if found {
			variable_count += 1
			continue
		} else {
			symbol_table.used_variables[symbol_table.variable_mapping[strings.Join(Type, "")]] = append(symbol_table.used_variables[symbol_table.variable_mapping[strings.Join(Type, "")]], variable_count)
			return symbol_table.variable_mapping[strings.Join(Type, "")] + "_" + strconv.FormatInt(int64(variable_count), 10), symbol_table
		}
	}
	return "", symbol_table
}

func free_variable(name string, symbol_table Symbol_Table) Symbol_Table {
	new_used_variable := make([]int, 0)
	variable_index_int64, _ := strconv.ParseInt(strings.Split(name, "_")[1], 10, 64)
	variable_index := int(variable_index_int64)
	for _, i := range symbol_table.used_variables[strings.Split(name, "_")[0]] {
		if i != variable_index {
			new_used_variable = append(new_used_variable, i)
		}
	}
	symbol_table.used_variables[strings.Split(name, "_")[0]] = new_used_variable
	return symbol_table
}

func var_init(variable_type []string, variable_name string, symbol_table Symbol_Table, function_name string, add_to_stack_init bool, strict_name bool) Symbol_Table {
	if variable_type[0] == "string" {
		string_index := str_index_in_arr("", symbol_table.data)
		if string_index == -1 {
			symbol_table.data = append(symbol_table.data, "")
			string_index = len(symbol_table.data) - 1
		}
		symbol_table.variables = append(symbol_table.variables, Variable{name: variable_name, Type: variable_type})
		if add_to_stack_init {
			symbol_table.global_variables = append(symbol_table.global_variables, []string{"str.set", variable_name, strconv.FormatInt(int64(string_index), 10)})
		} else {
			symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"str.set", variable_name, strconv.FormatInt(int64(string_index), 10)})
		}
	} else if variable_type[0] == "num" {
		symbol_table.variables = append(symbol_table.variables, Variable{name: variable_name, Type: variable_type})
		if add_to_stack_init {
			symbol_table.global_variables = append(symbol_table.global_variables, []string{"set", variable_name, "0"})
		} else {
			symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"set", variable_name, "0"})
		}
	} else if variable_type[0] == "[" {
		new_variable_type := make([]string, 0)
		for _, vb := range variable_type {
			new_variable_type = append(new_variable_type, vb)
		}
		data_index := str_index_in_arr(strings.Join(string_array_types_to_vitality_types(variable_type, symbol_table)[1:], ","), symbol_table.data)
		if data_index == -1 {
			symbol_table.data = append(symbol_table.data, strings.Join(string_array_types_to_vitality_types(variable_type, symbol_table)[1:], ","))
			data_index = len(symbol_table.data) - 1
		}
		symbol_table.variables = append(symbol_table.variables, Variable{name: variable_name, Type: new_variable_type})
		if add_to_stack_init {
			symbol_table.global_variables = append(symbol_table.global_variables, []string{"arr.init", variable_name, strconv.FormatInt(int64(data_index), 10)})
		} else {
			symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"arr.init", variable_name, strconv.FormatInt(int64(data_index), 10)})
		}
	} else if variable_type[0] == "{" {
		data_index := str_index_in_arr(strings.Join(string_array_types_to_vitality_types(variable_type, symbol_table)[1:], ","), symbol_table.data)
		if data_index == -1 {
			symbol_table.data = append(symbol_table.data, strings.Join(string_array_types_to_vitality_types(variable_type, symbol_table)[1:], ","))
			data_index = len(symbol_table.data) - 1
		}
		symbol_table.variables = append(symbol_table.variables, Variable{name: variable_name, Type: variable_type})
		// you need to add this
	} else if len(variable_type) == 1 {
		if !strings.Contains(variable_type[0], "-") {
			variable_type = []string{symbol_table.current_file + "-" + variable_type[0]}
		}
		if struct_index_in_symbol_table(variable_type[0], symbol_table) == -1 {
			return symbol_table
		}
		symbol_table.variables = append(symbol_table.variables, Variable{name: variable_name, Type: variable_type})
		
		if str_index_in_arr(variable_type[0], symbol_table.data) == -1 {
			symbol_table.data = append(symbol_table.data, variable_type[0])
		}
		var_name:=symbol_table.current_file + "-" + variable_name
		if strict_name {
			var_name=variable_name
		}
		if add_to_stack_init {
			symbol_table.global_variables= append(symbol_table.global_variables, []string{"struct.init", var_name, strconv.FormatInt(int64(str_index_in_arr(variable_type[0], symbol_table.data)), 10)})
		} else {
			symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"struct.init", var_name, strconv.FormatInt(int64(str_index_in_arr(variable_type[0], symbol_table.data)), 10)})
		}
	}
	return symbol_table
}

func resolve_function_name(code []Token, symbol_table Symbol_Table) string {
	path:=make([]string, 0)
	for _,each:=range code {
		path = append(path, each.string_value)
	}
	if len(path)==2 {
		return path[0]+"-"+path[1]
	}
	return ""
}

func expression_solver(tokens_original []Token, function_name string, symbol_table Symbol_Table, nested_nested_tokens bool) (string, []string, Symbol_Table) {
	tokens:=make([]Token, 0)
	for _,token_obj:=range tokens_original {
		tokens = append(tokens, clone_token(token_obj))
	}
	resultant_variable := ""
	used_variables := make([]string, 0)
	if len(tokens) == 1 {
		if tokens[0].Type == "num" {
			resultant_variable, symbol_table = get_variable([]string{"num"}, symbol_table)
			symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"set", resultant_variable, strings.TrimRight(strings.TrimRight(strconv.FormatFloat(float64(tokens[0].num_value), 'f', 8, 64), "0"), ".")})
			used_variables = append(used_variables, resultant_variable)
			return resultant_variable, used_variables, symbol_table
		}
		if tokens[0].Type == "string" {
			resultant_variable, symbol_table = get_variable([]string{"string"}, symbol_table)
			symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"str.set", resultant_variable, strconv.FormatInt(int64(str_index_in_arr(tokens[0].string_value, symbol_table.data)), 10)})
			used_variables = append(used_variables, resultant_variable)
			return resultant_variable, used_variables, symbol_table
		}
		if tokens[0].Type == "variable" {
			if symbol_table.variable_mapping[tokens[0].string_value] != "" {
				return symbol_table.variable_mapping[tokens[0].string_value], used_variables, symbol_table
			}
			return symbol_table.current_file + "-" + tokens[0].string_value, used_variables, symbol_table
		}
		if tokens[0].Type == "expression" {
			return expression_solver(tokens[0].children, function_name, symbol_table, false)
		}
		if tokens[0].Type == "array" {
			array_variable, new_symbol_table := get_variable(type_token_to_string_array(tokens[0].children[0]), symbol_table)
			symbol_table = new_symbol_table
			used_variables = append(used_variables, array_variable)
			type_index := str_index_in_arr(strings.Join(string_array_types_to_vitality_types(type_token_to_string_array(tokens[0].children[0]), symbol_table)[1:], ","), symbol_table.data)
			if type_index == -1 {
				symbol_table.data = append(symbol_table.data, strings.Join(string_array_types_to_vitality_types(type_token_to_string_array(tokens[0].children[0]), symbol_table)[1:], ","))
				type_index = len(symbol_table.data) - 1
			}
			symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"arr.init", array_variable, strconv.FormatInt(int64(type_index), 10)})
			arguments := make([][]Token, 0)
			cache := make([]Token, 0)
			for i := 0; i < len(tokens[0].children[1].children); i++ {
				if tokens[0].children[1].children[i].Type == "comma" {
					arguments = append(arguments, cache)
					cache = make([]Token, 0)
				} else {
					cache = append(cache, tokens[0].children[1].children[i])
				}
			}
			if len(cache) != 0 {
				arguments = append(arguments, cache)
			}
			array_children := make([]string, 0)
			for i := 0; i < len(arguments); i++ {
				argument := arguments[i]
				new_resultant_variable, new_used_variables, new_symbol_table := expression_solver(argument, function_name, symbol_table, false)
				used_variables = append(used_variables, new_used_variables...)
				array_children = append(array_children, new_resultant_variable)
				symbol_table = new_symbol_table
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"arr.push", array_variable, new_resultant_variable})
			}
			return array_variable, used_variables, symbol_table
		}
		if tokens[0].Type == "nested_tokens" {
			if len(tokens[0].children) == 2 {
				variable_type := evaluate_type(symbol_table, []Token{tokens[0].children[0]}, 0)
				variable_struct := symbol_table.structs[struct_index_in_symbol_table(variable_type[0], symbol_table)]
				if len(variable_struct.fields[tokens[0].children[1].string_value]) != 0 {
					if nested_nested_tokens || tokens[0].children[0].Type=="variable" {
						new_variable, new_symbol_table := get_variable(variable_struct.fields[tokens[0].children[1].string_value], symbol_table)
						symbol_table = new_symbol_table
						new_symbol_table = var_init(variable_struct.fields[tokens[0].children[1].string_value], new_variable, symbol_table, function_name, false, false)
						symbol_table = new_symbol_table
						used_variables = append(used_variables, new_variable)
						field_index_in_struct:=str_index_in_arr(tokens[0].children[1].string_value, symbol_table.data)
						if field_index_in_struct==-1 {
							field_index_in_struct=len(symbol_table.data)
							symbol_table.data = append(symbol_table.data, tokens[0].children[1].string_value)
						}
						//variable_name:=tokens[0].children[1].string_value
						symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"struct.pull", symbol_table.current_file + "-" + tokens[0].children[0].string_value, strconv.FormatInt(int64(field_index_in_struct), 10), new_variable})
						return new_variable, used_variables, symbol_table
					} else {
						rhs_pretype:=evaluate_type(symbol_table, []Token{tokens[0].children[0]}, 0)
						rhs_pretype=symbol_table.structs[struct_index_in_symbol_table(rhs_pretype[0], symbol_table)].fields[tokens[0].children[1].string_value]
						new_variable, new_symbol_table := get_variable(rhs_pretype, symbol_table)
						symbol_table = new_symbol_table
						new_symbol_table = var_init(rhs_pretype, new_variable, symbol_table, function_name, false, false)
						symbol_table = new_symbol_table
						lhs, new_used_variables, new_symbol_table := expression_solver([]Token{tokens[0].children[0]}, function_name, symbol_table, false)
						used_variables = append(used_variables, new_used_variables...)
						symbol_table = new_symbol_table

						field_index_in_struct:=str_index_in_arr(tokens[0].children[1].string_value, symbol_table.data)
						if field_index_in_struct==-1 {
							field_index_in_struct=len(symbol_table.data)
							symbol_table.data = append(symbol_table.data, tokens[0].children[1].string_value)
						}

						symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"struct.pull", lhs, strconv.FormatInt(int64(field_index_in_struct), 10), new_variable})
						return new_variable, used_variables, symbol_table
					}
				}
			} else {
				new_children := []Token{}
				new_children = append(new_children, tokens[0].children[:2]...)
				new_token := Token{Type: "nested_tokens", children: new_children}
				resolved_type := evaluate_type(symbol_table, []Token{new_token}, 0)
				new_variable, new_symbol_table := get_variable(resolved_type, symbol_table)
				symbol_table = new_symbol_table
				new_symbol_table = var_init(resolved_type, new_variable, symbol_table, function_name, false, false)
				symbol_table = new_symbol_table
				used_variables = append(used_variables, new_variable)
				symbol_table.variables = append(symbol_table.variables, Variable{name: new_variable, Type: resolved_type})
				new_modified_variable:=new_children[1].string_value
				if str_index_in_arr(new_modified_variable, symbol_table.data)==-1 {
					symbol_table.data = append(symbol_table.data, new_modified_variable)
				}
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"struct.pull", symbol_table.current_file+"-"+new_children[0].string_value, strconv.FormatInt(int64(str_index_in_arr(new_modified_variable, symbol_table.data)), 10), symbol_table.current_file+"-"+new_variable})
				new_children = make([]Token, 0)
				new_children = append(new_children, Token{Type: "variable", string_value: new_variable})
				new_children = append(new_children, tokens[0].children[2:]...)
				new_token = Token{Type: tokens[0].Type, children: new_children}
				new_children[0].string_value = new_variable
				//symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"struct.pull", symbol_table.current_file + "-struct-" + tokens[0].children[0].string_value, strconv.FormatInt(int64(str_index_in_arr(tokens[0].children[1].string_value, symbol_table.data)), 10), new_variable})
				varying_value, new_used_variables, new_symbol_table := expression_solver([]Token{new_token}, function_name, symbol_table, true)
				used_variables = append(used_variables, new_used_variables...)
				symbol_table = new_symbol_table
				return varying_value, used_variables, symbol_table
			}

		}
		if tokens[0].Type=="funcall" {
			provided_arguments:=make([][]Token, 0)
			cache:=make([]Token, 0)
			for _,tkn:=range tokens[0].children[1].children {
				if tkn.Type=="comma" {
					provided_arguments = append(provided_arguments, cache)
					cache=make([]Token, 0)
				} else {
					cache = append(cache, tkn)
				}
			}
			if len(cache)!=0 {
				provided_arguments = append(provided_arguments, cache)
			}
			resolved_arguments:=make([]string, 0)
			for _,argument:=range provided_arguments {
				lhs, new_used_variables, new_symbol_table := expression_solver(argument, function_name, symbol_table, false)
				used_variables = append(used_variables, new_used_variables...)
				symbol_table = new_symbol_table
				resolved_arguments = append(resolved_arguments, lhs)
			}
			calling_function_name:=""
			if tokens[0].children[0].Type!="nested_tokens" {
				calling_function_name=symbol_table.current_file+"-"+tokens[0].children[0].string_value
			} else {
				calling_function_name=resolve_function_name(tokens[0].children[0].children, symbol_table)
			}
			function:=symbol_table.functions[function_index_in_symbol_table(calling_function_name, symbol_table)]
			argument_counter:=0
			symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"scope.new"})
			symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"set","return_to","0"})
			for _,arg:=range function.args_keys {
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"obj_copy_update_scope",calling_function_name+"-function_input:"+arg, resolved_arguments[argument_counter]})
				argument_counter+=1
			}
			calling_function_name_index:=str_index_in_arr(calling_function_name, symbol_table.data)
			if calling_function_name_index==-1 {
				calling_function_name_index=len(symbol_table.data)
				symbol_table.data = append(symbol_table.data, calling_function_name)
			}
			symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"current_line_number","return_to"})
			symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"set", "four", "4"})
			symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"add", "return_to", "four", "return_to"})
			symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"jump.def.always", strconv.FormatInt(int64(calling_function_name_index), 10)})
			symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"scope.exit"})
			return calling_function_name+"-return-variable", used_variables, symbol_table
		}
		// you need to add function calls and nested_tokens
	} else {
		lhs, new_used_variables, new_symbol_table := expression_solver([]Token{tokens[0]}, function_name, symbol_table, false)
		used_variables = append(used_variables, new_used_variables...)
		symbol_table = new_symbol_table
		lhs_type := evaluate_type(symbol_table, []Token{tokens[0]}, 0)
		for i := 0; i < len(tokens); i += 2 {
			if i+2 > len(tokens) {
				break
			}
			rhs, new_used_variables, new_symbol_table := expression_solver([]Token{tokens[i+2]}, function_name, symbol_table, true)
			used_variables = append(used_variables, new_used_variables...)
			symbol_table = new_symbol_table
			rhs_type := evaluate_type(symbol_table, []Token{tokens[i+2]}, 0)
			switch operator := tokens[i+1].string_value; operator {
			case "+":
				if lhs_type[0] == "num" && rhs_type[0] == "num" {
					new_variable, new_symbol_table := get_variable(lhs_type, symbol_table)
					symbol_table = new_symbol_table
					used_variables = append(used_variables, new_variable)
					symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"set", new_variable, "0"})
					symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"add", lhs, rhs, new_variable})
					lhs = new_variable
					lhs_type = rhs_type
				}
				if lhs_type[0] == "string" && rhs_type[0] == "string" {
					new_variable, new_symbol_table := get_variable(lhs_type, symbol_table)
					symbol_table = new_symbol_table
					used_variables = append(used_variables, new_variable)
					symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"str.set", new_variable, "0"})
					symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"str.add", lhs, rhs, new_variable})
					lhs = new_variable
					lhs_type = rhs_type
				}
			case "*":
				if (lhs_type[0] == "num" && rhs_type[0] == "string") || (rhs_type[0] == "num" && lhs_type[0] == "string") {
					var_1 := lhs
					var_2 := rhs
					if rhs_type[0] == "string" {
						var_1 = rhs
						var_2 = lhs
					}
					new_variable, new_symbol_table := get_variable([]string{"string"}, symbol_table)
					symbol_table = new_symbol_table
					used_variables = append(used_variables, new_variable)
					symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"str.set", new_variable, "0"})

					symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"str.mult", var_1, var_2, new_variable})
					lhs = new_variable
					lhs_type = []string{"string"}
				}
				if lhs_type[0] == "num" && rhs_type[0] == "num" {
					new_variable, new_symbol_table := get_variable(lhs_type, symbol_table)
					symbol_table = new_symbol_table
					used_variables = append(used_variables, new_variable)
					symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"set", new_variable, "0"})
					symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"mult", lhs, rhs, new_variable})
					lhs = new_variable
					lhs_type = rhs_type
				}
			case "-", "/", "//", "^", "**", "%", "&&", "||":
				operation := ""
				switch operator {
				case "-":
					operation = "sub"
				case "/":
					operation = "div"
				case "//":
					operation = "floor"
				case "^":
					operation = "xor"
				case "**":
					operation = "power"
				case "%":
					operation = "mod"
				case "&&":
					operation = "and"
				case "||":
					operation = "or"
				}
				new_variable, new_symbol_table := get_variable(lhs_type, symbol_table)
				symbol_table = new_symbol_table
				used_variables = append(used_variables, new_variable)
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"set", new_variable, "0"})
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{operation, lhs, rhs, new_variable})
				lhs = new_variable
				lhs_type = rhs_type
			case "==", "!=", ">", "<":
				operation := ""
				if lhs_type[0] == "string" {
					switch operator {
					case "==":
						operation = "str.equals"
					case "!=":
						operation = "str.equals"
					}
					if operator == "==" {
						new_variable, new_symbol_table := get_variable([]string{"num"}, symbol_table)
						symbol_table = new_symbol_table
						used_variables = append(used_variables, new_variable)
						symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"set", new_variable, "0"})
						symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"str.equals", lhs, rhs, new_variable})
						lhs = new_variable
						lhs_type = rhs_type
					}
					if operator == "!=" {
						new_variable, new_symbol_table := get_variable([]string{"num"}, symbol_table)
						symbol_table = new_symbol_table
						used_variables = append(used_variables, new_variable)
						symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"set", new_variable, "0"})
						symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{operation, lhs, rhs, new_variable})
						symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"equals", new_variable, "false", new_variable})
						lhs = new_variable
						lhs_type = rhs_type
					}
				}
				if lhs_type[0] == "num" {
					switch operator {
					case "==":
						operation = "equals"
					case "!=":
						operation = "equals"
					case ">":
						operation = "greater"
					case "<":
						operation = "smaller"
					}
					if operator == "!=" || operator == "==" {
						new_variable, new_symbol_table := get_variable([]string{"num"}, symbol_table)
						symbol_table = new_symbol_table
						used_variables = append(used_variables, new_variable)
						symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"set", new_variable, "0"})
						symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"equals", lhs, rhs, new_variable})
						if operator == "!=" {
							symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"equals", new_variable, "false", new_variable})
						}
						lhs = new_variable
						lhs_type = rhs_type
					}
					if operator == ">" || operator == "<" {
						new_variable, new_symbol_table := get_variable([]string{"num"}, symbol_table)
						symbol_table = new_symbol_table
						used_variables = append(used_variables, new_variable)
						symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"set", new_variable, "0"})
						if operator == "<" {
							symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"smaller", lhs, rhs, new_variable})
						} else {
							symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"greater", lhs, rhs, new_variable})
						}
						lhs = new_variable
						lhs_type = rhs_type
					}
				}
			}
			resultant_variable=lhs
		}
	}
	return resultant_variable, used_variables, symbol_table
}

func get_branch_counts(symbol_table Symbol_Table) int {
	i:=0
	for {
		i+=1
		branch_str:="branch_"+strconv.FormatInt(int64(i), 10)
		if str_index_in_arr(branch_str, symbol_table.data)==-1 {
			return i
		}
	}
}

func compiler(symbol_table Symbol_Table, function_name string, depth int, code_original []Token, in_loop bool) (string, Symbol_Table) {
	code:=make([]Token, 0)
	for _,code_obj:=range code_original {
		code = append(code, clone_token(code_obj))
	}
	new_data:=make([]string, 0)
	if depth != 0 {
		fmt.Println("Compiling", function_name)
	}
	if depth == 0 {
		for i := 0; i < len(symbol_table.functions); i++ {
			// compiling functions
			if strings.Split(symbol_table.functions[i].name, "-")[0] != symbol_table.current_file {
				continue
			}
			symbol_table.functions[i].Code = nested_puller(symbol_table.functions[i].Code)
			if symbol_table.operations[symbol_table.functions[i].name] == nil {
				symbol_table.operations[symbol_table.functions[i].name] = make([][]string, 0)
			}
			variable_names := make([]string, 0)
			for _,arg := range symbol_table.functions[i].args_keys {
				arg_type:=symbol_table.functions[i].args[arg]
				// adding function inputs to stack
				if variable_index_in_symbol_table(arg, symbol_table) != -1 {
					return "input_of_function_cannot_have_same_name_as_global_variable", symbol_table
				}
				variable_type := arg_type
				variable_name := arg
				variable_names = append(variable_names, variable_name)
				symbol_table.variable_mapping[variable_name] = symbol_table.functions[i].name + "-" + "function_input:" + variable_name
				if variable_type[0] == "string" {
					string_index := str_index_in_arr("", symbol_table.data)
					if string_index == -1 {
						symbol_table.data = append(symbol_table.data, "")
						string_index = len(symbol_table.data) - 1
					}
					symbol_table.variables = append(symbol_table.variables, Variable{name: variable_name, Type: variable_type})
					symbol_table.global_variables = append(symbol_table.global_variables, []string{"str.set", symbol_table.variable_mapping[variable_name], strconv.FormatInt(int64(string_index), 10)})
				} else if variable_type[0] == "num" {
					symbol_table.variables = append(symbol_table.variables, Variable{name: variable_name, Type: variable_type})
					symbol_table.global_variables = append(symbol_table.global_variables, []string{"set", symbol_table.variable_mapping[variable_name], "0"})
				} else if variable_type[0] == "[" {
					new_variable_type := make([]string, 0)
					for _, vb := range variable_type {
						new_variable_type = append(new_variable_type, vb)
					}
					data_index := str_index_in_arr(strings.Join(string_array_types_to_vitality_types(variable_type, symbol_table)[1:], ","), symbol_table.data)
					if data_index == -1 {
						symbol_table.data = append(symbol_table.data, strings.Join(string_array_types_to_vitality_types(variable_type, symbol_table)[1:], ","))
						data_index = len(symbol_table.data) - 1
					}
					symbol_table.variables = append(symbol_table.variables, Variable{name: variable_name, Type: new_variable_type})
					symbol_table.global_variables = append(symbol_table.global_variables, []string{"arr.init", symbol_table.variable_mapping[variable_name], strconv.FormatInt(int64(data_index), 10)})
				} else if variable_type[0] == "{" {
					data_index := str_index_in_arr(strings.Join(string_array_types_to_vitality_types(variable_type, symbol_table)[1:], ","), symbol_table.data)
					if data_index == -1 {
						symbol_table.data = append(symbol_table.data, strings.Join(string_array_types_to_vitality_types(variable_type, symbol_table)[1:], ","))
						data_index = len(symbol_table.data) - 1
					}
					symbol_table.variables = append(symbol_table.variables, Variable{name: variable_name, Type: variable_type})
				} else if len(variable_type) == 1 {
					if !strings.Contains(variable_type[0], "-") {
						variable_type = []string{symbol_table.current_file + "-" + variable_type[0]}
					}
					if struct_index_in_symbol_table(variable_type[0], symbol_table) == -1 {
						return "Invalid variable initialisation", symbol_table
					}
					symbol_table.variables = append(symbol_table.variables, Variable{name: variable_name, Type: variable_type})
					symbol_table.data = append(symbol_table.data, variable_type[0])
					symbol_table.global_variables = append(symbol_table.global_variables, []string{"struct.init", symbol_table.current_file + "-" + symbol_table.variable_mapping[variable_name], strconv.FormatInt(int64(str_index_in_arr(variable_type[0], symbol_table.data)), 10)})
				}
			}
			err, returned_symbol_table := compiler(symbol_table, symbol_table.functions[i].name, depth+1, make([]Token, 0), in_loop)
			symbol_table = returned_symbol_table
			new_variables := make([]Variable, 0)
			for _, arg := range symbol_table.variables {
				if str_index_in_arr(arg.name, variable_names) == -1 {
					new_variables = append(new_variables, arg)
				}
			}
			symbol_table.variables = new_variables
			if err != "" {
				return err, symbol_table
			}
		}
		return "", symbol_table
	}
	if depth > 0 {
		// depth and recursive file compiling detection
		if len(function_name) != 0 && string(function_name[0]) != "-" {
			code = symbol_table.functions[function_index_in_symbol_table(function_name, symbol_table)].Code
		}
		if string(function_name[0]) == "-" {
			function_name = function_name[1:]
		}
		copy_of_code:=make([]Token, 0)
		for _,code_obj:=range code {
			copy_of_code = append(copy_of_code, clone_token(code_obj))
		}
		code=copy_of_code
		for i := 0; i < len(code); i++ {
			if len(new_data)!=0 {
				symbol_table.data=new_data
				new_data=make([]string, 0)
			}
			code:=make([]Token, 0)
			for _,code_obj:=range copy_of_code {
				code = append(code, code_obj)
			}
			if code[i].Type == "sys" && code[i].string_value == "var" && (len(code)-i) >= 4 && code[i+1].Type == "variable" && valid_var_name(code[i+1].string_value) && valid_type(type_token_to_string_array(code[i+2]), symbol_table) && code[i+3].Type == "EOS" && variable_doesnot_exist(symbol_table, code[i+1].string_value) && function_index_in_symbol_table(symbol_table.current_file+"-"+code[i+1].string_value, symbol_table)==-1 {
				symbol_table.variables = append(symbol_table.variables, Variable{name: code[i+1].string_value, Type: type_token_to_string_array(code[i+2])})
				variable_type := type_token_to_string_array(code[i+2])
				if variable_type[0] == "string" {
					string_index := str_index_in_arr("", symbol_table.data)
					if string_index == -1 {
						symbol_table.data = append(symbol_table.data, "")
						string_index = len(symbol_table.data) - 1
					}
					symbol_table.global_variables = append(symbol_table.global_variables, []string{"str.set", symbol_table.current_file + "-" + code[i+1].string_value, strconv.FormatInt(int64(string_index), 10)})
				} else if variable_type[0] == "num" {
					symbol_table.global_variables = append(symbol_table.global_variables, []string{"set", symbol_table.current_file + "-" + code[i+1].string_value, "0"})
				} else if variable_type[0] == "[" {
					new_variable_type := make([]string, 0)
					for _, vb := range variable_type {
						new_variable_type = append(new_variable_type, vb)
					}
					data_index := str_index_in_arr(strings.Join(string_array_types_to_vitality_types(variable_type, symbol_table)[1:], ","), symbol_table.data)
					if data_index == -1 {
						symbol_table.data = append(symbol_table.data, strings.Join(string_array_types_to_vitality_types(variable_type, symbol_table)[1:], ","))
						data_index = len(symbol_table.data) - 1
					}
					symbol_table.global_variables = append(symbol_table.global_variables, []string{"arr.init", symbol_table.current_file + "-" + code[i+1].string_value, strconv.FormatInt(int64(data_index), 10)})
				} else if variable_type[0] == "{" {
					data_index := str_index_in_arr(strings.Join(string_array_types_to_vitality_types(variable_type, symbol_table)[1:], ","), symbol_table.data)
					if data_index == -1 {
						symbol_table.data = append(symbol_table.data, strings.Join(string_array_types_to_vitality_types(variable_type, symbol_table)[1:], ","))
						data_index = len(symbol_table.data) - 1
					}
				} else if len(variable_type) == 1 {
					if !strings.Contains(variable_type[0], "-") {
						variable_type = []string{symbol_table.current_file + "-" + variable_type[0]}
					}
					if struct_index_in_symbol_table(variable_type[0], symbol_table) == -1 {
						return "Invalid variable initialisation", symbol_table
					}
					if str_index_in_arr(variable_type[0], symbol_table.data)==-1 {
						symbol_table.data = append(symbol_table.data, variable_type[0])
					}
					symbol_table.global_variables = append(symbol_table.global_variables, []string{"struct.init", symbol_table.current_file + "-" + code[i+1].string_value, strconv.FormatInt(int64(str_index_in_arr(variable_type[0], symbol_table.data)), 10)})
				}
				// symbol_table.operations[function_name]=append(symbol_table.operations[function_name], )
				// you need to add the compiler
				i += 3
				new_data=symbol_table.data
				continue
			}
			if (len(code)-i) >= 4 && (code[i].Type == "lookup" || code[i].Type == "variable" || code[i].Type == "nested_tokens") && code[i+1].Type == "operator" && strings.Contains(code[i+1].string_value, "=") {
				i++
				i++
				tokens, err := get_current_statement_tokens(code[i:])
				if err != 0 {
					return "unexpected end of statement", symbol_table
				}
				new_tokens := make([]Token, 0)
				for _, x := range tokens {
					new_struct, _, _ := CloneMyStruct(x)
					new_tokens = append(new_tokens, new_struct)
				}
				lhs := evaluate_type(symbol_table, []Token{code[i-2]}, 0)
				rhs := evaluate_type(symbol_table, new_tokens, 0)
				resultant_variable := ""
				used_variable := make([]string, 0)
				resultant_variable, used_variable, symbol_table = expression_solver(tokens, function_name, symbol_table, false)
				for _, variable := range used_variable {
					free_variable(variable, symbol_table)
				}
				if string_arr_compare(lhs, []string{}) {
					return "invalid type on lhs:" + symbol_table.current_file, symbol_table
				}
				if string_arr_compare(rhs, []string{}) {
					fmt.Println(new_tokens)
					return "invalid type on rhs:" + symbol_table.current_file, symbol_table
				}
				if !string_arr_compare(lhs, rhs) {
					return "types on lhs and rhs do not match", symbol_table
				}
				if lhs[0]=="num" || lhs[0]=="string" {
					lhs_resultant_variable:=""
					lhs_resultant_variable, used_variable, symbol_table = expression_solver([]Token{code[i-2]}, function_name, symbol_table, false)
					for _, variable := range used_variable {
						free_variable(variable, symbol_table)
					}
					if lhs[0]=="num" {
						symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"refset", lhs_resultant_variable, resultant_variable})
					}
					if lhs[0]=="string" {
						symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"str.refset", lhs_resultant_variable, resultant_variable})
					}
				}
				i += len(tokens)
				new_data=symbol_table.data
				continue
			}
			if code[i].Type == "branch" {
				if !string_arr_compare(evaluate_type(symbol_table, code[i].children[0].children, 0), []string{"num"}) {
					return "branch condition is invalid", symbol_table
				}
				used_variable := make([]string, 0)
				resultant_variable, used_variable, symbol_table := expression_solver(code[i].children[0].children, function_name, symbol_table, false)
				for _, variable := range used_variable {
					free_variable(variable, symbol_table)
				}
				flip_variable,symbol_table:=get_variable([]string{"num"}, symbol_table)
				symbol_table=free_variable(flip_variable, symbol_table)
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"refset", flip_variable, resultant_variable})
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"equals", flip_variable, "false", flip_variable})
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"jump.def", "", flip_variable})
				operations_before_compilation:=len(symbol_table.operations[function_name])
				branch_string:="branch_"+strconv.FormatInt(int64(get_branch_counts(symbol_table)), 10)
				branch_index_in_symbol_table:=str_index_in_arr(branch_string, symbol_table.data)
				if branch_index_in_symbol_table==-1 {
					branch_index_in_symbol_table=len(symbol_table.data)
					symbol_table.data = append(symbol_table.data, branch_string)
				}
				err, st := compiler(symbol_table, "-"+function_name, depth+1, code[i].children[1].children, in_loop)
				st.operations[function_name][operations_before_compilation-1][1]=strconv.FormatInt(int64(branch_index_in_symbol_table), 10)
				st.operations[function_name] = append(st.operations[function_name], []string{"define.jump", strconv.FormatInt(int64(branch_index_in_symbol_table), 10)+" // to continue code after if in function "+symbol_table.current_file+"-"+function_name})
				if err != "" {
					return err, st
				}
				symbol_table = st
				new_data=symbol_table.data
				continue
			}
			if code[i].Type == "while" {
				if !string_arr_compare(evaluate_type(symbol_table, code[i].children[0].children, 0), []string{"num"}) {
					return "while condition for while is invalid", symbol_table
				}
				loop_condition_variable:="branch_"+strconv.FormatInt(int64(get_branch_counts(symbol_table)), 10)
				loop_condition_index:=str_index_in_arr(loop_condition_variable, symbol_table.data)
				if loop_condition_index==-1 {
					symbol_table.data = append(symbol_table.data, loop_condition_variable)
					loop_condition_index = len(symbol_table.data)-1
				}
				after_loop_variable:="branch_"+strconv.FormatInt(int64(get_branch_counts(symbol_table)), 10)
				after_loop_index:=str_index_in_arr(after_loop_variable, symbol_table.data)
				if after_loop_index==-1 {
					symbol_table.data = append(symbol_table.data, after_loop_variable)
					after_loop_index = len(symbol_table.data)-1
				}
				current_scope:=symbol_table.current_scope
				symbol_table.current_scope.current_loop_continue_line =  loop_condition_index
				symbol_table.current_scope.current_loop_break_line = after_loop_index
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"define.jump", strconv.FormatInt(int64(loop_condition_index), 10)+" // to reevaluate while condition in function "+symbol_table.current_file+"-"+function_name})
				used_variable := make([]string, 0)
				resultant_variable, used_variable, symbol_table := expression_solver(code[i].children[0].children, function_name, symbol_table, false)
				for _, variable := range used_variable {
					free_variable(variable, symbol_table)
				}
				flip_variable,symbol_table:=get_variable([]string{"num"}, symbol_table)
				symbol_table=free_variable(flip_variable, symbol_table)
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"refset", flip_variable, resultant_variable})
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"equals", flip_variable, "false", flip_variable})
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"jump.def", strconv.FormatInt(int64(after_loop_index), 10), flip_variable})
				err, st := compiler(symbol_table, "-"+function_name, depth+1, code[i].children[1].children, true)
				if err != "" {
					return err, st
				}
				symbol_table = st
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"jump.def", strconv.FormatInt(int64(loop_condition_index), 10), "true"})
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"define.jump", strconv.FormatInt(int64(after_loop_index), 10)+" // to continue after while in function "+symbol_table.current_file+"-"+function_name})
				symbol_table.current_scope.current_loop_continue_line=current_scope.current_loop_continue_line
				symbol_table.current_scope.current_loop_break_line=current_scope.current_loop_break_line
				new_data=symbol_table.data
				continue
			}
			if code[i].string_value == "return" && code[i].Type == "sys" {
				tokens, err := get_current_statement_tokens(code[i+1:])
				if err != 0 {
					return "Unexpected EOS", symbol_table
				}
				if len(tokens) == 0 && len(symbol_table.functions[function_index_in_symbol_table(function_name, symbol_table)].Type) == 0 {
					i += 1
					continue
				}
				returned_type := evaluate_type(symbol_table, tokens, 0)
				if len(returned_type) == 0 {
					return "Invalid return type", symbol_table
				}
				if !string_arr_compare(returned_type, symbol_table.functions[function_index_in_symbol_table(function_name, symbol_table)].Type) {
					return "Return type does not match", symbol_table
				}
				used_variable := make([]string, 0)
				resultant_variable, used_variable, symbol_table := expression_solver(tokens, function_name, symbol_table, false)
				for _, variable := range used_variable {
					free_variable(variable, symbol_table)
				}
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"obj_copy", function_name+"-"+"return-variable", resultant_variable})
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"jump.always.var","return_to"})
				i += len(tokens) + 1
				new_data=symbol_table.data
				continue
			}
			if code[i].string_value == "continue" && code[i].Type == "sys" {
				if !in_loop {
					return "cannot loop", symbol_table
				}
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"jump.def", strconv.FormatInt(int64(symbol_table.current_scope.current_loop_continue_line), 10), "true"})
				i += 1
				new_data=symbol_table.data
				continue
			}
			if code[i].string_value == "break" && code[i].Type == "sys" {
				if !in_loop {
					return "cannot break loop", symbol_table
				}
				symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"jump.def", strconv.FormatInt(int64(symbol_table.current_scope.current_loop_break_line), 10), "true"})
				i += 1
				new_data=symbol_table.data
				continue
			}
			if code[i].Type == "array" {
				fmt.Println(code[i])
			}
			if code[i].Type == "funcall" {
				if code[i].children[0].string_value=="print" && string_arr_compare(evaluate_type(symbol_table, []Token{code[i].children[1]}, 0), []string{"string"}) {
					used_variable := make([]string, 0)
					resultant_variable, used_variable, symbol_table := expression_solver(code[i].children[1].children, function_name, symbol_table, false)
					for _, variable := range used_variable {
						free_variable(variable, symbol_table)
					}
					symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"debug.print", resultant_variable})
					if len(code)>i+1 && code[i+1].Type=="EOS" {
						i+=1
					}
					new_data=symbol_table.data
					continue
				}
				if code[i].children[0].string_value=="print" && string_arr_compare(evaluate_type(symbol_table, []Token{code[i].children[1]}, 0), []string{"num"}) {
					used_variable := make([]string, 0)
					resultant_variable, used_variable, symbol_table := expression_solver(code[i].children[1].children, function_name, symbol_table, false)
					for _, variable := range used_variable {
						free_variable(variable, symbol_table)
					}
					var_num_to_str, symbol_table:=get_variable([]string{"string"}, symbol_table)
					symbol_table=var_init([]string{"string"}, var_num_to_str, symbol_table, function_name, false, false)
					symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"num_to_str", resultant_variable, var_num_to_str})
					symbol_table.operations[function_name] = append(symbol_table.operations[function_name], []string{"debug.print", var_num_to_str})
					if len(code)>i+1 && code[i+1].Type=="EOS" {
						i+=1
					}
					new_data=symbol_table.data
					continue
				}
			}
			fmt.Println("missed token", code[i], symbol_table.current_file, function_name)
			return "error_token_not_parsed", symbol_table
		}
		if len(new_data)!=0 {
			symbol_table.data=new_data
		}
	}
	return "", symbol_table
}

func data_encoder(data []string) string {
	result := ""
	for _, data_item := range data {
		result += "    \"" + strings.ReplaceAll(data_item, "\n", "\\n") + "\"" + "    \n"
	}
	return strings.Trim(result, "\n")
}

func tokens_to_string_constants(tokens []Token) []string {
	out := make([]string, 0)
	for _, token := range tokens {
		if len(token.children) != 0 {
			for _, string_constants := range tokens_to_string_constants(token.children) {
				if str_index_in_arr(string_constants, out) == -1 {
					out = append(out, string_constants)
				}
			}
		}
		if token.Type == "string" {
			if str_index_in_arr(token.string_value, out) == -1 {
				out = append(out, token.string_value)
			}
		}
	}
	return out
}

func formatter(output string) string {
	formatted_code := ""
	keep_adding := false
	for i, line := range strings.Split(output, "\n") {
		if keep_adding {
			formatted_code += line + "\n"
			continue
		}
		if strings.HasPrefix(line, "define.jump") {
			formatted_code += "\n    " + line + "\n"
			continue
		} else if i == 0 && strings.HasPrefix(line, "jump.def.always") {
			formatted_code += "\n    " + line + "\n"
			continue
		}
		if strings.HasPrefix(line, ".data") {
			formatted_code += "\n.data"
			keep_adding = true
			continue
		}
		formatted_code += "        " + line + "\n"
	}
	return formatted_code
}

func is_base_type(Type []string) bool {
	if len(Type) == 0 {
		return true
	}
	valid_final_types := []string{"num", "string"}
	if len(Type)%2 == 0 {
		return false
	}
	for i := 0; i < ((len(Type)-1)/2)-1; i++ {
		if Type[i] == "{" && Type[len(Type)-1] == "}" {
			continue
		}
		if Type[i] == "[" && Type[len(Type)-1] == "]" {
			continue
		}
		return false
	}
	if str_index_in_arr(Type[(len(Type)-1)/2], valid_final_types) != -1 {
		return true
	}
	return false
}

func build(symbol_table Symbol_Table, tokens []Token, depth int) (string, Symbol_Table) {
	file_name := symbol_table.current_file
	symbol_table.data = append(symbol_table.data, tokens_to_string_constants(tokens)...)
	err, symbol_table := pre_parser(symbol_table, tokens, 0)
	if err != "" {
		return "Error: " + err, symbol_table
	}
	result := ""
	if depth == 0 {
		if !does_function_exist(symbol_table.current_file+"-"+"main", symbol_table) {
			return "Error: " + "main function does not exist", symbol_table
		}
		main_function := symbol_table.functions[function_index_in_symbol_table(symbol_table.current_file+"-"+"main", symbol_table)]
		if len(main_function.args) != 0 {
			return "Error: " + "main function cant have arguments", symbol_table
		}
		if len(main_function.Type) != 0 {
			return "Error: " + "main function cannot have a type definition", symbol_table
		}
	}
	err, final_symbol_table := compiler(symbol_table, "this_does_not_matter", 0, make([]Token, 0), false)
	symbol_table = final_symbol_table
	for _,fn := range symbol_table.functions {
		if depth==0 && strings.Split(fn.name, "-")[0] == file_name {
			if strings.Split(fn.name, "-")[1] != "main" {
				symbol_table.operations[fn.name] = append(symbol_table.operations[fn.name], []string{"jump.always.var","return_to // returning by default"})
			} else {
				symbol_table.operations[fn.name] = append(symbol_table.operations[fn.name], []string{"set","return_to","-1"})
				symbol_table.operations[fn.name] = append(symbol_table.operations[fn.name], []string{"jump.always.var","return_to // returning after main"})
			}
		}
		function:=fn.name
		result += "define.jump " + strconv.FormatInt(int64(str_index_in_arr(function, symbol_table.data)), 10) + " // " + function + "\n"
		for _, operations := range symbol_table.operations[function] {
			if len(operations) != 0 {
				result += operations[0] + " " + strings.Join(operations[1:], ",") + "\n"
			}
		}
	}
	if err != "" {
		return "Error: " + err, symbol_table
	}
	if depth == 0 {
		data_index := str_index_in_arr("stack-init", symbol_table.data)
		if data_index == -1 {
			symbol_table.data = append(symbol_table.data, "stack-init")
			data_index = len(symbol_table.data) - 1
		}
		prefix_result := ""
		build_info := make([]string, 0)
		for _, function := range symbol_table.functions {
			if !string_arr_compare(function.Type, []string{}) {
				actual_current_file:=strings.Clone(symbol_table.current_file)
				symbol_table.current_file=strings.Split(function.name, "-")[0]
				symbol_table=var_init(function.Type, function.name+"-"+"return-variable", symbol_table, "does_not_matter", true, true)
				symbol_table.current_file=strings.Clone(actual_current_file)
			}
			if strings.Split(function.name, "-")[0] != file_name {
				continue
			}
			continue_external := false
			all_arguments := make([]string, 0)
			for _, argument := range function.args_keys {
				if !is_base_type(function.args[argument]) {
					continue_external = true
					break
				}
				all_arguments = append(all_arguments, strings.Join(string_array_types_to_vitality_types(function.args[argument], symbol_table), ","))
			}
			if continue_external {
				continue
			}
			build_info = append(build_info, strings.Split(function.name, "-")[1]+"::"+strings.Join(string_array_types_to_vitality_types(function.Type, symbol_table), ",")+"::"+strings.Join(all_arguments, ".."))
		}
		prefix_result += "jump.def.always " + strconv.FormatInt(int64(data_index), 10) + "\ndefine.jump " + strconv.FormatInt(int64(data_index), 10) + " // stack init\n"
		prefix_result += "set true,1\nset false,0\nset return_to,0\n"
		for _, var_initialisation := range symbol_table.struct_registration {
			prefix_result += var_initialisation[0] + " " + strings.Join(var_initialisation[1:], ",") + "\n"
		}
		for _, var_initialisation := range symbol_table.global_variables {
			prefix_result += var_initialisation[0] + " " + strings.Join(var_initialisation[1:], ",") + "\n"
		}
		prefix_result += "jump.def.always " + strconv.FormatInt(int64(str_index_in_arr(symbol_table.current_file+"-"+"main", symbol_table.data)), 10) + "\n"
		result = formatter(prefix_result + result)
		symbol_table.data = append(symbol_table.data, strings.Join(build_info, ";"))
	}
	symbol_table.finished_importing = true
	return ".code\n" + result + ".data\n" + data_encoder(symbol_table.data), symbol_table
}
