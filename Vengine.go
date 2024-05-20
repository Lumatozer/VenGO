package main

const (
	Type_Bool            int=iota
	Type_Int 	 	 	 int=iota
	Type_Int64   	 	 int=iota
	Type_Float 	 	 	 int=iota
	Type_Float64 	 	 int=iota
	Type_String  	 	 int=iota
	Type_Bytes   	 	 int=iota
	Type_Array   	 	 int=iota
	Type_Map     	 	 int=iota
	Type_Function        int=iota
)

const (
	TOK_SYS      	 	 int=iota
	TOK_VAR      	 	 int=iota
	TOK_EQUALS   	 	 int=iota
	TOK_NUMBER   	 	 int=iota
	TOK_DOT      	 	 int=iota
	TOK_CURLY_OPEN   	 int=iota
	TOK_CURLY_CLOSE  	 int=iota
	TOK_SQUARE_OPEN  	 int=iota
	TOK_SQUARE_CLOSE 	 int=iota
	TOK_ROUND_OPEN   	 int=iota
	TOK_ROUND_CLOSE  	 int=iota
	TOK_MINUS        	 int=iota
	TOK_PLUS         	 int=iota
	TOK_DIV          	 int=iota
	TOK_MULTIPLY     	 int=iota
	TOK_EXPONENT     	 int=iota
	TOK_HASH         	 int=iota
	TOK_AT           	 int=iota
	TOK_EXCLAMATION  	 int=iota
	TOK_AND          	 int=iota
	TOK_GREATER      	 int=iota
	TOK_LESSER       	 int=iota
	TOK_ARROW        	 int=iota
	TOK_SEMI_COLON   	 int=iota
	TOK_COLON        	 int=iota
	TOK_DOUBLE_QUOTE 	 int=iota
	TOK_SINGLE_QUOTE 	 int=iota
	TOK_COMMA        	 int=iota
	TOK_MOD          	 int=iota
	TOK_CALL             int=iota
	TOK_FUNCTION         int=iota
	TOK_STRUCT           int=iota
	TOK_TYPE             int=iota
	TOK_STRING           int=iota
)

type Flow struct {
	Token_Type 			 int
	Value_Type 			 []int
}

func Parse(code string) string {
	return ""
}