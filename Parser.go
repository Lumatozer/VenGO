package main

type Object struct {
	Type 			    Type
	Location        	int
	Callable_Functions  []int
	Int_Mapping         map[int]*Object
	Int64_Mapping       map[int64]*Object
	Float64_Mapping     map[float32]*Object // uses the same space as a pointer when uninitialised (8 bytes on 64 bit)
	Float_Mapping       map[float64]*Object
	Field_Children      map[int]*Object
	Children            []*Object
}

type Type struct {
	Name				string
	Is_Array            bool
	Is_Dict             bool
	Is_Primitive        bool
	Primitive_Type      string
	Map_Fields          map[string]Type
	Child				*Type
}

type Function struct {
	Id                  int
	Name 				string
	Callable_Scopes		[]int // *, function_a etc.
	Arguments           map[string]Type
	Out_Type            Type
}

type Scope struct {
	Instruction_Pointer int
	Scope_Id            int
	Objects             []*Object
	Integers            []int
	Integers_64         []int64
	Floats              []float32
	Floats_64           []float64
	Return_Scope        int
	Scopes_On_Top       []int
}

type Flow struct {
	Instructions        [][]int
	Scopes				map[int]Scope
	Structs             map[string]Type
}

type Program struct {
	Flows []Flow
}

func Parse() Program {
	return Program{}
}