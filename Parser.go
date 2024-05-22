package main

import (
	"errors"
	"strings"
)

type Group struct {
	Name               string
	Token_Children     []Token
	Mapped_Children    map[string][]Token
}

type Token struct {
	Type  		  string
	Value 		  string
	Str_Children  []string
	Tok_Children  []Token
}

var Keywords []string=[]string{
	"var", "function", "set", "return","import", "as",
}

var Primitive_Types []string=[]string{
	"bytes", "int", "int64", "float", "float64", "string",
}

func Is_Parsed_Type_Valid(parsed_type []string) bool {
	if len(parsed_type)%2!=1 {
		return false
	}
	if len(parsed_type)==1 && !strings.Contains("()[]{}", parsed_type[0]) {
		return true
	}
	if parsed_type[0]=="{" && parsed_type[len(parsed_type)-1]=="}" && len(parsed_type)>=5 && parsed_type[2]=="->" && str_index_in_str_arr(parsed_type[1], []string{"bytes", "int", "int64", "float", "float64", "string"})!=-1 {
		return Is_Parsed_Type_Valid(parsed_type[3:len(parsed_type)-1])
	}
	if parsed_type[0]=="[" && parsed_type[len(parsed_type)-1]=="]" && len(parsed_type)>=3 {
		return Is_Parsed_Type_Valid(parsed_type[1:len(parsed_type)-1])
	}
	return false
}

func Type_Parser(parsed_type []string) []Token {
	if len(parsed_type)==1 {
		return []Token{Token{Type: "raw", Value: parsed_type[0]}}
	}
	if parsed_type[0]=="[" {
		return []Token{Token{Type: "array", Tok_Children: Type_Parser(parsed_type[1:len(parsed_type)-1])}}
	}
	if parsed_type[0]=="{" {
		return []Token{Token{Type: "dict", Tok_Children: Type_Parser(parsed_type[3:len(parsed_type)-1]), Str_Children: []string{parsed_type[1]}}}
	}
	return []Token{Token{}}
}

