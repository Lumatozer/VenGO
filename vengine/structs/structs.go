package structs

import (
	"sync"
)

type Execution_Result struct {
	Gas_Used        int
	Return_Value    interface{}
	Error           error
}

type Package struct {
	Name      string
	Functions         map[string]func([]*interface{})Execution_Result
}

type Database_Interface struct {
	Locking_Databases []int
	DB_Read           func(database int, key string) (Database_Value []byte, Gas_Used int, Error error)
	DB_Write          func(database int, key string, value []byte) (Gas_Used int, Error error)
}

type Mutex_Interface struct {
	Locked            bool
	Mutex             *sync.Mutex
}

func Lock(m *Mutex_Interface) {
	m.Mutex.Lock()
	m.Locked=true
}

func Unlock(m *Mutex_Interface) {
	m.Mutex.Unlock()
	m.Locked=false
}