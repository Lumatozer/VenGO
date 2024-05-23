package main

// no type validation for instructions

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
	Stack_Spec          []int        // initialize these objects with new pointers for this scope
	Instructions        [][]int      // Instruction set for this function
}

type Scope struct {
	Ip                  int
	Objects             []*Object
	// Return_Scope        *Scope not needed as you will anyways recursively grow the scope, it will anyways return to the last scope
}

type Program struct {
	Functions			[]*Function
	Structs             []*Type
}

type Execution struct {
	
}