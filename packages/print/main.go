package print

import (
	"fmt"
	"github.com/lumatozer/VenGO/structs"
)

func Print(objects []*interface{}) structs.Execution_Result {
	fmt.Println("Console:", *objects[0])
	return structs.Execution_Result{}
}

func Get_Package() structs.Package {
	return structs.Package{Name: "print", Functions: map[string]func([]*interface{})structs.Execution_Result{
		"print":Print,
	}}
}