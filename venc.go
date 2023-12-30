package main

import (
	"errors"
	"fmt"
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

type Function struct {
	name		string
	args		[][]string
	Type		[]string
	Code 		[]Token
}

type Struct struct {
	name 		string
	fields		map[string][]string
}

type Variable struct {
	name 		string
	Type 		[]string
}

type Symbol_Table struct {
	functions 			[]Function
	structs   			[]Struct
	variables			[]Variable
	data				[]string
	operations			map[string][][]string
	used_variables		map[string]int
	variable_mapping	map[string]Variable
}

var reserved_tokens = []string{"var", "fn", "if", "while", "continue", "break", "struct","return", "function"}
var type_tokens = []string{"string", "num"}
var operators = []string{"+","-","*","/","^",">","<","=","&","!","|","%"}
var end_of_statements = []string{";"}
var brackets = []string{"(",")","[","]","{","}"}
var string_quotes = []string{"\"","'"}
var comma = ","
var allowed_variable_character="abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
var comments = []string{"#"}

func remove_token_at_index(i int, tokens []Token) []Token {
	return append(tokens[:i], tokens[i+1:]...)
}

func tokensier(code string, debug bool) []Token {
	tokens := make([]Token, 0)
	cache := ""
	for i := 0; i < len(code); i++ {
		char := string(code[i])
		if strings.Contains("1234567890.",char) {
			for {
				char := string(code[i])
				if strings.Contains(".1234567890",char) {
					cache+=char
				} else {
					if cache=="." {
						tokens = append(tokens, Token{Type: "dot"})
						cache=""
						i--
						break
					}
					number,_:=strconv.ParseFloat(cache,64)
					tokens = append(tokens, Token{Type: "num", num_value: number})
					i--
					cache=""
					break
				}
				i++
				if i==len(code) {
					if cache=="." {
						tokens = append(tokens, Token{Type: "dot"})
						cache=""
						i--
						break
					}
					number,_:=strconv.ParseFloat(cache,64)
					tokens = append(tokens, Token{Type: "num", num_value: number})
					i--
					cache=""
					break
				}
			}
			continue
		}
		if str_index_in_arr(char,brackets)!=-1 {
			open_type:=""
			if (str_index_in_arr(char,brackets)%2)!=1 {
				open_type="open"
			} else {
				open_type="close"
			}
			tokens = append(tokens, Token{Type: "bracket_"+open_type,string_value: char})
			continue
		}
		if str_index_in_arr(char,operators)!=-1 {
			tokens = append(tokens, Token{Type: "operator",string_value: char})
			continue
		}
		if str_index_in_arr(char,end_of_statements)!=-1 {
			tokens = append(tokens, Token{Type: "EOS"})
			continue
		}
		if str_index_in_arr(char,string_quotes)!=-1 {
			string_init:=char
			for {
				i++
				if i==len(code) {
					if debug {
						fmt.Println("Unexpected EOF")
					}
					return make([]Token, 0)
				}
				char := string(code[i])
				if char!=string_init {
					cache+=char
				} else {
					tokens = append(tokens, Token{Type: "string", string_value: str_parser(cache)})
					cache=""
					break
				}
			}
			continue
		}
		if char==comma {
			tokens = append(tokens, Token{Type: "comma"})
			continue
		}
		if char==":" {
			tokens = append(tokens, Token{Type: "colon"})
			continue
		}
		if char=="." {
			tokens = append(tokens, Token{Type: "dot"})
			continue
		}
		if strings.Contains(allowed_variable_character,char) {
			for {
				char := string(code[i])
				if strings.Contains(allowed_variable_character,char) {
					cache+=char
				} else {
					if str_index_in_arr(cache, reserved_tokens)!=-1 {
						tokens = append(tokens, Token{Type: "sys", string_value: cache})
					} else {
						tokens = append(tokens, Token{Type: "variable", string_value: cache})
					}
					cache=""
					i--
					break
				}
				i++
				if i==len(code) {
					tokens = append(tokens, Token{Type: "variable", string_value: cache})
					cache=""
					i--
					break
				}
			}
			continue
		}
		if str_index_in_arr(char,comments)!=-1 {
			for {
				i++
				if i==len(code) {
					break
				}
				char := string(code[i])
				if str_index_in_arr(char,end_of_statements)==-1 && char!="\n" {
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
	if var_name=="" {
		return false
	}
	if str_index_in_arr(var_name, reserved_tokens)!=-1 {
		return false
	}
	for _,char:=range "1234567890" {
		var_name=strings.ReplaceAll(var_name,string(char),"")
	}
	if var_name=="" {
		return false
	}
	for _,char:=range allowed_variable_character {
		var_name=strings.ReplaceAll(var_name,string(char),"")
	}
	return var_name==""
}

func tokens_parser(code []Token, debug bool) ([]Token, error) {
	parsed_tokens:=make([]Token,0)
	for i := 0; i < len(code); i++ {
		current_token:=code[i]
		if len(code)>i+1 && current_token.Type=="variable" && valid_var_name(current_token.string_value) && code[i+1].Type=="dot" {
			nested_variables:=make([]Token,0)
			first:=true
			for {
				if len(code)>i+1 && current_token.Type=="variable" && valid_var_name(code[i].string_value) && (!first || code[i+1].Type=="dot") { 
					if !first {
						if code[i-1].Type!="dot" {
							break
						}
					}
					nested_variables = append(nested_variables, code[i])
					if code[i+1].Type=="dot" {
						i++
					}
					i++
					first=false
				} else {
					if !first && !(len(code)>i+1) && current_token.Type=="variable" && valid_var_name(code[i].string_value) && code[i-1].string_value=="." {
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
		if len(code)>i+1 && current_token.Type=="operator" && code[i+1].Type=="operator" {
			combined_operator:=current_token.string_value+code[i+1].string_value
			accepted_combined_operator_array:=[]string{"+=","-=","/=","*=","%=","//","!=","==","->",":=","||","&&"}
			if str_index_in_arr(combined_operator,accepted_combined_operator_array)!=-1 {
				parsed_tokens = append(parsed_tokens, Token{Type: "operator", string_value: combined_operator})
				i++
				continue
			}
		}
		if len(code)>i+1 && current_token.Type=="colon" && code[i+1].Type=="operator" {
			combined_operator:=":"+code[i+1].string_value
			if combined_operator==":=" {
				parsed_tokens = append(parsed_tokens, Token{Type: "operator", string_value: combined_operator})
				i++
				continue
			}
		}
		if len(parsed_tokens)>0 && parsed_tokens[len(parsed_tokens)-1].Type=="operator" && parsed_tokens[len(parsed_tokens)-1].string_value=="->" && ((current_token.string_value=="string" || current_token.string_value=="num" || current_token.Type=="variable") || (len(code)>i+1 && len(parsed_tokens)!=0 && current_token.Type=="bracket_open" && (current_token.string_value=="[" || current_token.string_value=="{"))) {
			type_define_tokens:=make([]Token,0)
			brackets:=0
			for {
				if len(code)<i+1 {
					return make([]Token, 0), errors.New("Unexpected EOF")
				}
				if code[i].Type=="bracket_open" && code[i].string_value==current_token.string_value {
					brackets+=1
				}
				if code[i].Type=="bracket_close" && ((code[i].string_value=="]" && current_token.string_value=="[") || (code[i].string_value=="}" && current_token.string_value=="{")) {
					brackets-=1
				}
				if str_index_in_arr(code[i].Type, []string{"bracket_open","bracket_close","variable"})==-1 {
					return make([]Token, 0), errors.New("Illegal type definition")
				}
				type_define_tokens = append(type_define_tokens, code[i])
				if brackets==0 {
					break
				}
				i++
			}
			parsed_tokens[len(parsed_tokens)-1] = Token{Type: "type", keys: type_define_tokens}
			continue
		}
		if len(code)>i+1 && len(parsed_tokens)!=0 && current_token.Type=="variable" && str_index_in_arr(current_token.string_value, type_tokens)!=-1 && parsed_tokens[len(parsed_tokens)-1].Type=="operator" && parsed_tokens[len(parsed_tokens)-1].string_value=="->" {
			parsed_tokens[len(parsed_tokens)-1] = Token{Type: "type", children: []Token{current_token}}
			continue
		}
		if code[i].Type=="bracket_open" && code[i].string_value=="[" {
			bracket_count:=1
			childrentokens:=make([]Token,0)
			for {
				i++
				if len(code)<i+1 {
					return make([]Token, 0), errors.New("Unexpected EOF")
				}
				if code[i].Type=="bracket_open" && code[i].string_value=="[" {
					bracket_count+=1
				}
				if code[i].Type=="bracket_close" && code[i].string_value=="]" {
					bracket_count-=1
				}
				if bracket_count==0 {
					break
				}
				childrentokens = append(childrentokens, code[i])
			}
			tokens, err:=tokens_parser(childrentokens, debug)
			if err!=nil {
				return make([]Token, 0), err
			}
			parsed_tokens = append(parsed_tokens, Token{Type: "expression_wrapper_[]", children: tokens})
			continue
		}
		if code[i].Type=="bracket_open" && code[i].string_value=="(" {
			bracket_count:=1
			childrentokens:=make([]Token,0)
			for {
				i++
				if len(code)<i+1 {
					return make([]Token, 0), errors.New("Unexpected EOF")
				}
				if code[i].Type=="bracket_open" && code[i].string_value=="(" {
					bracket_count+=1
				}
				if code[i].Type=="bracket_close" && code[i].string_value==")" {
					bracket_count-=1
				}
				if bracket_count==0 {
					break
				}
				childrentokens = append(childrentokens, code[i])
			}
			tokens, err:=tokens_parser(childrentokens, debug)
			if err!=nil {
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
	grouped_tokens:=make([]Token,0)
	for i := 0; i < len(code); i++ {
		tokens_children,err:=token_grouper(code[i].children, debug)
		if err!=nil {
			return make([]Token, 0), err
		}
		code[i].children=tokens_children
		if code[i].Type=="expression_wrapper_[]" {
			if len(grouped_tokens)>0 && (grouped_tokens[len(grouped_tokens)-1].Type=="variable" || grouped_tokens[len(grouped_tokens)-1].Type=="nested_tokens" || grouped_tokens[len(grouped_tokens)-1].Type=="lookup" || grouped_tokens[len(grouped_tokens)-1].Type=="expression" || grouped_tokens[len(grouped_tokens)-1].Type=="funcall") {
				grouped_tokens[len(grouped_tokens)-1]=Token{Type: "lookup", children: []Token{Token{Type: "parent", children: []Token{grouped_tokens[len(grouped_tokens)-1]}}, Token{Type: "tokens", children: code[i].children}}}
				continue
			}
		}
		if len(grouped_tokens)>1 && (code[i].Type=="variable" || code[i].Type=="funcall") && grouped_tokens[len(grouped_tokens)-1].Type=="dot" && (grouped_tokens[len(grouped_tokens)-2].Type=="nested_tokens" || grouped_tokens[len(grouped_tokens)-2].Type=="lookup" || grouped_tokens[len(grouped_tokens)-2].Type=="variable" || grouped_tokens[len(grouped_tokens)-2].Type=="expression" || grouped_tokens[len(grouped_tokens)-2].Type=="funcall") {
			grouped_tokens[len(grouped_tokens)-2]=Token{Type: "nested_tokens", children: []Token{grouped_tokens[len(grouped_tokens)-2], code[i]}}
			grouped_tokens=remove_token_at_index(len(grouped_tokens)-1, grouped_tokens)
			continue
		}
		if len(grouped_tokens)>0 && code[i].Type=="expression" && (grouped_tokens[len(grouped_tokens)-1].Type=="variable" || grouped_tokens[len(grouped_tokens)-1].Type=="lookup" || grouped_tokens[len(grouped_tokens)-1].Type=="expression" || grouped_tokens[len(grouped_tokens)-1].Type=="funcall") {
			grouped_tokens[len(grouped_tokens)-1]=Token{Type: "funcall", children: []Token{grouped_tokens[len(grouped_tokens)-1], code[i]}}
			continue
		}
		grouped_tokens = append(grouped_tokens, code[i])
	}
	return grouped_tokens, nil
}

func bracket_token_getter(tokens []Token, bracket string) ([]Token, int64) {
	res:=make([]Token,0)
	brackets:=0
	close_bracket:=""
	if bracket=="{" {
		close_bracket="}"
	}
	if bracket=="(" {
		close_bracket=")"
	}
	if bracket=="[" {
		close_bracket="]"
	}
	for _,token:=range tokens {
		if token.Type=="bracket_open" && token.string_value==bracket {
			brackets+=1
		}
		if token.Type=="bracket_close" && token.string_value==close_bracket {
			brackets-=1
		}
		res = append(res, token)
		if brackets==0 {
			break
		}
	}
	return res, int64(brackets)
}

func get_current_statement_tokens(tokens []Token) ([]Token, int64) {
	res:=make([]Token,0)
	for _,token:=range tokens {
		if token.Type=="EOS" {
			return res, 0
		}
		res = append(res, token)
	}
	return res, 1
}

func variable_doesnot_exist(symbol_table Symbol_Table, variable string) bool {
	for i := 0; i < len(symbol_table.functions); i++ {
		if symbol_table.functions[i].name==variable {
			return false
		}
	}
	for i := 0; i < len(symbol_table.structs); i++ {
		if symbol_table.structs[i].name==variable {
			return false
		}
	}
	for i := 0; i < len(symbol_table.variables); i++ {
		if symbol_table.variables[i].name==variable {
			return false
		}
	}
	return true
}

func type_token_to_string_array(type_token Token) ([]string) {
	res:=make([]string, 0)
	for _,tokenx:=range type_token.keys {
		res = append(res, tokenx.string_value)
	}
	return res
}

func valid_type(Type []string, symbol_table Symbol_Table) bool {
	valid_final_types:=[]string{"num","string"}
	if len(Type)%2==0 {
		return false
	}
	for i:=0;i<((len(Type)-1)/2)-1;i++ {
		if Type[i]=="{" && Type[len(Type)-1]=="}" {
			continue
		}
		if Type[i]=="[" && Type[len(Type)-1]=="]" {
			continue
		}
		return false
	}
	for _,Struct:=range symbol_table.structs {
		valid_final_types = append(valid_final_types, Struct.name)
	}
	if str_index_in_arr(Type[(len(Type)-1)/2], valid_final_types)!=-1 {
		return true
	}
	return false
}

func does_function_exist(function_name string, symbol_table Symbol_Table) bool {
	for i := 0; i < len(symbol_table.functions); i++ {
		if symbol_table.functions[i].name==function_name {
			return true
		}
	}
	return false
}

func function_index_in_symbol_table(function_name string, symbol_table Symbol_Table) int {
	for i := 0; i < len(symbol_table.functions); i++ {
		if symbol_table.functions[i].name==function_name {
			return i
		}
	}
	return -1
}

func variable_index_in_symbol_table(variable_name string, symbol_table Symbol_Table) int {
	for i := 0; i < len(symbol_table.variables); i++ {
		if symbol_table.variables[i].name==variable_name {
			return i
		}
	}
	return -1
}

func struct_index_in_symbol_table(struct_name string, symbol_table Symbol_Table) int {
	for i := 0; i < len(symbol_table.structs); i++ {
		if symbol_table.structs[i].name==struct_name {
			return i
		}
	}
	return -1
}

func branch_parser(code []Token) (string, []Token) {
	parsed_tokens:=make([]Token, 0)
	for i := 0; i < len(code); i++ {
		if code[i].Type=="sys" && code[i].string_value=="if" && (len(code)-i)>=4 {
			if code[i+1].Type!="expression" || code[i+2].Type!="bracket_open" || code[i+2].string_value!="{" {
				return "invalid expression", make([]Token, 0)
			}
			tokens,err:=bracket_token_getter(code[i+2:], "{")
			if err!=0 {
				return "tokeniser_bracket_error", make([]Token, 0)
			}
			var errstring string
			errstring,tokens=branch_parser(tokens[1:len(tokens)-1])
			if errstring!="" {
				return errstring, make([]Token, 0)
			}
			parsed_tokens=append(parsed_tokens, Token{Type: "branch", children: []Token{code[i+1], Token{Type: "code", children: tokens}}})
			i+=3+len(tokens)
			continue
		}
		if code[i].Type=="sys" && code[i].string_value=="while" && (len(code)-i)>=4 {
			if code[i+1].Type!="expression" || code[i+2].Type!="bracket_open" || code[i+2].string_value!="{" {
				return "invalid expression", make([]Token, 0)
			}
			tokens,err:=bracket_token_getter(code[i+2:], "{")
			if err!=0 {
				return "tokeniser_bracket_error", make([]Token, 0)
			}
			var errstring string
			errstring,tokens=branch_parser(tokens[1:len(tokens)-1])
			if errstring!="" {
				return errstring, make([]Token, 0)
			}
			parsed_tokens=append(parsed_tokens, Token{Type: "while", children: []Token{code[i+1], Token{Type: "code", children: tokens}}})
			i+=3+len(tokens)
			continue
		}
		parsed_tokens = append(parsed_tokens, code[i])
	}
	return "", parsed_tokens
}

func pre_parser(symbol_table Symbol_Table, code []Token, depth int) (string, Symbol_Table) {
	result:=""
	for i := 0; i < len(code); i++ {
		// fmt.Println(code[i])
		if depth==0 {
			if code[i].Type=="sys" && code[i].string_value=="struct" && len(code)>i+2 && code[i+1].Type=="variable" && valid_var_name(code[i+1].string_value) && code[i+2].Type=="bracket_open" && code[i+2].string_value=="{" {
				if !variable_doesnot_exist(symbol_table ,code[i+1].string_value) {
					return "error", Symbol_Table{}
				}
				i++
				i++
				tokens,err:=bracket_token_getter(code[i:], code[i].string_value)
				if err!=0 {
					return "error", Symbol_Table{}
				}
				tokens=tokens[1:len(tokens)-1]
				if len(tokens)%2!=0 {
					return "error", Symbol_Table{}
				}
				Struct_Variables:=make(map[string][]string)
				for index:=0; int64(index) < int64(len(tokens))-1; index+=2 {
					variable_type:=type_token_to_string_array(tokens[index+1])
					if tokens[index].Type=="variable" && valid_var_name(tokens[index].string_value) && variable_doesnot_exist(symbol_table, tokens[index].string_value) && valid_type(variable_type, symbol_table) {
						Struct_Variables[tokens[index].string_value]=variable_type
					} else {
						return "struct_error", Symbol_Table{}
					}
				}
				symbol_table.structs = append(symbol_table.structs, Struct{name: code[i-1].string_value, fields: Struct_Variables})
				i+=len(tokens)+1
				continue
			}
			if code[i].Type=="sys" && code[i].string_value=="function" && len(code)>i+1 && code[i+1].Type=="funcall" && variable_doesnot_exist(symbol_table ,code[i+1].children[0].string_value) {
				function_name:=code[i+1].children[0].string_value
				function_arguments:=make([][]string, 0)
				if len(code[i+1].children[1].children)%2!=0 {
					return "function_error", Symbol_Table{}
				}
				if !valid_var_name(code[i+1].children[0].string_value) {
					return "function_error_identifier", Symbol_Table{}
				}
				i++
				i++
				for index:=0; index<len(code[i-1].children[1].children); index+=2 {
					function_arguments=append(function_arguments, type_token_to_string_array(code[i-1].children[1].children[index+1]))
					if !valid_type(function_arguments[len(function_arguments)-1], symbol_table) {
						return "invalid function argument definition", Symbol_Table{}
					}
				}
				function_type:=make([]string, 0)
				if code[i].Type=="type" {
					function_type=type_token_to_string_array(code[i])
					if !valid_type(function_type, symbol_table) {
						return "function_return_type_is_invalid", Symbol_Table{}
					}
				}
				if code[i].Type=="type" {
					i++
				}
				tokens,err:=bracket_token_getter(code[i:], code[i].string_value)
				if err!=0 {
					return "error", Symbol_Table{}
				}
				to_ignore:=len(tokens[1:len(tokens)-1])
				err_,tokens:=branch_parser(tokens[1:len(tokens)-1])
				if err_!="" {
					return "error", Symbol_Table{}
				}
				symbol_table.functions = append(symbol_table.functions, Function{args: function_arguments, name: function_name, Type: function_type, Code: tokens})
				symbol_table.data = append(symbol_table.data, function_name)
				i+=to_ignore+1
				continue
			}
			if code[i].Type=="sys" && code[i].string_value=="var" && (len(code)-i)>=4 && code[i+1].Type=="variable" && valid_var_name(code[i+1].string_value) && valid_type(type_token_to_string_array(code[i+2]), symbol_table) && code[i+3].Type=="EOS" && variable_doesnot_exist(symbol_table, code[i+1].string_value) {
				symbol_table.variables = append(symbol_table.variables, Variable{name: code[i+1].string_value, Type: type_token_to_string_array(code[i+2])})
				symbol_table.variable_mapping[code[i+1].string_value]=Variable{name: code[i+1].string_value, Type: type_token_to_string_array(code[i+2])}
				i+=3
				continue
			}
		}
		return "error_token_not_parsed", symbol_table
	}
	return result, symbol_table
}

func is_valid_statement(symbol_table Symbol_Table, code []Token) bool {
	if len(code)%2!=1 {
		return false
	}
	for i,token:=range code {
		if token.Type=="expression" {
			if !is_valid_statement(symbol_table, token.children) {
				return false
			}
		}
		if i%2==0 {
			if str_index_in_arr(token.Type, []string{"lookup", "variable", "expression", "string", "num", "funcall", "nested_tokens"})==-1 {
				return false
			}
		}
		if i%2==1 {
			if token.Type!="operator" && str_index_in_arr(token.string_value, operators)==-1 {
				return false
			}
			if str_index_in_arr(token.string_value, []string{"|","&","!"})!=-1 {
				return false
			}
		}
	}
	return true
}

func are_function_arguments_valid(argument_expression Token, function Function, symbol_table Symbol_Table) bool {
	if len(argument_expression.children)%2!=1 {
		return false
	}
	arguments:=make([]Token,0)
	cache:=make([]Token,0)
	for i := 0; i < len(argument_expression.children); i++ {
		if argument_expression.children[i].Type=="comma" {
			if len(cache)==0 {
				return false
			}
			arguments = append(arguments, Token{Type: "expression", children: cache})
			cache=make([]Token, 0)
		} else {
			cache = append(cache, argument_expression.children[i])
		}
	}
	if len(cache)!=0 {
		arguments = append(arguments, Token{Type: "expression", children: cache})
	}
	if len(arguments)!=len(function.args) {
		return false
	}
	for i := 0; i < len(arguments); i++ {
		if !string_arr_compare(evaluate_type(symbol_table, arguments[i].children), function.args[i]) {
			return false
		}
	}
	return true
}

func evaluate_type(symbol_table Symbol_Table, code []Token) ([]string) {
	if len(code)==0 {
		return make([]string, 0)
	}
	current_type:=make([]string, 0)
	if is_valid_statement(symbol_table, code) {
		if len(code)==1 {
			if code[0].Type=="string" {
				return []string{"string"}
			}
			if code[0].Type=="num" {
				return []string{"num"}
			}
			if code[0].Type=="expression" {
				return evaluate_type(symbol_table, code[0].children)
			}
			if code[0].Type=="variable" {
				variable_index:=variable_index_in_symbol_table(code[0].string_value,symbol_table)
				if variable_index==-1 {
					return make([]string, 0)
				}
				return symbol_table.variables[variable_index].Type
			}
			if code[0].Type=="lookup" {
				parent_type:=evaluate_type(symbol_table, []Token{code[0].children[0].children[0]})
				if len(parent_type)==0 {
					return make([]string, 0)
				}
				if str_index_in_arr(parent_type[0], []string{"[","{"})==-1 {
					return make([]string, 0)
				}
				lookup_type:=evaluate_type(symbol_table, code[0].children[1].children)
				if len(lookup_type)==0 {
					return make([]string, 0)
				}
				if string_arr_compare(lookup_type, []string{"string"}) || string_arr_compare(lookup_type, []string{"num"}) {
					return parent_type[1:len(parent_type)-1]
				}
			}
			if code[0].Type=="nested_tokens" {
				variable_index:=variable_index_in_symbol_table(code[0].children[0].string_value, symbol_table)
				struct_index:=struct_index_in_symbol_table(symbol_table.variables[variable_index].Type[0], symbol_table)
				if struct_index==-1 {
					return make([]string, 0)
				}
				return symbol_table.structs[struct_index].fields[code[0].children[1].string_value]
			}
			if code[0].Type=="funcall" {
				function_index:=function_index_in_symbol_table(code[0].children[0].string_value, symbol_table)
				if function_index==-1 {
					return make([]string, 0)
				}
				if are_function_arguments_valid(code[0].children[1], symbol_table.functions[function_index], symbol_table) {
					return symbol_table.functions[function_index].Type
				}
				return make([]string, 0)
			}
			return make([]string, 0)
		}
		if len(code)>1 {
			for i := 0; i < len(code); i++ {
				if i+2>=len(code) {
					return current_type
				}
				lhs:=current_type
				if string_arr_compare(make([]string, 0), current_type) {
					lhs=evaluate_type(symbol_table, []Token{code[i]})
				}
				rhs:=evaluate_type(symbol_table, []Token{code[i+2]})
				if string_arr_compare(lhs, make([]string, 0)) {
					return make([]string, 0)
				}
				if string_arr_compare(rhs, make([]string, 0)) {
					return make([]string, 0)
				}
				i++
				switch operator:=code[i]; operator.string_value {
				case "+","*":
					if lhs[0]=="[" && string_arr_compare(rhs,[]string{"num"}) {
						current_type=lhs
						continue
					}
					if rhs[0]=="[" && string_arr_compare(lhs,[]string{"num"}) {
						current_type=rhs
						continue
					}
					if operator.string_value=="*" {
						if (string_arr_compare(rhs, []string{"string"}) && string_arr_compare(lhs, []string{"num"})) || (string_arr_compare(rhs, []string{"num"}) && string_arr_compare(lhs, []string{"string"})) {
							if rhs[0]=="[" {
								current_type=rhs
							} else {
								current_type=lhs
							}
							continue
						}
					}
					if string_arr_compare(rhs, lhs) || (string_arr_compare(rhs, []string{"num"}) && string_arr_compare(lhs, []string{"string"})) {
						current_type=rhs
						continue
					}
					return make([]string, 0)
				case "-","/","//","^","**","%","&&","||":
					if (string_arr_compare(lhs, rhs) && string_arr_compare(lhs, []string{"num"})) {
						current_type=rhs
						continue
					}
					return make([]string, 0)
				case "==","!=",">","<":
					if !string_arr_compare(lhs, rhs) {
						return make([]string, 0)
					}
					if string_arr_compare(rhs,[]string{"string"}) {
						current_type=lhs
						continue
					}
					if string_arr_compare(rhs,[]string{"num"}) {
						current_type=lhs
						continue
					}
					return make([]string, 0)
				}
			}
		}
	} else {
		fmt.Println("invalid statement", code)
		return make([]string, 0)
	}
	return current_type
}

func compiler(symbol_table Symbol_Table, function_name string, depth int, code []Token) (string, Symbol_Table) {
	if depth==0 {
		for i := 0; i < len(symbol_table.functions); i++ {
			err,symbol_table:=compiler(symbol_table, symbol_table.functions[i].name, depth+1, make([]Token, 0))
			if symbol_table.operations[function_name]==nil {
				symbol_table.operations[function_name]=make([][]string, 0)
			}
			if err!="" {
				return err, symbol_table
			}
		}
	}
	if depth>0 {
		if len(function_name)!=0 {
			code=symbol_table.functions[function_index_in_symbol_table(function_name, symbol_table)].Code
		}
		for i := 0; i < len(code); i++ {
			if code[i].Type=="sys" && code[i].string_value=="var" && (len(code)-i)>=4 && code[i+1].Type=="variable" && valid_var_name(code[i+1].string_value) && valid_type(type_token_to_string_array(code[i+2]), symbol_table) && code[i+3].Type=="EOS" && variable_doesnot_exist(symbol_table, code[i+1].string_value) {
				symbol_table.variables = append(symbol_table.variables, Variable{name: code[i+1].string_value, Type: type_token_to_string_array(code[i+2])})
				symbol_table.variable_mapping[code[i+1].string_value]=Variable{name: function_name+"-"+code[i+1].string_value, Type: type_token_to_string_array(code[i+2])}
				
				// you need to add the compiler
				i+=3
				continue
			}
			if (len(code)-i)>=4 && (code[i].Type=="lookup" || code[i].Type=="variable" || code[i].Type=="nested_tokens") && code[i+1].Type=="operator" && code[i+1].string_value=="=" {
				i++
				i++
				tokens,err:=get_current_statement_tokens(code[i:])
				if err!=0 {
					return "unexpected end of statement", symbol_table
				}
				lhs:=evaluate_type(symbol_table, []Token{code[i-2]})
				rhs:=evaluate_type(symbol_table, tokens)
				if string_arr_compare(lhs, []string{}) {
					return "invalid type on lhs", symbol_table
				}
				if string_arr_compare(rhs, []string{}) {
					return "invalid type on rhs", symbol_table
				}
				if !string_arr_compare(lhs, rhs) {
					return "types on lhs and rhs do not match", symbol_table
				}
				i+=len(tokens)
				continue
			}
			if (code[i].Type=="branch") {
				if !string_arr_compare(evaluate_type(symbol_table, code[i].children[0].children), []string{"num"}) {
					return "branch condition is invalid", Symbol_Table{}
				}
				err,st:=compiler(symbol_table, "", depth+1, code[i].children[1].children)
				symbol_table=st
				if err!="" {
					return err, Symbol_Table{}
				}
				continue
			}
			if (code[i].Type=="while") {
				if !string_arr_compare(evaluate_type(symbol_table, code[i].children[0].children), []string{"num"}) {
					return "branch condition for while is invalid", Symbol_Table{}
				}
				err,st:=compiler(symbol_table, "", depth+1, code[i].children[1].children)
				symbol_table=st
				if err!="" {
					return err, Symbol_Table{}
				}
				continue
			}
			fmt.Println("missed token", code[i])
			return "error_token_not_parsed", symbol_table
		}
	}
	return "", symbol_table
}

func data_encoder(data []string) string {
	result:=""
	for _,data_item:=range data {
		result+="    \""+strings.ReplaceAll(data_item, "\n", "\\n")+"\""+"    \n"
	}
	return strings.Trim(result, "\n")
}

func build(tokens []Token, depth int) string {
	symbol_table:=Symbol_Table{operations: make(map[string][][]string), used_variables: make(map[string]int), variable_mapping: make(map[string]Variable)}
	err,symbol_table:=pre_parser(symbol_table, tokens, 0)
	if err!="" {
		return "Error: "+err
	}
	if !does_function_exist("main", symbol_table) {
		return "Error: "+"main function does not exist"
	}
	main_function:=symbol_table.functions[function_index_in_symbol_table("main", symbol_table)]
	if len(main_function.args)!=0 {
		return "Error: "+"main function cant have arguments"
	}
	if len(main_function.Type)!=0 {
		return "Error: "+"main function cannot have a type definition"
	}
	result:=""
	err,final_symbol_table:=compiler(symbol_table, "main", 0, make([]Token, 0))
	if err!="" {
		return "Error: "+err
	}
	fmt.Println(final_symbol_table)
	return ".code\n"+result+".data\n"+data_encoder(symbol_table.data)
}