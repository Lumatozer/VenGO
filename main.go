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
	Vengine(code)
}