package main

import (
	"errors"
	"strings"
)

type Group struct {

}

type Token struct {
	Type  		  string
	Value 		  string
	Str_Children  []string
	Tok_Children  []string
}

var Keywords []string=[]string{
	"var", "function", "set", "return","import", "as",
}

var Primitive_Types []string=[]string{
	"bytes", "int", "int64", "float", "float64", "string",
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
					if len(cache)!=0 {
						tokens = append(tokens, Token{Type: "sys", Value: cache})
						cache=""
					}
					tokens = append(tokens, Token{Type: "sys", Value: char})
					continue
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
	return filtered_tokens, nil
}

func Grouper(code []Token) ([]Group, error) {
	result:=make([]Group, 0)
	for i:=0; i<len(code); i++ {
		
	}
	return result, nil
}