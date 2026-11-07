//go:build kernel

package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	assembler "zepa-machine/cross-assembler"
	"zepa-machine/machine"
)

const (
	kernelMappingOffset = machine.KernelVirtualOffset // kernel mapeado 3GB acima
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

	memorySize, err := parseMemorySize(os.Args[1])
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	if memorySize%4 != 0 {
		log.Fatalf("memory_size must be a multiple of 4, got %dB", memorySize)
	}
	if memorySize < (1<<30)+(1<<12) {
		log.Fatalf("memory_size must be at least 1GB + 4KB, got %dB", memorySize)
	}
	if memorySize > (1 << 32) {
		log.Fatalf("memory_size must be at most 4GB, got %dB", memorySize)
	}

	timeSlice, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	machine := machine.NewMachine(memorySize, true)
	machine.LoadProgram(binaryCode)

	if err := initializeKernelGlobals(machine.GetMemory(), uint32(timeSlice)); err != nil {
		log.Fatalf("Initialize kernel globals: %v", err)
	}

	go machine.Boot()
	runTUIWithMachine(machine)
}

func initializeKernelGlobals(memory []byte, timeSlice uint32) error {
	offsets := [...]uint32{
		machine.GlobalMaxProcessesOffset,
		machine.GlobalTimeSliceOffset,
		machine.GlobalBufferSizeOffset,
	}
	values := [...]uint32{
		machine.PCBCount,
		timeSlice,
		machine.TransferBufferBytes,
	}

	for _, offset := range offsets {
		address := machine.GlobalAddress(offset)
		if uint64(address)+4 > uint64(len(memory)) {
			return fmt.Errorf("global address %#x exceeds memory backing (%d bytes)", address, len(memory))
		}
	}

	for i, offset := range offsets {
		address := int(machine.GlobalAddress(offset))
		binary.LittleEndian.PutUint32(memory[address:address+4], values[i])
	}
	return nil
}

func parseMemorySize(sizeStr string) (int, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	multiplier := 1

	switch {
	case strings.HasSuffix(sizeStr, "GB"):
		multiplier = 1 << 30
		sizeStr = strings.TrimSuffix(sizeStr, "GB")
	case strings.HasSuffix(sizeStr, "MB"):
		multiplier = 1 << 20
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
	case strings.HasSuffix(sizeStr, "KB"):
		multiplier = 1 << 10
		sizeStr = strings.TrimSuffix(sizeStr, "KB")
	case strings.HasSuffix(sizeStr, "B"):
		sizeStr = strings.TrimSuffix(sizeStr, "B")
	}

	val, err := strconv.Atoi(sizeStr)
	if err != nil {
		return 0, fmt.Errorf("invalid memory format: %s", sizeStr)
	}

	return val * multiplier, nil
}
