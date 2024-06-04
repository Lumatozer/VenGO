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
	tokens,err:=Tokenizer(string(data))
	if err!=nil {
		fmt.Println(err)
		return
	}
	fmt.Println(tokens)
	// _,err=Parser(tokens)
	// if err!=nil {
	// 	fmt.Println(err)
	// }
	// program,err:=Parse_Program(tokens, []string{}, os.Args[1], nil)
	// entry_function:=-1
	// if err!=nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// for i,function:=range program.Functions {
	// 	if function.Name=="_start" {
	// 		entry_function=i
	// 		break
	// 	}
	// }
	// if entry_function==-1 {
	// 	fmt.Println("Entry file has no _start function")
	// 	return
	// }
	// Interpreter(program.Functions[entry_function])
}