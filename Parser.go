package main

type Object struct {
	Type 			    []string
	Location        	int
	Callable_Functions  []int
	Int_Mapping         map[int]*Object
	Int64_Mapping       map[int64]*Object
	Float64_Mapping     map[float32]*Object
	Float_Mapping       map[float64]*Object
	Field_Children      map[int]*Object
	Children            []*Object
}

type Function struct {
	Id                  int
	Name 				string
	Stack_Spec          []int                             // initialize these objects with default properties with new pointers for this scope
	Instructions        [][]int                           // Instruction set for this function
	Arguments           map[string][]string
	Out_Type            []string
	Base_Scope          *Scope                            // for initializing a new Function scope under the program this function belongs to. Will always be the Program's Rendered Scope as only the arguments change
}

type Scope struct {
	Ip                  int
	Objects             []*Object
}

type Program struct {
	Functions			[]*Function
	Structs             map[string]map[string][]string
	Rendered_Scope      Scope                            // This Scope will be used for initalizing functions of this file + will retain all the final global states of the variables
	State_Variables     []int                            // Indices of variables to be stored on the blockchain for this program
}

type Execution struct {
	Gas_Limit            int64
	Entry_Program        *Program
	Entry_Function       *Function
	Programs             []Program
}