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
		fmt.Println("Usage: go run ./main.go program1.asm [program2.asm...]")
		return
	}

	machine := machine.NewMachine(64) // 64 bytes fixos

	loaderInstrs, err := assembler.LoadAssemblyFile("diskLoader.asm")
	if err != nil {
		fmt.Printf("Error loading loader: %v\n", err)
		return
	}

	loaderCode, err := assembler.ConvertInstructionsToBinary(loaderInstrs)
	if err != nil {
		fmt.Printf("Error assembling loader: %v\n", err)
		return
	}

	// Configura disco
	machine.InitDisk()
	for _, filePath := range os.Args[1:] {
		programInstrs, err := assembler.LoadAssemblyFile(filePath)
		if err != nil {
			fmt.Printf("Error loading %s: %v\n", filePath, err)
			continue
		}

		programCode, err := assembler.ConvertInstructionsToBinary(programInstrs)
		if err != nil {
			fmt.Printf("Error assembling %s: %v\n", filePath, err)
			continue
		}

		machine.AddToDisk(append(programCode, []byte{0, 0, 0, 0}...))
	}

	// Configura memória
	copy(machine.GetMemory()[:32], loaderCode) // Loader nos primeiros 32 bytes

	fmt.Println("=== Executing ===")
	machine.Boot()
	DebugRegisters(machine)
	DebugMemory(machine)
}
