package venc

import (
	"fmt"
	"go/types"
	"os"
	"strings"
)

const (
	_ int = iota
	String_Kind
	Struct_Kind
	Slice_Kind
	Bool_Kind
	Optional_Kind
	Number_Kind // float, int, uint
)

type Venc_Type struct {
	Kind int
	Fields_Keys []string // struct
	Fields_Types []*Venc_Type // struct
	String_Value *string // string
	Child *Venc_Type // slice, optional
}

func Get_Type(t types.Type, type_info map[string]*Venc_Type) *Venc_Type {
	potential_alias := t.String()
	potential_type, ok := type_info[potential_alias]

	if ok {
		return potential_type
	}

	var out_type Venc_Type = Venc_Type{}

	if strings.Contains(potential_alias, ".") {
		type_info[potential_alias] = &out_type
	}

	switch underlying := t.Underlying().(type) {

	case *types.Basic: // int, string, float, bool
		info := underlying.Info()

		if info&(types.IsOrdered|types.IsBoolean|types.IsUnsigned) == 0 {
			fmt.Println("Unsupported type", underlying)
			os.Exit(1)
		}

		final_type := underlying.String()

		out_type.Kind = Number_Kind
		out_type.String_Value = &final_type
	
	case *types.Struct:
		out_type.Kind = Struct_Kind

		for field := range underlying.Fields() {
			field_name := field.Name()
			field_type := field.Type()

			out_type.Fields_Keys = append(out_type.Fields_Keys, field_name)
			out_type.Fields_Types = append(out_type.Fields_Types, Get_Type(field_type, type_info))
		}
	
	case *types.Slice:
		inner_type := underlying.Elem()
		out_type.Kind = Slice_Kind
		out_type.Child = Get_Type(inner_type, type_info)
	
	case *types.Pointer:
		elem := underlying.Elem()
		if _, ok := elem.(*types.Pointer); ok {
			fmt.Println("Unsupported type: double pointer", t)
			os.Exit(1)
		}

		out_type.Kind = Optional_Kind
		out_type.Child = Get_Type(elem, type_info)

	default:
		fmt.Println("Unsupported type", underlying)
		os.Exit(1)
	}

	return &out_type
}