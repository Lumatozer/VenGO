package print

import (
	"github.com/lumatozer/VenGO/structs"
)

func Print(objects []*interface{}) structs.Execution_Result {
	*objects[0]=20
	return structs.Execution_Result{}
}