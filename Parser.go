package main

import (
	"errors"
	"strings"
)

func Tokenizer(code string) ([]string, error) {
	tokens:=make([]string, 0)
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
				return make([]string, 0), errors.New("Unexpected EOF")
			}
		}
		if char=="\"" {
			if in_string {
				tokens = append(tokens, cache)
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
		if char==" " {
			if len(cache)!=0 {
				tokens = append(tokens, cache)
				cache=""
			}
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
							tokens = append(tokens, cache)
							cache=""
						}
						tokens = append(tokens, char)
						continue
					}
					cache+=char
					continue
				} else {
					if len(cache)!=0 {
						tokens = append(tokens, cache)
						cache=""
					}
					tokens = append(tokens, char)
					continue
				}
			}
		} else if in_number {
			if len(cache)!=0 {
				tokens = append(tokens, cache)
				cache=""
			}
			in_number=false
			i--
			continue
		}
		if strings.Contains("()*&^%$#@!~`-=+[{]}';:,<>/?]\\|", char) {
			if len(cache)!=0 {
				tokens = append(tokens, cache)
				cache=""
			}
			tokens = append(tokens, char)
			continue
		}
		if char=="\n" {
			if len(cache)!=0 {
				tokens = append(tokens, cache)
				cache=""
			}
			continue
		}
		cache+=char
	}
	if cache!="" {
		tokens = append(tokens, cache)
	}
	return tokens, nil
}