package main

import (
	"github.com/lumatozer/VenGO/packages/print"
	"github.com/lumatozer/VenGO/structs"
)

type Package struct {
	Name      string
	Function_Names    []string
	Functions         []func([]*interface{})structs.Execution_Result
}

func Get_Packages() []Package {
	packages:=make([]Package, 0)
	packages = append(packages, Package{Name: "print", Function_Names: []string{"print"}, Functions: []func([]*interface{})structs.Execution_Result{print.Print}})
	return packages
}