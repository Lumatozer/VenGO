package main

const (
	SET_INSTRUCTION                   int = iota
	RETURN_INSTRUCTION                int = iota
	CALL_INSTRUCTION                  int = iota
	ADD_INSTRUCTION                   int = iota
	SUB_INSTRUCTION                   int = iota
	MULT_INSTRUCTION                  int = iota
	DIV_INSTRUCTION                   int = iota
	FLOOR_INSTRUCTION                 int = iota
	POWER_INSTRUCTION                 int = iota
	GREATER_INSTRUCTION               int = iota
	SMALLER_INSTRUCTION               int = iota
	EQUALS_INSTRUCTION                int = iota
	NEQUALS_INSTRUCTION               int = iota
	AND_INSTRUCTION                   int = iota
	OR_INSTRUCTION                    int = iota
	XOR_INSTRUCTION                   int = iota
	MOD_INSTRUCTION                   int = iota
	DEEP_COPY_OBJECT_INSTRUCTION      int = iota
	JUMP_INSTRUCTION                  int = iota
	NOT_INSTRUCTION                   int = iota
)