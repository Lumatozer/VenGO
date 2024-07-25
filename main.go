package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lumatozer/VenGO/structs"
	"github.com/lumatozer/VenGO/venc"
)

func Load_Packages(program *Program, packages []structs.Package) {
	if program.Is_Dynamic {
		for _,Package:=range packages {
			if Package.Name==program.Package_Name {
				for i,function:=range program.Functions {
					external_Function,ok:=Package.Functions[function.Name]
					if ok {
						*program.Functions[i].External_Function=external_Function
					}
				}
			}
		}
	}
	for _,Function:=range program.Functions {
		if Function.Base_Program!=program {
			Load_Packages(Function.Base_Program, packages)
		}
	}
}

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
	if strings.HasSuffix(os.Args[1] ,".vi") {
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
		program,err:=venc.Parser(os.Args[1], definitions)
		if err!=nil {
			fmt.Println(err)
			return
		}
		fmt.Println(program)
		return
	}
	tokens,err:=Tokenizer(string(data))
	if err!=nil {
		fmt.Println(err)
		return
	}
	fmt.Println(tokens)
	Absolute_Path,err:=filepath.Abs(os.Args[1])
	if err!=nil {
		fmt.Println(err)
		return
	}
	program,err:=Parser(tokens, Absolute_Path)
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
	Load_Packages(&program, Get_Packages())
	exec_Result:=Interpreter(&program.Functions[index], Stack{})
	if exec_Result.Error!=nil {
		fmt.Println(exec_Result.Return_Value)
		return
	}
}