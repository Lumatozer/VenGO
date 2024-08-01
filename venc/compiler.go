package venc

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

func Hash(a string) string {
	hashed_Bytes:=make([]byte, 0)
	for _,b:=range sha256.Sum256([]byte(a)) {
		hashed_Bytes = append(hashed_Bytes, b)
	}
	return hex.EncodeToString(hashed_Bytes)
}

func Same_Types(Type_A *Type, Type_B *Type) bool {
	return Type_Signature(Type_A, make([]*Type, 0))==Type_Signature(Type_B, make([]*Type, 0))
}

func Type_Signature(a *Type, traversed []*Type) string {
	if a.Is_Array {
		return Hash("array"+Type_Signature(a.Child, traversed))
	}
	if a.Is_Dict {
		return Hash("dict:key->"+Reverse_Standard_Type_Map[a.Raw_Type]+":value->"+Type_Signature(a.Child, traversed))
	}
	if a.Is_Raw {
		return Hash("raw"+Reverse_Standard_Type_Map[a.Raw_Type])
	}
	if a.Is_Pointer {
		return Hash("pointer"+Type_Signature(a.Child, traversed))
	}
	if a.Is_Struct {
		for _,traveresed_Struct:=range traversed {
			if traveresed_Struct==a {
				return Hash("struct-recursed-at"+strconv.FormatInt(int64(len(traversed)), 10))
			}
		}
		traversed = append(traversed, a)
		struct_Keys:=make([]string, 0)
		for Field:=range a.Struct_Details {
			struct_Keys = append(struct_Keys, Field)
		}
		slices.Sort(struct_Keys)
		out:="struct{"
		for _,Field:=range struct_Keys {
			out+=Field+"->"+Type_Signature(a.Struct_Details[Field], traversed)
		}
		out+="}"
		return Hash(out)
	}
	return ""
}

func Type_Object_To_String(Type_Object *Type, program *Program) string {
	if Type_Object.Is_Array {
		return "["+Type_Object_To_String(Type_Object.Child, program)+"]"
	}
	if Type_Object.Is_Dict {
		Key_Type:=Reverse_Standard_Type_Map[Type_Object.Raw_Type]
		return "{"+Key_Type+"->"+Type_Object_To_String(Type_Object.Child, program)+"}"
	}
	if Type_Object.Is_Pointer {
		return "*"+Type_Object_To_String(Type_Object.Child, program)
	}
	if Type_Object.Is_Raw {
		return Reverse_Standard_Type_Map[Type_Object.Raw_Type]
	}
	if Type_Object.Is_Struct {
		struct_Name:=""
		for Struct:=range program.Structs {
			if Same_Types(program.Structs[Struct], Type_Object) {
				struct_Name=Struct
				break
			}
		}
		return struct_Name
	}
	return ""
}

func Get_Program_Import_Tree(program *Program) map[string]*Program {
	out:=make(map[string]*Program)
	for _,Imported_Program:=range program.Imported_Libraries {
		out[Imported_Program.Path]=Imported_Program
		for Path, Internal_Program:=range Get_Program_Import_Tree(Imported_Program) {
			out[Path]=Internal_Program
		}
	}
	return out
}

func Compile_Program(program *Program) {
	current_Dir,_:=os.Getwd()
	compiled_Program:=Compile(program)
	Program_Path:=filepath.Join("distributable", strings.TrimPrefix(program.Path, current_Dir))
	if strings.HasSuffix(Program_Path, ".vi") {
		Program_Path=Program_Path[:len(Program_Path)-2]+"vasm"
	}
	os.MkdirAll(Program_Path, os.ModePerm)
	os.Remove(Program_Path)
	os.Create(Program_Path)
	os.WriteFile(Program_Path, []byte(compiled_Program), 0644)
	old_Dir,_:=filepath.Abs(current_Dir)
	for Program_Path, Imported_Program:=range Get_Program_Import_Tree(program) {
		Program_Path=filepath.Join("distributable", strings.TrimPrefix(Program_Path, current_Dir))
		if strings.HasSuffix(Program_Path, ".vi") {
			Program_Path=Program_Path[:len(Program_Path)-2]+"vasm"
		}
		os.MkdirAll(Program_Path, os.ModePerm)
		os.Remove(Program_Path)
		os.Create(Program_Path)
		os.Chdir(filepath.Dir(Program_Path))
		compiled_Program=Compile(Imported_Program)
		os.Chdir(old_Dir)
		os.WriteFile(Program_Path, []byte(compiled_Program), 0644)
	}
}

func Compile(program *Program) string {
	compiled := "package "+program.Package_Name+"\n\n"
	if len(program.Imported_Libraries) != 0 {
		compiled += "import (\n"
	}
	current_Dir,_:=os.Getwd()
	Absolute_Current_File_Path,_:=filepath.Abs(current_Dir)
	for Import_Alias, Imported_Program := range program.Imported_Libraries {
		Imported_Program.Path=strings.TrimPrefix(Imported_Program.Path, Absolute_Current_File_Path)
		compiled += "    " + "\"" + Imported_Program.Path +"\" as "+Import_Alias+"\n"
	}
	if len(program.Imported_Libraries)!=0 {
		compiled+=")\n"
	}
	for Struct:=range program.Structs {
		if strings.Contains(Struct, ".") {
			continue
		}
		compiled+="struct "+Struct+" {\n"
		for Field, Field_Type:=range program.Structs[Struct].Struct_Details {
			compiled+="    "+Field+"->"+Type_Object_To_String(Field_Type, program)+"\n"
		}
		compiled+="}\n\n"
	}
	for Variable, Variable_Type:=range program.Global_Variables {
		compiled+="var "+Variable+"->"+Type_Object_To_String(Variable_Type, program)+";\n\n"
	}
	sorted_Functions:=make([]string, 0)
	for Function:=range program.Functions {
		sorted_Functions = append(sorted_Functions, Function)
	}
	slices.Sort(sorted_Functions)
	for _,Function:=range sorted_Functions {
		if strings.Contains(Function, ".") {
			continue
		}
		compiled+="function "+Function+"("
		Argument_String:=""
		for Argument:=range program.Functions[Function].Arguments {
			Argument_String+=program.Functions[Function].Arguments[Argument].Name+"->"+Type_Object_To_String(program.Functions[Function].Arguments[Argument].Type, program)+", "
		}
		Argument_String=strings.Trim(Argument_String, ", ")
		compiled+=Argument_String
		compiled+=")"
		compiled+="->"+Type_Object_To_String(program.Functions[Function].Out_Type, program)+" {\n"
		for _,Instruction_Line:=range program.Functions[Function].Instructions {
			compiled+="    "+strings.Join(Instruction_Line, " ")+"\n"
		}
		compiled+="}\n\n"
	}
	return strings.Trim(compiled, "\n")
}