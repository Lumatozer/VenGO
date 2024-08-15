package time

import (
	"sync"
	"time"

	"github.com/lumatozer/VenGO/structs"
)

var times []time.Time=make([]time.Time, 0)

var mutex *sync.Mutex=&sync.Mutex{}

func Time(objects []*interface{}) structs.Execution_Result {
	mutex.Lock()
	times = append(times, time.Now())
	mutex.Unlock()
	return structs.Execution_Result{Return_Value: len(times)-1}
}

func Since(objects []*interface{}) structs.Execution_Result {
	return structs.Execution_Result{Return_Value: int(time.Since(times[(*objects[0]).(int)]).Seconds())}
}

func Get_Package() structs.Package {
	return structs.Package{Name: "time", Functions: map[string]func([]*interface{})structs.Execution_Result{
		"Time":Time,
		"Since":Since,
	}}
}