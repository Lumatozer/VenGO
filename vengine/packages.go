package Vengine

import (
	"github.com/lumatozer/VenGO/packages/print"
	"github.com/lumatozer/VenGO/packages/time"
	"github.com/lumatozer/VenGO/structs"
)

func Get_Packages() []structs.Package {
	packages:=make([]structs.Package, 0)
	packages = append(packages, structs.Package{Name: "print", Functions: print.Get_Package().Functions})
	packages = append(packages, structs.Package{Name: "time", Functions: time.Get_Package().Functions})
	return packages
}