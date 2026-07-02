package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	fmt.Print("\n\n")
}

func DebugRegisters(m *machine.Machine) {
	registers := m.GetRegisters()
	fmt.Println("\n----------Registers----------")

	for k, v := range registers {
		// ignore IR
		if k == 8 {
			continue
		}
		fmt.Printf("%v: %d\n", getRegisterName(k), int32(v))
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
	case 12:
		return "ecr"
	case 13:
		return "esa"
	case 14:
		return "esr"
	case 15:
		return "epc"
	case 16:
		return "base"
	case 17:
		return "limit"
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

	machine := machine.NewMachine(1073741824, true)
	machine.LoadProgram(binaryCode)
	go machine.Boot()

	// IO loop
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("ZEPA Machine — digite 'kill <pid>', 'input <path>' ou 'd'")
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "kill":
			if len(parts) < 2 {
				fmt.Println("uso: kill <pid>")
				continue
			}
			pid, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("PID inválido")
				continue
			}
			buffer := make([]byte, 4)
			binary.LittleEndian.PutUint32(buffer, uint32(pid))
			machine.LoadBuffer(buffer)
			machine.SetKillFlag()
			fmt.Printf("kill %d enviado\n", pid)

		case "input":
			if len(parts) < 2 {
				fmt.Println("uso: input <path>")
				continue
			}
			binaryCode, err := assembler.RunAssembler(parts[1])
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			machine.LoadBuffer(binaryCode)
			machine.SetInputFlag()
			fmt.Printf("input enviado\n")

		case "d":
			DebugRegisters(machine)
			DebugMemory(machine)
			machine.Mutex.Unlock()
		}
	}

}
