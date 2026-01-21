package venc

import (
	"go/types"
)

func isNumeric(t types.Type) bool {
	switch u := t.Underlying().(type) {
	case *types.Basic:
		info := u.Info()
		return info&(types.IsInteger|types.IsFloat) != 0
	default:
		return false
	}
}

type Function_Type struct {
	Name string
	Parameter_Keys []string
	Parameter_Types []*Venc_Type
	Results []*Venc_Type
}