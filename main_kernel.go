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
	buffer        = 64 * 1024       // 64KB
	minKernelSize = 8 * 1024 * 1024 // 8MB
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run -tags kernel . <memory_size> <partition_size> <time_slice>")
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
	if memorySize > (1 << 32) {
		log.Fatalf("memory_size must be at most 4GB, got %dB", memorySize)
	}

	partitionSize, err := parseMemorySize(os.Args[2])
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	if partitionSize%4 != 0 {
		log.Fatalf("partition_size must be a multiple of 4, got %dB", partitionSize)
	}

	if memorySize < minKernelSize+buffer+partitionSize {
		log.Fatalf("memory_size must be at least 8MB + 6KB + partition_size, got %dB", memorySize)
	}

	timeSlice, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	machine := machine.NewMachine(memorySize, true)
	machine.LoadProgram(binaryCode)

	memSlice := machine.GetMemory()[4096:4100]
	binary.LittleEndian.PutUint32(memSlice, uint32(partitionSize))

	memSlice = machine.GetMemory()[4100:4104]
	binary.LittleEndian.PutUint32(memSlice, uint32(timeSlice))

	memSlice = machine.GetMemory()[4108:4112]
	binary.LittleEndian.PutUint32(memSlice, uint32(0x10000))

	go machine.Boot()
	runTUIWithMachine(machine)
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
