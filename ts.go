package main

import (
	"fmt"
	"regexp"
	"strings"
)

func ts_translator(input string) string {
	keywords := []string{"let", "const", "var", "function", "class", "interface", "type", "enum"}
	identifiers := regexp.MustCompile(`[a-zA-Z_]\w*`)
	operators := regexp.MustCompile(`(\+|-|\*|\/|=|==|!=|>|<|>=|<=|\.)`)
	literals := regexp.MustCompile(`("[^"]*"|'[^']*'|\d+(\.\d+)?)`)
	colon := regexp.MustCompile(`:`)
	delimiters := regexp.MustCompile(`[{}()[\],;]`)
	returnType := regexp.MustCompile(`=>`)
	lines := strings.Split(input, "\n")
	var cleanLines []string
	for _, line := range lines {
		if !strings.Contains(line, "//") && !strings.Contains(line, "/*") && !strings.Contains(line, "*/") {
			cleanLines = append(cleanLines, line)
		}
	}
	cleanCode := strings.Join(cleanLines, "\n")
	pattern := fmt.Sprintf("(%s)|(%s)|(%s)|(%s)|(%s)|(%s)|(%s)", strings.Join(keywords, "|"), identifiers, operators, literals, colon, delimiters, returnType)
	tokenPattern := regexp.MustCompile(pattern)
	tokens := tokenPattern.FindAllString(cleanCode, -1)
	new_tokens:=make([]string, 0)
	for token:=0; token < len(tokens); token++ {
		if len(tokens)>=token+2 && tokens[token]=="console" && tokens[token+1]=="." && tokens[token+2]=="log" {
			new_tokens = append(new_tokens, "print")
			token+=2
			continue
		}
		if len(tokens)>=token+1 && tokens[token]=="!=" && tokens[token+1]=="="{
			new_tokens = append(new_tokens, "!=")
			token+=1
			continue
		}
		if len(tokens)>=token+2 && tokens[token]=="=" && tokens[token+1]=="=" && tokens[token+2]=="=" {
			new_tokens = append(new_tokens, "==")
			token+=2
			continue
		}
		if tokens[token] == ":" {
			tokens[token] = "->"
		}
		if tokens[token] == "interface" {
			tokens[token] = "struct"
		}
		if tokens[token] == "number" {
			tokens[token] = "num"
		}
		if tokens[token] == "console.log" {
			tokens[token] = "print"
		}
		if tokens[token] == "let" || tokens[token] == "const" {
			tokens[token] = "var"
		}
		new_tokens = append(new_tokens, tokens[token])
	}
	return strings.Join(new_tokens, " ")
}
