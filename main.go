package main

import (
	"fmt"
	"os"
	"venc"
	// "time"
	// "github.com/lumatozer/VenGO"
)

// func main() {
// 	data, _ := os.ReadFile("test.wasm")

// 	start := time.Now().Local().UnixMilli()
// 	_, err := vengine.CompileWasm(data)
// 	fmt.Println(time.Now().Local().UnixMilli() - start, err)
// }

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Must include directory path as CLI arguments")
		return
	}

	venc.CompilePackage(os.Args[1])
}