func Tokenizer(code string) ([]Token, error) {
	tokens:=make([]Token, 0)
	in_string:=false
	cache:=""
	in_number:=false
	for i:=0; i<len(code); i++ {
		char:=string(code[i])
		if char=="\r" {
			continue
		}
		if char=="\\" && in_string {
			if len(code)>i+1 {
				if string(code[i+1])=="\\" || string(code[i+1])=="\"" {
					cache+=string(code[i+1])
					i+=1
					continue
				}
			} else {
				return make([]Token, 0), errors.New("Unexpected EOF")
			}
		}
		if char=="\"" {
			if in_string {
				tokens = append(tokens, Token{Type: "string", Value: cache})
				cache=""
				in_string=false
				continue
			} else {
				if cache=="" {
					in_string=true
					continue
				}
			}
		}
		if in_string {
			cache+=char
			continue
		}
		if strings.Contains("1234567890.", char) {
			if char!="." {
				cache+=char
				in_number=true
				continue
			} else {
				if in_number {
					if strings.Contains(cache, ".") {
						if len(cache)!=0 {
							tokens = append(tokens, Token{Type: "number", Value: cache})
							cache=""
						}
						tokens = append(tokens, Token{Type: "sys", Value: char})
						continue
					}
					cache+=char
					continue
				} else {
					// prevents a.b to become a dot b
					// if len(cache)!=0 {
					// 	tokens = append(tokens, Token{Type: "sys", Value: cache})
					// 	cache=""
					// }
					// tokens = append(tokens, Token{Type: "sys", Value: char})
					// continue
				}
			}
		} else if in_number {
			if len(cache)!=0 {
				tokens = append(tokens, Token{Type: "number", Value: cache})
				cache=""
			}
			in_number=false
			i--
			continue
		}
		if char==" " {
			if len(cache)!=0 {
				tokens = append(tokens, Token{Type: "sys", Value: cache})
				cache=""
			}
			continue
		}
		if strings.Contains("()*&^%$#@!~`-=+[{]}';:,<>/?]\\|", char) {
			if len(cache)!=0 {
				tokens = append(tokens, Token{Type: "sys", Value: cache})
				cache=""
			}
			tokens = append(tokens, Token{Type: "sys", Value: char})
			continue
		}
		if char=="\n" {
			if len(cache)!=0 {
				tokens = append(tokens, Token{Type: "sys", Value: cache})
				cache=""
			}
			continue
		}
		cache+=char
	}
	if cache!="" {
		tokens = append(tokens, Token{Type: "sys", Value: cache})
	}
	filtered_tokens:=make([]Token, 0)
	for i:=0; i<len(tokens); i++ {
		if tokens[i].Type=="sys" {
			if str_index_in_str_arr(tokens[i].Value, Keywords)!=-1 {
				filtered_tokens = append(filtered_tokens, tokens[i])
				continue
			}
			if str_index_in_str_arr(tokens[i].Value, Primitive_Types)!=-1 {
				filtered_tokens = append(filtered_tokens, Token{Type: "type", Value: tokens[i].Value})
				continue
			}
			if tokens[i].Value=="," {
				filtered_tokens = append(filtered_tokens, Token{Type: "comma", Value: tokens[i].Value})
				continue
			}
			if tokens[i].Value=="." {
				filtered_tokens = append(filtered_tokens, Token{Type: "dot", Value: tokens[i].Value})
				continue
			}
			if tokens[i].Value==";" {
				filtered_tokens = append(filtered_tokens, Token{Type: "semicolon", Value: tokens[i].Value})
				continue
			}
			if tokens[i].Value==":" {
				filtered_tokens = append(filtered_tokens, Token{Type: "colon", Value: tokens[i].Value})
				continue
			}
			if tokens[i].Value==">" && i>=1 && tokens[i-1].Value=="-" {
				filtered_tokens[len(filtered_tokens)-1]=Token{Type: "arrow", Value: "->"}
				continue
			}
			if strings.Contains("(){}[]", tokens[i].Value) {
				filtered_tokens = append(filtered_tokens, Token{Type: "bracket", Value: tokens[i].Value})
				continue
			}
			tokens[i].Type="variable"
		}
		filtered_tokens = append(filtered_tokens, tokens[i])
	}
	filtered_tokens_2:=make([]Token, 0)
	for i:=0; i<len(filtered_tokens); i++ {
		if filtered_tokens[i].Type=="arrow" {
			j:=i
			arr_count:=0
			dict_count:=0
			type_strings:=make([]string, 0)
			for {
				j++
				if j==len(filtered_tokens) {
					return make([]Token, 0), errors.New("Unexpected EOF of type definition")
				}
				if filtered_tokens[j].Type=="bracket" {
					if filtered_tokens[j].Value=="{" {
						dict_count+=1
					}
					if filtered_tokens[j].Value=="}" {
						dict_count-=1
					}
					if filtered_tokens[j].Value=="[" {
						arr_count+=1
					}
					if filtered_tokens[j].Value=="]" {
						arr_count-=1
					}
					type_strings = append(type_strings, filtered_tokens[j].Value)
				}
				if filtered_tokens[j].Type!="bracket" {
					type_strings = append(type_strings, filtered_tokens[j].Value)
				}
				if arr_count==0 && dict_count==0 {
					break
				}
			}
			if !Is_Parsed_Type_Valid(type_strings) {
				return make([]Token, 0), errors.New("Invalid Type specification")
			}
			filtered_tokens_2 = append(filtered_tokens_2, Token{Type: "type", Tok_Children: Type_Parser(type_strings)})
			i+=j-i
			continue
		}
		filtered_tokens_2 = append(filtered_tokens_2, filtered_tokens[i])
	}
	return filtered_tokens_2, nil
}

func Grouper(code []Token) ([]Group, error) {
	result:=make([]Group, 0)
	for i:=0; i<len(code); i++ {
		
	}
	return result, nil
}