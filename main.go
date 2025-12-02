package main

import (
	"fmt"
	"os"
	"time"

	"github.com/lumatozer/VenGO"
)

func main() {
	data, _ := os.ReadFile("test.wasm")

	start := time.Now().Local().UnixMilli()
	_, err := vengine.CompileWasm(data)
	fmt.Println(time.Now().Local().UnixMilli() - start, err)
}