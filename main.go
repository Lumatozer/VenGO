package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args)<2 {
		fmt.Println("Run this binary with the format:\nvengine target.file --flags=value")
		return
	}
	data,err:=os.ReadFile(os.Args[1])
	if err!=nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Tokenizer")
	tokens,err:=Tokenizer(string(data))
	if err!=nil {
		fmt.Println(err)
		return
	}
	fmt.Println(tokens)
	Parse_Program(tokens, []string{})
}