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

func Convert_Raw_Types_For_Vitality(a int8) int8 {
	if a==INT_TYPE {
		return venc.INT_TYPE
	}
	if a==INT64_TYPE {
		return venc.INT64_TYPE
	}
	if a==STRING_TYPE {
		return venc.STRING_TYPE
	}
	if a==FLOAT_TYPE {
		return venc.FLOAT_TYPE
	}
	if a==FLOAT64_TYPE {
		return venc.FLOAT64_TYPE
	}
	if a==POINTER_TYPE {
		return venc.POINTER_TYPE
	}
	if a==VOID_TYPE {
		return venc.VOID_TYPE
	}
	return 0
}

func VASM_Type_To_Vitality_Type(a *Type) *venc.Type {
	if a==nil {
		return &venc.Type{}
	}
	out:=&venc.Type{Is_Array: a.Is_Array, Is_Dict: a.Is_Dict, Is_Raw: a.Raw_Type!=0, Raw_Type: Convert_Raw_Types_For_Vitality(a.Raw_Type), Is_Struct: a.Is_Struct, Is_Pointer: a.Raw_Type==POINTER_TYPE, Child: VASM_Type_To_Vitality_Type(a.Child), Struct_Details: make(map[string]*venc.Type)}
	for Field,Field_Type:=range a.Struct_Details {
		out.Struct_Details[Field]=VASM_Type_To_Vitality_Type(Field_Type)
	}
	return out
}

func VASM_Program_To_Vitality_Program(program Program, path string) venc.Program {
	venc_Program:=venc.Program{Vitality: false, Path: path, Package_Name: program.Package_Name, Structs: make(map[string]*venc.Type), Functions: make(map[string]*venc.Function), Global_Variables: make(map[string]*venc.Type), Imported_Libraries: make(map[string]*venc.Program)}
	for Struct:=range program.Structs {
		venc_Program.Structs[Struct]=VASM_Type_To_Vitality_Type(program.Structs[Struct])
	}
	for _,Function:=range program.Functions {
		venc_Program.Functions[Function.Name]=&venc.Function{Out_Type: VASM_Type_To_Vitality_Type(&Function.Out_Type), Arguments: make([]struct{Name string; Type *venc.Type}, 0), Scope: make(map[string]*venc.Type), Instructions: make([][]string, 0)}
		for Function_Argument:=range Function.Arguments {
			Argument_Type:=Function.Arguments[Function_Argument]
			venc_Program.Functions[Function.Name].Arguments = append(venc_Program.Functions[Function.Name].Arguments, struct{Name string; Type *venc.Type}{Name: Function_Argument, Type: VASM_Type_To_Vitality_Type(&Argument_Type)})
		}
	}
	return venc_Program
}

func VASM_Translator(path string) (venc.Program, error) {
	data,err:=os.ReadFile(path)
	if err!=nil {
		fmt.Println(err)
		return venc.Program{}, err
	}
	tokens,err:=Tokenizer(string(data))
	if err!=nil {
		fmt.Println(err)
		return venc.Program{}, err
	}
	VASM_Program, err:=Parser(tokens, path, make(map[string]Program))
	if err!=nil {
		return venc.Program{}, err
	}
	return VASM_Program_To_Vitality_Program(VASM_Program, path), nil
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
	Absolute_Path,err:=filepath.Abs(os.Args[1])
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
		program,err:=venc.Parser(os.Args[1], definitions, make(map[string]venc.Program), VASM_Translator)
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
	tokens,err:=Tokenizer(string(data))
	if err!=nil {
		fmt.Println(err)
		return
	}
	fmt.Println(tokens)
	if err!=nil {
		fmt.Println(err)
		return
	}
	program,err:=Parser(tokens, Absolute_Path, make(map[string]Program))
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
		fmt.Println(exec_Result.Error)
		return
	}
	if exec_Result.Return_Value!=nil {
		fmt.Println(exec_Result.Return_Value)
	}
}