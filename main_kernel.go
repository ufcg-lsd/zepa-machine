//go:build kernel

package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	assembler "zepa-machine/cross-assembler"
	"zepa-machine/machine"
)

const (
	kernelMappingOffset = 0xC0000000 // kernel mapeado 3GB acima
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run -tags kernel . <memory_size> <time_slice>")
		return
	}

	sourceFile := "sackOS/kernel.asm"
	binaryCode, err := assembler.RunAssembler(sourceFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	memorySize, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	if memorySize%4 != 0 {
		log.Fatalf("memory_size must be a multiple of 4, got %d", memorySize)
	}
	if memorySize < (1<<30)+(1<<12) {
		log.Fatalf("memory_size must be at least 1GB + 4KB, got %d", memorySize)
	}

	timeSlice, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	machine := machine.NewMachine(memorySize, true)
	machine.LoadProgram(binaryCode)

	memSlice := machine.GetMemory()[0x2000:0x2004]
	binary.LittleEndian.PutUint32(memSlice, uint32(256)) // max processes

	memSlice = machine.GetMemory()[0x2004:0x2008]
	binary.LittleEndian.PutUint32(memSlice, uint32(timeSlice))

	memSlice = machine.GetMemory()[0x2008:0x200C]
	binary.LittleEndian.PutUint32(memSlice, uint32(0x10000)) // buffer size

	go machine.Boot()
	runTUIWithMachine(machine)
}
