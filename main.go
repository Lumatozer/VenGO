package main

import (
	"fmt"
	"os"
	"github.com/lumatozer/VenGO/database"
)

func diff(arr1 []Token, arr2 []Token) {

}

func main() {
	if (database.Init())==0 {
		return
	}
	// database.DB_set("alu","aa")
	// database.DB_set("alux","bb")
	// database.DB_set("alu","a")
	// database.DB_set("alu","aa")
	// database.DB_delete("alux")
	// fmt.Println(database.DB_get("alu"))
	if len(os.Args)<2 {
		fmt.Println("No file mentioned")
		return
	}
	dat, err := os.ReadFile(os.Args[1])

	if (err!=nil) {
		fmt.Println(err)
		return
	}
	code:=string(dat)
	if (false) {
		fmt.Println(Vengine(code,true))
	} else {
		// // fmt.Println(tokensier(code,true))
		tokens_, _:=tokens_parser(tokensier(code,true),false)
		// fmt.Println(tokens_,"hi")
		// fmt.Println(tokens_parser(tokensier(code,true),true))
		// fmt.Println("----------")
		// // fmt.Println(tokens_)
		tokens_, _=token_grouper(tokens_, true)
		// fmt.Println(tokens_, len(tokens_))
		fmt.Println("Processing:")
		symbol_table:=Symbol_Table{operations: make(map[string][][]string), used_variables: make(map[string][]int), variable_mapping: make(map[string]string), files: make(map[string]Symbol_Table), current_file: "alu.vi", imported_libraries: make(map[string]string), global_variables: make([][]string, 0), struct_mapping: make(map[string]string)}
		symbol_table.files[symbol_table.current_file]=symbol_table
		build_output,_:=build(symbol_table, tokens_, 0)
		fmt.Println("OUTPUT:")
		fmt.Println(build_output)
		fmt.Println("VENGINE RUN:")
		Vengine(build_output, true)
	}
}