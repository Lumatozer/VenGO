package venc

import (
	"strings"
)

func Same_Types(Type_A *Type, Type_B *Type) bool {
	if (Type_A==nil && Type_B!=nil) || (Type_B==nil && Type_A!=nil) {
		return false
	}
	if Type_A==nil && Type_B==nil {
		return true
	}
	if Type_A.Is_Array!=Type_B.Is_Array {
		return false
	}
	if Type_A.Is_Dict!=Type_B.Is_Dict {
		return false
	}
	if Type_A.Is_Pointer!=Type_B.Is_Pointer {
		return false
	}
	if Type_A.Is_Raw!=Type_B.Is_Raw {
		return false
	}
	if Type_A.Is_Struct!=Type_B.Is_Struct {
		return false
	}
	if !Same_Types(Type_A.Child, Type_B.Child) {
		return false
	}
	if (Type_A.Struct_Details==nil && Type_B.Struct_Details!=nil) || (Type_A.Struct_Details!=nil && Type_B.Struct_Details==nil) {
		return false
	}
	if Type_A.Struct_Details!=nil {
		if len(Type_A.Struct_Details)!=len(Type_B.Struct_Details) {
			return false
		}
		for Key:=range Type_A.Struct_Details {
			_,ok:=Type_B.Struct_Details[Key]
			if !ok {
				return false
			}
			if !Same_Types(Type_A.Struct_Details[Key], Type_B.Struct_Details[Key]) {
				return false
			}
		}
	}
	return true
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
		compiled+="struct "+Struct+" {\n"
		for Field, Field_Type:=range program.Structs[Struct].Struct_Details {
			compiled+="    "+Field+"->"+Type_Object_To_String(Field_Type, &program)+"\n"
		}
		compiled+="}\n"
	}
	return strings.Trim(compiled, "\n")
}