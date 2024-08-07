package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/lumatozer/VenGO/venc"
	"github.com/lumatozer/VenGO"
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
	Absolute_Path,err:=filepath.Abs(os.Args[1])
	if strings.HasSuffix(os.Args[1], ".vi") {
		tokens:=venc.Tokensier(string(data), false)
		tokens,err:=venc.Tokens_Parser(tokens, false)
		if err!=nil {
			fmt.Println(err)
			return
		}
		tokens,err=venc.Token_Grouper(tokens, false)
		if err!=nil {
			fmt.Println(err)
			return
		}
		fmt.Println(tokens)
		definitions,err:=venc.Definition_Parser(tokens)
		if err!=nil {
			fmt.Println(err)
			return
		}
		fmt.Println(definitions)
		program,err:=venc.Parser(os.Args[1], definitions, make(map[string]venc.Program), vengine.VASM_Translator)
		if err!=nil {
			fmt.Println(err)
			return
		}
		fmt.Println(program)
		venc.Compile_Program(&program)
		current_Dir,_:=os.Getwd()
		Absolute_Current_File_Path,_:=filepath.Abs(current_Dir)
		Absolute_Path=filepath.Join("distributable", strings.Replace(strings.TrimPrefix(program.Path, Absolute_Current_File_Path), ".vi", ".vasm", 1))
		file_Data,err:=os.ReadFile(Absolute_Path)
		data=file_Data
		if err!=nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(data))
	}
	tokens,err:=vengine.Tokenizer(string(data))
	if err!=nil {
		fmt.Println(err)
		return
	}
	fmt.Println(tokens)
	if err!=nil {
		fmt.Println(err)
		return
	}
	program,err:=vengine.Parser(tokens, Absolute_Path, make(map[string]vengine.Program))
	if err!=nil {
		fmt.Println(err)
		return
	}
	fmt.Println()
	index:=-1
	for i:=0; i<len(program.Functions); i++ {
		if program.Functions[i].Name=="main" {
			index=i
		}
	}
	vengine.Load_Packages(&program, vengine.Get_Packages())
	exec_Result:=vengine.Interpreter(&program.Functions[index], vengine.Stack{})
	if exec_Result.Error!=nil {
		fmt.Println(exec_Result.Error)
		return
	}
	if exec_Result.Return_Value!=nil {
		fmt.Println(exec_Result.Return_Value)
	}
}