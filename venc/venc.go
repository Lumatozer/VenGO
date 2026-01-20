package venc

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"sort"

	"golang.org/x/tools/go/packages"
)

func CompilerPackage(directory string) {
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

	for _, pkg := range pkgs {
		buildTypeGraph(pkg, type_info)
	}
}

func buildTypeGraph(pkg *packages.Package, type_info map[string]*Venc_Type) {
	scope := pkg.Types.Scope()

	names := scope.Names()
	sort.Strings(names)

	for _, name := range names {
		if !ast.IsExported(name) {
			continue
		}

		fn, ok := scope.Lookup(name).(*types.Func)
		if !ok {
			continue
		}

		sig := fn.Type().(*types.Signature)

		results := sig.Results()

		// add logic for function construction

		for i := 0; i < results.Len(); i++ {
			result := results.At(i)
			Get_Type(result.Type(), type_info)
		}

		params := sig.Params()

		for i := 0; i < params.Len(); i++ {
			param := params.At(i)
			Get_Type(param.Type(), type_info)
		}

		fmt.Println(type_info)

		// the_type := sig.Results().At(0).Type()

		// fmt.Println(the_type, "->", types.Unalias(the_type.Underlying()))

		// fmt.Printf(
		// 	"%s.%s %s\n",
		// 	pkg.PkgPath,
		// 	fn.Name(),
		// 	types.TypeString(sig, nil),
		// )
	}
}
