package venc

import (
	"go/ast"
	"go/types"
	"os"
	"sort"

	"golang.org/x/tools/go/packages"
)

func CompilePackage(directory string) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedModule,
		Dir: directory,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		panic(err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	var type_info map[string]*Venc_Type = make(map[string]*Venc_Type)

	function_map := []*Function_Type{}

	for _, pkg := range pkgs {
		function_map = append(function_map, buildFunctionTypeGraph(pkg, type_info)...)
	}
}

func buildFunctionTypeGraph(pkg *packages.Package, type_info map[string]*Venc_Type) []*Function_Type {
	scope := pkg.Types.Scope()

	names := scope.Names()
	sort.Strings(names)

	functions := []*Function_Type{}

	for _, name := range names {
		if !ast.IsExported(name) {
			continue
		}

		fn, ok := scope.Lookup(name).(*types.Func)
		if !ok {
			continue
		}

		sig := fn.Type().(*types.Signature)

		function := Function_Type{}
		function.Name = pkg.PkgPath + "." + fn.Name()

		params := sig.Params()

		for i := 0; i < params.Len(); i++ {
			param := params.At(i)
			function.Parameter_Keys = append(function.Parameter_Keys, param.Name())
			function.Parameter_Types = append(function.Parameter_Types, Get_Type(param.Type(), type_info))
		}

		results := sig.Results()
		for i := 0; i < results.Len(); i++ {
			result := results.At(i)
			function.Results = append(function.Results, Get_Type(result.Type(), type_info))
		}

		functions = append(functions, &function)
	}

	return functions
}
