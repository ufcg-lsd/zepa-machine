package main

import (
	"fmt"
	"os"
	"zepa-machine/core"
	assembler "zepa-machine/cross-assembler"
	"zepa-machine/machine"
)

func DebugMemory(m *machine.Machine) {
	const bytesPerRow = 4
	memory := m.GetMemory()

	fmt.Print("\n----------Memory----------")
	for i := 0; i < len(memory); i += bytesPerRow {
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

func loadProgramsToDisk(m *machine.Machine, filePaths []string) error {
	for _, filePath := range filePaths {
		binaryCode, err := assembler.RunAssembler(filePath)
		if err != nil {
			return fmt.Errorf("error assembling %s: %v", filePath, err)
		}
		m.AddToDisk(binaryCode)
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./main.go <asm/file/path>")
		return
	}
	machine := machine.NewMachine(64)
	machine.SetDebugMode(true)
	machine.InitDisk()

	filePaths := os.Args[1:]

	// Mostrar status inicial
	fmt.Println("\n=== Initial State ===")
	machine.PrintDiskStatus()
	machine.PrintMemoryMap()

	if err := loadProgramsToDisk(machine, filePaths); err != nil {
		fmt.Printf("Error loading programs: %v\n", err)
		return
	}

	// Mostrar status após carregar para o disco
	fmt.Println("\n=== After Loading to Disk ===")
	machine.PrintDiskStatus()

	// Executar cada programa
	for i := range filePaths {
		fmt.Printf("\n=== Preparing to execute program %d ===\n", i+1)

		if err := machine.LoadFromDisk(i); err != nil {
			fmt.Printf("Error loading program %d: %v\n", i+1, err)
			continue
		}

		// Mostrar status após carregar para memória
		fmt.Println("\n=== After Loading to Memory ===")
		machine.PrintMemoryMap()

		fmt.Printf("\nExecuting program %d...\n", i+1)
		machine.Boot()

		DebugRegisters(machine)
		DebugMemory(machine)

	}
}
