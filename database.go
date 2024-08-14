package main

import (
	"errors"
)

var Memory_Database map[int]map[string][]byte = make(map[int]map[string][]byte)

func DB_Write(database int, key string, value []byte) (Gas_Used int, Error error) {
	Gas_Used = 0
	Error = nil
	_,ok:=Memory_Database[database]
	if !ok {
		Memory_Database[database]=make(map[string][]byte)
	}
	Memory_Database[database][key] = value
	return Gas_Used, Error
}

func DB_Read(database int, key string) (Database_Value []byte, Gas_Used int, Error error) {
	value, ok := Memory_Database[database][key]
	if !ok {
		Error = errors.New("key not found in database")
	}
	return value, 0, Error
}