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

var reserved_tokens = []string{"var", "fn", "if", "while", "continue", "break", "struct","return"}
var type_tokens = []string{"str", "number"}
var operators = []string{"+","-","*","/","^",">","<","=","&","!","|","%"}
var end_of_statements = []string{";"}
var brackets = []string{"(",")","[","]","{","}"}
var string_quotes = []string{"\"","'"}
var comma = ","
var allowed_variable_character="abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
var comments = []string{"#"}

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
					tokens = append(tokens, Token{Type: "variable", string_value: cache})
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
	for _,char:=range "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_" {
		var_name=strings.ReplaceAll(var_name,string(char),"")
	}
	return var_name==""
}

func tokens_parser(code []Token, debug bool) ([]Token, error) {
	parsed_tokens:=make([]Token,0)
	for i := 0; i < len(code); i++ {
		current_token:=code[i]
		if len(code)>i+1 && current_token.Type=="variable" && valid_var_name(current_token.string_value) {
			if code[i+1].Type=="bracket_open" && code[i+1].string_value=="[" {
				i++
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
				parsed_tokens = append(parsed_tokens, Token{Type: "lookup", children: childrentokens})
				continue
			}
		}
		if len(code)>i+1 && current_token.Type=="variable" && valid_var_name(current_token.string_value) && code[i+1].Type=="dot" {
			nested_variables:=make([]Token,0)
			first:=true
			for {
				if len(code)>i+1 && current_token.Type=="variable" && valid_var_name(current_token.string_value) && (code[i+1].Type=="dot" || !first) { 
					nested_variables = append(nested_variables, code[i])
					if code[i+1].Type=="dot" {
						i++
					}
					i++
					first=false
				} else {
					break
				}
			}
			parsed_tokens = append(parsed_tokens, Token{Type: "nested_tokens", children: nested_variables})
			continue
		}
		if len(code)>i+1 && current_token.Type=="operator" && code[i+1].Type=="operator" {
			combined_operator:=current_token.string_value+code[i+1].string_value
			accepted_combined_operator_array:=[]string{"+=","-=","/=","*=","%=","//","!=","=="}
			if str_index_in_arr(combined_operator,accepted_combined_operator_array)!=-1 {
				parsed_tokens = append(parsed_tokens, Token{Type: "operator", string_value: combined_operator})
				i++
				continue
			}
		}
		parsed_tokens = append(parsed_tokens, current_token)
	}
	return parsed_tokens, nil
}