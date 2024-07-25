package venc

import (
	"crypto/sha256"
	"encoding/hex"
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

func Compile(program Program) string {
	compiled := "package "+program.Package_Name+"\n\n"
	if len(program.Imported_Libraries) != 0 {
		compiled += "import (\n"
	}
	for Import_Alias, Imported_Program := range program.Imported_Libraries {
		compiled += "    " + "\"" + strings.Split(Imported_Program.Path, ".")[0]+"\" as "+Import_Alias+"\n"
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
			compiled+="    "+Field+"->"+Type_Object_To_String(Field_Type, &program)+"\n"
		}
		compiled+="}\n"
	}
	return strings.Trim(compiled, "\n")
}