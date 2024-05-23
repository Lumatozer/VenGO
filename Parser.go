package main

type Object struct {
	Type 			    *Type
	Location        	int
	Callable_Functions  []int
	Int_Mapping         map[int]int
	Int64_Mapping       map[int64]int
	Float64_Mapping     map[float32]int // uses the same space as a pointer when uninitialised (8 bytes on 64 bit)
	Float_Mapping       map[float64]int
	Field_Children      map[int]int
	Children            []int // index in scope objects to allow scope based object localisation
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

func Create_Flow() Flow {
	return Flow{
		Structs: map[string]Type{
			"int":Type{
				Is_Primitive: true,
				Primitive_Type: "int",
			},
			"int64":Type{
				Is_Primitive: true,
				Primitive_Type: "int64",
			},
			"float":Type{
				Is_Primitive: true,
				Primitive_Type: "float",
			},
			"float64":Type{
				Is_Primitive: true,
				Primitive_Type: "float64",
			},
			"string":Type{
				Is_Primitive: true,
				Primitive_Type: "string",
			},
		},
	}
}

func Parse_Flow(tokens []Token, flow *Flow) {
	for i:=0; i<len(tokens); i++ {
		if tokens[i].Type=="struct" {

		}
	}
}