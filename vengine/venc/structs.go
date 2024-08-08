package Venc

const (
	INT_TYPE               int8 = iota
	INT64_TYPE             int8 = iota
	STRING_TYPE            int8 = iota
	FLOAT_TYPE             int8 = iota
	FLOAT64_TYPE           int8 = iota
	POINTER_TYPE           int8 = iota
	VOID_TYPE              int8 = iota
)

var TYPE_MAP map[string]int8 = map[string]int8{
	"int":INT_TYPE,
	"int64":INT64_TYPE,
	"string":STRING_TYPE,
	"float":FLOAT_TYPE,
	"float64":FLOAT64_TYPE,
	"pointer":POINTER_TYPE,
	"void":VOID_TYPE,
}

var Reverse_Standard_Type_Map map[int8]string = map[int8]string{
	INT_TYPE:"int",
	INT64_TYPE:"int64",
	STRING_TYPE:"string",
	FLOAT_TYPE:"float",
	FLOAT64_TYPE:"float64",
	VOID_TYPE:"void",
}

type Token struct {
	Type                 string
	Num_Value            float64
	Value                string
	Line_Number          int
	Children             []Token
}

type Type struct {
	Is_Array             bool
	Is_Dict              bool
	Is_Raw               bool
	Raw_Type             int8
	Is_Struct            bool
	Is_Pointer           bool
	Struct_Details       map[string]*Type
	Child                *Type
}

type Function struct {
	Out_Type             *Type
	Arguments            []struct{Name string; Type *Type}
	Scope                map[string]*Type
	Instructions         [][]string
}

type Program struct {
	Vitality             bool
	Path                 string
	Package_Name         string
	Structs              map[string]*Type
	Functions            map[string]*Function
	Global_Variables     map[string]*Type
	Imported_Libraries   map[string]*Program
}

type Function_Definition struct {
	Name                 string
	Arguments            map[string]Token
	Out_Type             Token
	Internal_Tokens      []Token
}

type Definitions struct {
	Package_Name         string
	Imports              map[string]string
	Variables            map[string]Token
	Functions            []Function_Definition
	Structs              map[string]map[string]Token
}

type Temp_Variables struct {
	Signature_Lookup    map[string]int
	Variable_Lookup     map[int][]struct{Free bool; Allocated bool}
}

type Loop_Details struct {
	Continue_Variable   string
	Break_Variable      string
}