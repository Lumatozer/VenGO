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
		if len(code)>i+1 && code[i+1].Type=="dot" {
			nested_variables:=make([]Token,0)
			first:=true
			for {
				if len(code)>i+1 && (!first || code[i+1].Type=="dot") { 
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
		tokens_children,err:=token_grouper(code[i].children, false)
		if err!=nil {
			return make([]Token, 0), err
		}
		code[i].children=tokens_children
		if code[i].Type=="expression_wrapper_[]" {
			if len(grouped_tokens)>0 && (grouped_tokens[len(grouped_tokens)-1].Type=="variable" || grouped_tokens[len(grouped_tokens)-1].Type=="nested_tokens" || grouped_tokens[len(grouped_tokens)-1].Type=="lookup" || grouped_tokens[len(grouped_tokens)-1].Type=="expression" || grouped_tokens[len(grouped_tokens)-1].Type=="funcall" || grouped_tokens[len(grouped_tokens)-1].Type=="array") {
				grouped_tokens[len(grouped_tokens)-1]=Token{Type: "lookup", children: []Token{Token{Type: "parent", children: []Token{grouped_tokens[len(grouped_tokens)-1]}}, Token{Type: "tokens", children: code[i].children}}}
				continue
			}
		}
		if len(grouped_tokens)>0 && code[i].Type=="type" && grouped_tokens[len(grouped_tokens)-1].Type=="expression_wrapper_[]" {
			grouped_tokens[len(grouped_tokens)-1]=Token{Type: "array", children: []Token{code[i], grouped_tokens[len(grouped_tokens)-1]}}
			continue
		}
		if len(grouped_tokens)>1 && grouped_tokens[len(grouped_tokens)-1].Type=="dot" && (grouped_tokens[len(grouped_tokens)-2].Type=="nested_tokens" || grouped_tokens[len(grouped_tokens)-2].Type=="lookup" || grouped_tokens[len(grouped_tokens)-2].Type=="variable" || grouped_tokens[len(grouped_tokens)-2].Type=="expression" || grouped_tokens[len(grouped_tokens)-2].Type=="funcall") {
			grouped_tokens[len(grouped_tokens)-2]=Token{Type: "nested_tokens", children: []Token{grouped_tokens[len(grouped_tokens)-2], code[i]}}
			grouped_tokens=grouped_tokens[:len(grouped_tokens)-1]
			continue
		}
		if len(grouped_tokens)>0 && code[i].Type=="expression" {
			if (grouped_tokens[len(grouped_tokens)-1].Type=="variable" || grouped_tokens[len(grouped_tokens)-1].Type=="lookup" || grouped_tokens[len(grouped_tokens)-1].Type=="expression" || grouped_tokens[len(grouped_tokens)-1].Type=="funcall" || grouped_tokens[len(grouped_tokens)-1].Type=="nested_tokens") {
				grouped_tokens[len(grouped_tokens)-1]=Token{Type: "funcall", children: []Token{grouped_tokens[len(grouped_tokens)-1], code[i]}}
				continue
			}
		}
		grouped_tokens = append(grouped_tokens, code[i])
	}
	return grouped_tokens, nil
}

func deep_check(tokens []Token) bool  {
	for _,token:=range tokens {
		if token.Type=="nested_tokens" {
			if len(token.children)<2 {
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

func calculate_i_skip(code []Token) int {
	total:=0
	for i := 0; i < len(code); i++ {
		if code[i].Type=="branch" || code[i].Type=="while" {
			total+=4
			total+=calculate_i_skip(code[i].children[1].children)
		} else {
			total+=1
		}
	}
	return total
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
			i+=3+calculate_i_skip(tokens)
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
			i+=3+calculate_i_skip(tokens)
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
					return "error", symbol_table
				}
				i++
				i++
				tokens,err:=bracket_token_getter(code[i:], code[i].string_value)
				if err!=0 {
					return "error", symbol_table
				}
				tokens=tokens[1:len(tokens)-1]
				if len(tokens)%2!=0 {
					return "error", symbol_table
				}
				Struct_Variables:=make(map[string][]string)
				for index:=0; int64(index) < int64(len(tokens))-1; index+=2 {
					variable_type:=type_token_to_string_array(tokens[index+1])
					if tokens[index].Type=="variable" && valid_var_name(tokens[index].string_value) && valid_type(variable_type, symbol_table) {
						Struct_Variables[tokens[index].string_value]=variable_type
					} else {
						return "struct_error", symbol_table
					}
				}
				data_index:=str_index_in_arr(code[i-1].string_value, symbol_table.data)
				if data_index==-1 {
					symbol_table.data = append(symbol_table.data, code[i-1].string_value)
					data_index=len(symbol_table.data)-1
				}
				fields_converted:=make([]string, 0)
				for key,val:=range Struct_Variables {
					fields_converted = append(fields_converted, key+"->"+strings.Join(string_array_types_to_vitality_types(val), ","))
				}
				data_index=str_index_in_arr(strings.Join(fields_converted, ";"), symbol_table.data)
				if data_index==-1 {
					symbol_table.data = append(symbol_table.data, strings.Join(fields_converted, ";"))
					data_index=len(symbol_table.data)-1
				}
				symbol_table.operations["main"] = append(symbol_table.operations["main"], []string{"register_struct", code[i-1].string_value, strconv.FormatInt(int64(data_index), 10)})
				symbol_table.structs = append(symbol_table.structs, Struct{name: code[i-1].string_value, fields: Struct_Variables})
				i+=len(tokens)+1
				continue
			}
			if code[i].Type=="sys" && code[i].string_value=="function" && len(code)>i+1 && code[i+1].Type=="funcall" && variable_doesnot_exist(symbol_table ,code[i+1].children[0].string_value) {
				function_name:=code[i+1].children[0].string_value
				function_arguments:=make([][]string, 0)
				if len(code[i+1].children[1].children)%2!=0 {
					return "function_error", symbol_table
				}
				if !valid_var_name(code[i+1].children[0].string_value) {
					return "function_error_identifier", symbol_table
				}
				i++
				i++
				for index:=0; index<len(code[i-1].children[1].children); index+=2 {
					function_arguments=append(function_arguments, type_token_to_string_array(code[i-1].children[1].children[index+1]))
					if !valid_type(function_arguments[len(function_arguments)-1], symbol_table) {
						return "invalid function argument definition", symbol_table
					}
				}
				function_type:=make([]string, 0)
				if code[i].Type=="type" {
					function_type=type_token_to_string_array(code[i])
					if !valid_type(function_type, symbol_table) {
						return "function_return_type_is_invalid", symbol_table
					}
				}
				if code[i].Type=="type" {
					i++
				}
				tokens,err:=bracket_token_getter(code[i:], code[i].string_value)
				if err!=0 {
					return "error", symbol_table
				}
				to_ignore:=len(tokens[1:len(tokens)-1])
				err_,tokens:=branch_parser(tokens[1:len(tokens)-1])
				if err_!="" {
					return "error", symbol_table
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
		fmt.Println(code[i])
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
			if str_index_in_arr(token.Type, []string{"lookup", "variable", "expression", "string", "num", "funcall", "nested_tokens", "array"})==-1 {
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
	if len(argument_expression.children)%2!=1 && len(argument_expression.children)!=0 {
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
		if !string_arr_compare(evaluate_type(symbol_table, arguments[i].children, 0), function.args[i]) {
			return false
		}
	}
	return true
}

func are_array_arguments_valid(argument_expression Token, Type []string, symbol_table Symbol_Table) bool {
	if len(argument_expression.children)==0 {
		return true
	}
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
	for i := 0; i < len(arguments); i++ {
		if len(Type)<3 {
			return false
		}
		if !string_arr_compare(evaluate_type(symbol_table, arguments[i].children, 0), Type[1:len(Type)-1]) {
			return false
		}
	}
	return true
}

func nested_puller(code []Token) []Token {
	grouped_tokens:=make([]Token, 0)
	for i := 0; i < len(code); i++ {
		if code[i].Type=="nested_tokens" {
			new_children:=make([]Token, 0)
			for j := 0; j < len(code[i].children); j++ {
				if code[i].children[j].Type=="nested_tokens" {
					new_children = append(new_children, nested_puller(code[i].children[j].children)...)
					continue
				}
				new_children = append(new_children, code[i].children[j])
			}
			code[i].children=new_children
			grouped_tokens = append(grouped_tokens, code[i])
			continue
		}
		code[i].children=nested_puller(code[i].children)
		grouped_tokens = append(grouped_tokens, code[i])
	}
	return grouped_tokens
}

func evaluate_type(symbol_table Symbol_Table, code []Token, depth int) ([]string) {
	if len(code)==0 {
		return make([]string, 0)
	}
	current_type:=make([]string, 0)
	if (is_valid_statement(symbol_table, code) || depth!=0) {
		if len(code)==1 {
			if code[0].Type=="string" {
				return []string{"string"}
			}
			if code[0].Type=="num" {
				return []string{"num"}
			}
			if code[0].Type=="expression" {
				return evaluate_type(symbol_table, code[0].children, 0)
			}
			if code[0].Type=="variable" {
				variable_index:=variable_index_in_symbol_table(code[0].string_value,symbol_table)
				if variable_index==-1 {
					return make([]string, 0)
				}
				return symbol_table.variables[variable_index].Type
			}
			if code[0].Type=="array" {
				if !valid_type(type_token_to_string_array(code[0].children[0]), symbol_table) {
					return make([]string, 0)
				}
				if are_array_arguments_valid(code[0].children[1], type_token_to_string_array(code[0].children[0]), symbol_table) {
					return type_token_to_string_array(code[0].children[0])
				}
				return make([]string, 0)
			}
			if code[0].Type=="lookup" {
				parent_type:=evaluate_type(symbol_table, []Token{code[0].children[0].children[0]}, 0)
				if len(parent_type)==0 {
					return make([]string, 0)
				}
				if str_index_in_arr(parent_type[0], []string{"[","{"})==-1 {
					return make([]string, 0)
				}
				lookup_type:=evaluate_type(symbol_table, code[0].children[1].children, 0)
				if len(lookup_type)==0 {
					return make([]string, 0)
				}
				if parent_type[0]=="[" && string_arr_compare(lookup_type, []string{"num"}) {
					return parent_type[1:len(parent_type)-1]
				}
				if parent_type[0]=="{" && string_arr_compare(lookup_type, []string{"string"}) {
					return parent_type[1:len(parent_type)-1]
				}
				return make([]string, 0)
			}
			if code[0].Type=="nested_tokens" {
				parent_type:=make([]string, 0)
				if depth==0 {
					parent_type=evaluate_type(symbol_table, []Token{code[0].children[0]}, 0)
					if len(parent_type)==0 {
						return make([]string, 0)
					}
				} else {
					parent_type=[]string{code[0].children[0].string_value}
				}
				if len(parent_type)==1 && struct_index_in_symbol_table(parent_type[0], symbol_table)!=-1 {
					struct_index:=struct_index_in_symbol_table(parent_type[0], symbol_table)
					res:=symbol_table.structs[struct_index].fields[code[0].children[1].string_value]
					if len(res)==0 {
						return make([]string, 0)
					}
					if len(code[0].children)>2 {
						if struct_index_in_symbol_table(res[0], symbol_table)==-1 {
							return make([]string, 0)
						}
						code[0].children[1]=Token{string_value: res[0]}
						code_:=code
						code_[0].children=code[0].children[1:]
						return evaluate_type(symbol_table, code_, 1)
					} else {
						if depth==0 {
							return res
						} else {
							return res
						}
					}
					return make([]string, 0)
				}
				return make([]string, 0)
			}
			if code[0].Type=="funcall" {
				fmt.Println("hi", code[0].children[0].Type=="variable")
				if code[0].children[0].Type=="variable" {
					function_index:=function_index_in_symbol_table(code[0].children[0].string_value, symbol_table)
					fmt.Println(function_index)
					if function_index==-1 {
						return make([]string, 0)
					}
					fmt.Println(are_function_arguments_valid(code[0].children[1], symbol_table.functions[function_index], symbol_table) )
					if are_function_arguments_valid(code[0].children[1], symbol_table.functions[function_index], symbol_table) {
						fmt.Println("hi")
						return symbol_table.functions[function_index].Type
					}
				}
				if code[0].children[0].Type=="nested_tokens" {
					code_:=code[0].children[0]
					code_.children=code_.children[:len(code[0].children[0].children)-1]
					if code_.Type=="nested_tokens" && len(code_.children)==1 {
						code_=code_.children[0]
					}
					parent_type:=evaluate_type(symbol_table, []Token{code_}, 0)
					if len(parent_type)==0 {
						return make([]string, 0)
					}
					function_name:=code[0].children[0].children[len(code[0].children[0].children)-1].string_value
					arguments:=code[0].children[1]
					switch function_name {
					case "replace":
						if !string_arr_compare(parent_type, []string{"string"}) {
							return make([]string, 0)
						}
						if !are_function_arguments_valid(arguments, Function{args: [][]string{
							[]string{"string"}, 
							[]string{"string"},
							}}, symbol_table) {
							return make([]string, 0)
						}
						return []string{"string"}
					case "includes":
						if !string_arr_compare(parent_type, []string{"string"}) {
							return make([]string, 0)
						}
						if !are_function_arguments_valid(arguments, Function{args: [][]string{
							[]string{"string"}, 
							[]string{"string"},
							}}, symbol_table) {
							return make([]string, 0)
						}
						return []string{"number"}
					case "index":
						if !string_arr_compare(parent_type, []string{"string"}) {
							return make([]string, 0)
						}
						if !are_function_arguments_valid(arguments, Function{args: [][]string{
							[]string{"string"}, 
							[]string{"string"},
							}}, symbol_table) {
							return make([]string, 0)
						}
						return []string{"number"}
					case "split":
						if !string_arr_compare(parent_type, []string{"string"}) {
							return make([]string, 0)
						}
						if !are_function_arguments_valid(arguments, Function{args: [][]string{
							[]string{"string"},
							}}, symbol_table) {
							return make([]string, 0)
						}
						return []string{"[","string","]"}
					case "string":
						if !string_arr_compare(parent_type, []string{"num"}) {
							return make([]string, 0)
						}
						if len(arguments.children)!=0 {
							return make([]string, 0)
						}
						return []string{"string"}
					case "num":
						if !string_arr_compare(parent_type, []string{"string"}) {
							return make([]string, 0)
						}
						if len(arguments.children)!=1 {
							return make([]string, 0)
						}
						if arguments.children[0].Type!="variable" || variable_index_in_symbol_table(arguments.children[0].string_value, symbol_table)==-1 {
							return make([]string, 0)
						}
						if !string_arr_compare(symbol_table.variables[variable_index_in_symbol_table(arguments.children[0].string_value, symbol_table)].Type, []string{"num"}) {
							return make([]string, 0)
						}
						return []string{"num"}
					}
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
					lhs=evaluate_type(symbol_table, []Token{code[i]}, 0)
				}
				rhs:=evaluate_type(symbol_table, []Token{code[i+2]}, 0)
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

func string_array_types_to_vitality_types(types []string) []string  {
	output:=make([]string, 0)
	if types[0]=="[" {
		output = append(output, "arr")
		for _,type_string:=range string_array_types_to_vitality_types(types[1:len(types)-1]) {
			output = append(output, type_string)
		}
	} else if types[0]=="{" {
		output = append(output, "dict")
		for _,type_string:=range string_array_types_to_vitality_types(types[1:len(types)-1]) {
			output = append(output, type_string)
		}
	} else {
		output = append(output, types[0])
	}
	return output
}

func compiler(symbol_table Symbol_Table, function_name string, depth int, code []Token, in_loop bool) (string, Symbol_Table) {
	if depth==0 {
		fmt.Println(symbol_table)
		for i := 0; i < len(symbol_table.functions); i++ {
			symbol_table.functions[i].Code=nested_puller(symbol_table.functions[i].Code)
			if symbol_table.operations[function_name]==nil {
				symbol_table.operations[function_name]=make([][]string, 0)
			}
			err,returned_symbol_table:=compiler(symbol_table, symbol_table.functions[i].name, depth+1, make([]Token, 0), in_loop)
			symbol_table=returned_symbol_table
			if err!="" {
				return err, symbol_table
			}
		}
		return "", symbol_table
	}
	if depth>0 {
		if len(function_name)!=0 && string(function_name[0])!="-"  {
			code=symbol_table.functions[function_index_in_symbol_table(function_name, symbol_table)].Code
		}
		if string(function_name[0])=="-" {
			function_name=function_name[1:]
		}
		for i := 0; i < len(code); i++ {
			if code[i].Type=="sys" && code[i].string_value=="var" && (len(code)-i)>=4 && code[i+1].Type=="variable" && valid_var_name(code[i+1].string_value) && valid_type(type_token_to_string_array(code[i+2]), symbol_table) && code[i+3].Type=="EOS" && variable_doesnot_exist(symbol_table, code[i+1].string_value) {
				symbol_table.variables = append(symbol_table.variables, Variable{name: code[i+1].string_value, Type: type_token_to_string_array(code[i+2])})
				symbol_table.variable_mapping[code[i+1].string_value]=Variable{name: function_name+"-"+code[i+1].string_value, Type: type_token_to_string_array(code[i+2])}
				variable_type:=type_token_to_string_array(code[i+2])
				if variable_type[0]=="string" {
					string_index:=str_index_in_arr("", symbol_table.data)
					if string_index==-1 {
						symbol_table.data = append(symbol_table.data, "")
						string_index=len(symbol_table.data)-1
					}
					symbol_table.operations[function_name]=append(symbol_table.operations[function_name], []string{"str.set", code[i+1].string_value, strconv.FormatInt(int64(string_index), 10)})
				} else if variable_type[0]=="num" {
					symbol_table.operations[function_name]=append(symbol_table.operations[function_name], []string{"set", code[i+1].string_value, "0"})
				} else if variable_type[0]=="[" {
					data_index:=str_index_in_arr(strings.Join(string_array_types_to_vitality_types(variable_type)[1:], ","), symbol_table.data)
					if data_index==-1 {
						symbol_table.data = append(symbol_table.data, strings.Join(string_array_types_to_vitality_types(variable_type)[1:], ","))
						data_index=len(symbol_table.data)-1
					}
					symbol_table.operations[function_name]=append(symbol_table.operations[function_name], []string{"arr.init", code[i+1].string_value, strconv.FormatInt(int64(data_index), 10)})
				} else if variable_type[0]=="{" {
					data_index:=str_index_in_arr(strings.Join(string_array_types_to_vitality_types(variable_type)[1:], ","), symbol_table.data)
					if data_index==-1 {
						symbol_table.data = append(symbol_table.data, strings.Join(string_array_types_to_vitality_types(variable_type)[1:], ","))
						data_index=len(symbol_table.data)-1
					}
					symbol_table.operations[function_name]=append(symbol_table.operations[function_name], []string{"dict.init", code[i+1].string_value, strconv.FormatInt(int64(data_index), 10)})
				} else if len(variable_type)==1 {
					if struct_index_in_symbol_table(variable_type[0], symbol_table)==-1 {
						return "Invalid variable initialisation", symbol_table
					}
					symbol_table.operations[function_name]=append(symbol_table.operations[function_name], []string{"struct.init", code[i+1].string_value, strconv.FormatInt(int64(str_index_in_arr(variable_type[0], symbol_table.data)), 10)})
				}
				// symbol_table.operations[function_name]=append(symbol_table.operations[function_name], )
				// you need to add the compiler
				i+=3
				continue
			}
			if (len(code)-i)>=4 && (code[i].Type=="lookup" || code[i].Type=="variable" || code[i].Type=="nested_tokens") && code[i+1].Type=="operator" && strings.Contains(code[i+1].string_value, "=") {
				i++
				i++
				tokens,err:=get_current_statement_tokens(code[i:])
				if err!=0 {
					return "unexpected end of statement", symbol_table
				}
				lhs:=evaluate_type(symbol_table, []Token{code[i-2]}, 0)
				rhs:=evaluate_type(symbol_table, tokens, 0)
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
				if !string_arr_compare(evaluate_type(symbol_table, code[i].children[0].children, 0), []string{"num"}) {
					return "branch condition is invalid", symbol_table
				}
				err,st:=compiler(symbol_table, "-"+function_name, depth+1, code[i].children[1].children, in_loop)
				symbol_table=st
				if err!="" {
					return err, symbol_table
				}
				continue
			}
			if (code[i].Type=="while") {
				if !string_arr_compare(evaluate_type(symbol_table, code[i].children[0].children, 0), []string{"num"}) {
					return "branch condition for while is invalid", symbol_table
				}
				err,st:=compiler(symbol_table, "-"+function_name, depth+1, code[i].children[1].children, true)
				symbol_table=st
				if err!="" {
					return err, symbol_table
				}
				continue
			}
			if code[i].string_value=="return" && code[i].Type=="sys" {
				tokens,err :=get_current_statement_tokens(code[i+1:])
				if err!=0 {
					return "Unexpected EOS", symbol_table
				}
				if len(tokens)==0 && len(symbol_table.functions[function_index_in_symbol_table(function_name, symbol_table)].Type)==0 {
					i+=1
					continue
				}
				returned_type:=evaluate_type(symbol_table, tokens, 0)
				if len(returned_type)==0 {
					return "Invalid return type", symbol_table
				}
				if !string_arr_compare(returned_type, symbol_table.functions[function_index_in_symbol_table(function_name, symbol_table)].Type) {
					return "Return type does not match", symbol_table
				}
				i+=len(tokens)+1
				continue
			}
			if code[i].string_value=="continue" && code[i].Type=="sys" {
				if !in_loop {
					return "cannot loop", symbol_table
				}
				i+=1
				continue
			}
			if code[i].string_value=="break" && code[i].Type=="sys" {
				if !in_loop {
					return "cannot break loop", symbol_table
				}
				i+=1
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

func tokens_to_string_constants(tokens []Token) []string {
	out:=make([]string, 0)
	for _,token:=range tokens {
		if len(token.children)!=0 {
			for _,string_constants:=range tokens_to_string_constants(token.children) {
				if str_index_in_arr(string_constants, out)==-1 {
					out = append(out, string_constants)
				}
			}
		}
		if token.Type=="string" {
			if str_index_in_arr(token.string_value, out)==-1 {
				out = append(out, token.string_value)
			}
		}
	}
	return out
}

func build(tokens []Token, depth int) string {
	symbol_table:=Symbol_Table{operations: make(map[string][][]string), used_variables: make(map[string]int), variable_mapping: make(map[string]Variable)}
	symbol_table.data = append(symbol_table.data, tokens_to_string_constants(tokens)...)
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
	err,final_symbol_table:=compiler(symbol_table, "main", 0, make([]Token, 0), false)
	symbol_table=final_symbol_table
	for function:=range symbol_table.operations {
		result+="define.jump "+strconv.FormatInt(int64(str_index_in_arr(function, symbol_table.data)), 10)+"\n"
		for _,operations:=range symbol_table.operations[function] {
			result+=operations[0]+" "+strings.Join(operations[1:], ",")+"\n"
		}
	}
	if err!="" {
		return "Error: "+err
	}
	fmt.Println(final_symbol_table)
	return ".code\n"+result+".data\n"+data_encoder(symbol_table.data)
}