package main

import (
	"fmt"
	"os"
	"zepa-machine/core"
	assembler "zepa-machine/cross-assembler"
	"zepa-machine/machine"
)

func DebugMemory(m *machine.Machine) {
	const (
		bytesPerRow   = 4
		diskLoaderEnd = 0x20 // Define o fim do diskLoader (32 em decimal)
	)

	memory := m.GetMemory()

	fmt.Print("\n----------Memory----------")
	for i := 0; i < len(memory); i += bytesPerRow {
		if i < diskLoaderEnd {
			continue
		}

		allZero := true
		for j := 0; j < bytesPerRow; j++ {
			if i+j < len(memory) && memory[i+j] != 0 {
				allZero = false
				break
			}
		}

		if allZero {
			continue
		}

		fmt.Printf("\nInitial Address: 0x%04X -- Instruction: ", i)

		for j := 0; j < bytesPerRow; j++ {
			if i+j < len(memory) {
				fmt.Printf("%08b ", memory[i+j])
			} else {
				fmt.Print("   ")
			}
		}
	}
	fmt.Println("\n")
}

func DebugRegisters(m *machine.Machine) {
	registers := m.GetRegisters()
	fmt.Println("\n----------Registers----------")
	for k, v := range registers {
		// ignore IR
		if k == 8 {
			continue
		}
		fmt.Printf("%v: %d\n", getRegisterName(k), v)
	}
}

func getRegisterName(reg core.Register) string {
	switch reg {
	case 0:
		return "w0"
	case 1:
		return "w1"
	case 2:
		return "w2"
	case 3:
		return "w3"
	case 4:
		return "w4"
	case 5:
		return "w5"
	case 6:
		return "pc"
	case 7:
		return "sp"
	case 8:
		return "ir"
	case 9:
		return "sr"
	case 10:
		return "mdr"
	case 11:
		return "mar"
	case 12:
		return "lr"
	case 13:
		return "evt"
	case 14:
		return "ssr"
	default:
		return "invalid"
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./main.go program1.asm [program2.asm...]")
		return
	}

	programPaths := os.Args[1:]

	// Execute each program
	for i, path := range programPaths {
		fmt.Printf("\n===== Program %d: %s =====\n", i+1, path)

		// Create a new machine for each program
		m := machine.NewMachine(2048)
		m.InitDisk()

		// Assemble and load this program to disk
		code, err := assembler.RunAssembler(path)
		if err != nil {
			fmt.Printf("Error assembling %s: %v\n", path, err)
			continue
		}
		m.AddToDisk(code)

		// Load diskLoader.asm
		loaderInstrs, err := assembler.LoadAssemblyFile("diskLoader.asm")
		if err != nil {
			fmt.Printf("Error loading diskLoader: %v\n", path, err)
			continue
		}

		loaderCode, err := assembler.ConvertInstructionsToBinary(loaderInstrs)
		if err != nil {
			fmt.Printf("Error assembling diskLoader: %v\n", path, err)
			continue
		}
		copy(m.GetMemory()[:len(loaderCode)], loaderCode)

		m.GetRegisters()[core.PC] = 0
		m.Boot()
		DebugRegisters(m)
		DebugMemory(m)
	}
}
