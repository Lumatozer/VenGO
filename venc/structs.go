package venc

const (
	INT_TYPE               int8 = iota
	INT64_TYPE             int8 = iota
	STRING_TYPE            int8 = iota
	FLOAT_TYPE             int8 = iota
	FLOAT64_TYPE           int8 = iota
	POINTER_TYPE           int8 = iota
	VOID_TYPE              int8 = iota
)

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
	Name                 string
	Out_Type             Type
	Arguments            map[string]Type
	Scope                map[string]Type
	Instructions         [][]string
}

type Program struct {
	Path                 string
	Structs              map[string]*Type
	Functions            []Function
	Global_Variables     map[string]Type
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