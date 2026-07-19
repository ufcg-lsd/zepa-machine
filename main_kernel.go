package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
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
		return "w6"
	case 7:
		return "w7"
	case 8:
		return "w8"
	case 9:
		return "w9"
	case 10:
		return "pc"
	case 11:
		return "sp"
	case 12:
		return "ir"
	case 13:
		return "sr"
	case 14:
		return "mdr"
	case 15:
		return "mar"
	case 16:
		return "ecr"
	case 17:
		return "esa"
	case 18:
		return "esr"
	case 19:
		return "epc"
	case 20:
		return "base"
	case 21:
		return "limit"
	default:
		return "invalid"
	}
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run ./main_kernel.go <memory_size> <partition_size> <time_slice> [--no-debug]")
		return
	}

	sourceFile := "sackOS/kernel.asm"
	binaryCode, err := assembler.RunAssembler(sourceFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	debugMode := len(os.Args) < 5 || os.Args[4] != "--no-debug"

	memory_size, err := strconv.Atoi(os.Args[1])

	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	partition_size, err := strconv.Atoi(os.Args[2])

	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	time_slice, err := strconv.Atoi(os.Args[3])

	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	machine := machine.NewMachine(memory_size, debugMode)
	machine.LoadProgram(binaryCode)

	memSlice := machine.GetMemory()[4096:4100]
	binary.LittleEndian.PutUint32(memSlice, uint32(partition_size))

	memSlice = machine.GetMemory()[4100:4104]
	binary.LittleEndian.PutUint32(memSlice, uint32(time_slice))

	memSlice = machine.GetMemory()[4108:4112]
	binary.LittleEndian.PutUint32(memSlice, uint32(0x10000))

	go machine.Boot()

	// IO loop
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("ZEPA Machine — digite 'kill <pid>', 'input <path>' ou 'd' (debug)")
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
			if !machine.IsDebugMode() {
				fmt.Printf("Máquina não está em debug mode!\n")
				continue
			}
			DebugRegisters(machine)
			DebugMemory(machine)
			machine.Mutex.Unlock()
		}
	}

}
