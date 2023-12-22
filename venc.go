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
	args		map[string][]string
	Type		[]string
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
	functions 	[]Function
	structs   	[]Struct
	variables	[]Variable
	data		[]string
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
			accepted_combined_operator_array:=[]string{"+=","-=","/=","*=","%=","//","!=","==","->",":="}
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
		if len(parsed_tokens)>0 && parsed_tokens[len(parsed_tokens)-1].Type=="operator" && parsed_tokens[len(parsed_tokens)-1].string_value=="->" && ((current_token.string_value=="string" || current_token.string_value=="num") || (len(code)>i+1 && len(parsed_tokens)!=0 && current_token.Type=="bracket_open" && (current_token.string_value=="[" || current_token.string_value=="{"))) {
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
		if len(grouped_tokens)>0 && code[i].Type=="expression" && (grouped_tokens[len(grouped_tokens)-1].Type=="variable" || grouped_tokens[len(grouped_tokens)-1].Type=="nested_tokens" || grouped_tokens[len(grouped_tokens)-1].Type=="lookup" || grouped_tokens[len(grouped_tokens)-1].Type=="expression" || grouped_tokens[len(grouped_tokens)-1].Type=="funcall") {
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
		if symbol_table.structs[i].name==variable {
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

func valid_type(Type []string) bool {
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
	if str_index_in_arr(Type[(len(Type)-1)/2], valid_final_types)!=-1 {
		return true
	}
	return false
}

func internal(symbol_table Symbol_Table, code []Token, depth int) (string, Symbol_Table) {
	result:=""
	for i := 0; i < len(code); i++ {
		fmt.Println(code[i])
		if code[i].Type=="sys" && code[i].string_value=="struct" && len(code)>i+2 && code[i+1].Type=="variable" && valid_var_name(code[i+1].string_value) && code[i+2].Type=="bracket_open" && code[i+2].string_value=="{" {
			if !variable_doesnot_exist(symbol_table ,code[i+1].string_value) {
				return "error", Symbol_Table{}
			}
			if depth!=0 {
				return "depth_error", Symbol_Table{}
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
				if tokens[index].Type=="variable" && valid_var_name(tokens[index].string_value) && variable_doesnot_exist(symbol_table, tokens[index].string_value) && valid_type(variable_type) {
					Struct_Variables[tokens[index].string_value]=variable_type
				} else {
					return "struct_error", Symbol_Table{}
				}
			}
			symbol_table.structs = append(symbol_table.structs, Struct{name: code[i-1].string_value, fields: Struct_Variables})
			i+=len(tokens)+1
			continue
		}
		if code[i].Type=="sys" && code[i].string_value=="function" && len(code)>i+1 && code[i+1].Type=="funcall" {
			function_arguments:=make(map[string][]string)
			if len(code[i+1].children[1].children)%2!=0 {
				return "function_error", Symbol_Table{}
			}
			if !valid_var_name(code[i+1].children[0].string_value) {
				return "function_error_identifier", Symbol_Table{}
			}
			i++
			i++
			for index:=0; index<len(code[i-1].children[1].children); index+=2 {
				function_arguments[code[i-1].children[1].children[index].string_value]=type_token_to_string_array(code[i-1].children[1].children[index+1])
			}
			symbol_table.functions = append(symbol_table.functions, Function{args: function_arguments, name: code[i-1].children[0].string_value})
		}
	}
	return result, symbol_table
}

func data_encoder(data []string) string {
	result:=""
	for _,data_item:=range data {
		result+="\""+strings.ReplaceAll(data_item, "\n", "\\n")+"\""+"    \n"
	}
	return strings.Trim(result, "\n")
}

func compiler(tokens []Token, depth int) string {
	symbol_table:=Symbol_Table{}
	result,symbol_table:=internal(symbol_table, tokens, 0)
	fmt.Println(symbol_table)
	return ".code\n"+result+".data\n"+data_encoder(symbol_table.data)
}