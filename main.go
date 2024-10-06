package main

import (
	"fmt"
	"os"
	"zepa-machine/assembler"
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

func getRegisterName(reg machine.Register) string {
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
	default:
		return "invalid"
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./main.go <asm/file/path>")
		return
	}

	sourceFile := os.Args[1]
	binaryCode, err := assembler.RunAssembler(sourceFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	machine := machine.NewMachine(64)
	machine.LoadProgram(binaryCode)
	machine.Boot()

	DebugRegisters(machine)
	DebugMemory(machine)
}
