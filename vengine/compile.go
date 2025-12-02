package vengine

import (
	"github.com/bytecodealliance/wasmtime-go/v39"
)

func CompileWasm(wasm []byte) ([]byte, error) {
    cfg := wasmtime.NewConfig()
    cfg.SetStrategy(wasmtime.StrategyCranelift)
    cfg.SetCraneliftOptLevel(wasmtime.OptLevelSpeed)
    cfg.SetCraneliftNanCanonicalization(true)
    cfg.SetCraneliftDebugVerifier(false)
    cfg.SetParallelCompilation(true)
    cfg.SetMemoryInitCOWSet(true)
    cfg.SetNativeUnwindInfo(false)
    cfg.SetDebugInfo(false)
    cfg.SetConsumeFuel(true)
    cfg.SetEpochInterruption(false)
    cfg.SetMaxWasmStack(512 * 1024)
    cfg.SetWasmSIMD(true)
    cfg.SetWasmRelaxedSIMD(false)
    cfg.SetWasmBulkMemory(true)
    cfg.SetWasmMultiValue(true)
    cfg.SetWasmReferenceTypes(true)
    cfg.SetWasmThreads(false)
    cfg.SetWasmMultiMemory(false)
    cfg.SetWasmMemory64(false)
    cfg.SetWasmTailCall(false)
    cfg.SetWasmFunctionReferences(false)
    cfg.SetWasmGC(false)
    cfg.SetWasmWideArithmetic(false)
	engine := wasmtime.NewEngineWithConfig(cfg)

	module, err := wasmtime.NewModule(engine, wasm)

	if err != nil {
		return nil, err
	}

	serialized, err := module.Serialize()

	if err != nil {
		return nil, err
	}

    return serialized, nil
}
