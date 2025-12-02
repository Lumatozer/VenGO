module github.com/lumatozer/Vengine

go 1.25

replace github.com/lumatozer/VenGO => ./vengine

require github.com/lumatozer/VenGO v0.0.0

replace github.com/lumatozer/VenC => ./venc

require github.com/lumatozer/VenC v0.0.0

require github.com/bytecodealliance/wasmtime-go/v39 v39.0.1 // indirect
