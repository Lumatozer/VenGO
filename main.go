package main

import (
	"github.com/lumatozer/VenGO/database"
	"os"
	"fmt"
)

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
	dat, err := os.ReadFile("alu.vi")

	if (err!=nil) {
		fmt.Println(err)
		return
	}
	code:=string(dat)
	if (true) {
		fmt.Println(Vengine(code,true))
	} else {
		// // fmt.Println(tokensier(code,true))
		tokens_, _:=tokens_parser(tokensier(code,true),true)
		// // fmt.Println(tokens_)
		tokens_, _=token_grouper(tokens_, true)
		// fmt.Println(tokens_, len(tokens_))
		fmt.Println("Processing:")
		build_output:=build(tokens_, 0)
		fmt.Println("OUTPUT:")
		fmt.Println(build_output)
	}
}