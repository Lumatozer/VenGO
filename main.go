package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"github.com/lumatozer/VenGO"
	"github.com/lumatozer/VenGO/structs"
	"github.com/lumatozer/VenGO/venc"
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
	if err!=nil {
		fmt.Println(err)
		return
	}
	if strings.HasSuffix(os.Args[1], ".vi") {
		tokens:=Venc.Tokensier(string(data), false)
		tokens,err:=Venc.Tokens_Parser(tokens, false)
		if err!=nil {
			fmt.Println(err)
			return
		}
		tokens,err=Venc.Token_Grouper(tokens, false)
		if err!=nil {
			fmt.Println(err)
			return
		}
		fmt.Println(tokens)
		definitions,err:=Venc.Definition_Parser(tokens)
		if err!=nil {
			fmt.Println(err)
			return
		}
		fmt.Println(definitions)
		program,err:=Venc.Parser(os.Args[1], definitions, make(map[string]Venc.Program), Vengine.VASM_Translator)
		if err!=nil {
			fmt.Println(err)
			return
		}
		fmt.Println(program)
		Venc.Compile_Program(&program)
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
	tokens,err:=Vengine.Tokenizer(string(data))
	if err!=nil {
		fmt.Println(err)
		return
	}
	fmt.Println(tokens)
	program,err:=Vengine.Parser(tokens, Absolute_Path, make(map[string]Vengine.Program))
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
	Vengine.Load_Packages(&program, Vengine.Get_Packages())
	wg:=sync.WaitGroup{}
	threads:=make([]*structs.Mutex_Interface, 0)
	n:=2
	for i := 0; i < n; i++ {
		threads = append(threads, &structs.Mutex_Interface{Channel: make(chan int)})
	}
	for _,thread:=range threads {
		wg.Add(1)
		go func(thread *structs.Mutex_Interface) {
			exec_Result:=Vengine.Interpreter(&program.Functions[index], Vengine.Stack{}, thread, structs.Database_Interface{
				Locking_Databases: []int{1},
				DB_Read: DB_Read,
				DB_Write: DB_Write,
			})
			if exec_Result.Error!=nil {
				fmt.Println(exec_Result.Error)
				return
			}
			if exec_Result.Return_Value!=nil {
				fmt.Println(exec_Result.Return_Value)
			}
			wg.Done()
			thread.Exited=true
			thread.Channel <- 0
		}(thread)
	}
	for {
		to_exit:=true
		make_sequential:=true
		for i := 0; i < n; i++ {
			if !threads[i].Exited {
				to_exit=false
			}
			if !threads[i].Inner_Waiting && !threads[i].Exited {
				make_sequential=false
			}
		}
		if to_exit {
			break
		}
		if make_sequential {
			for i := 0; i < n; i++ {
				threads[i].Channel <- 0
				<-threads[i].Channel
			}
		}
	}
	wg.Wait()
	val,_,_:=DB_Read(1, "hi")
	fmt.Println(Vengine.Decode_Object(val))
}