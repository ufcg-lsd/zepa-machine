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

	debugMode := len(os.Args) >= 3 && os.Args[2] == "debug"

	machine := machine.NewMachine(1073741824, debugMode)
	machine.LoadProgram(binaryCode)
	go machine.Boot()

	// IO loop
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Comandos: d (step), reg (registradores), pcb (processos), kill <pid>, input <path>")
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
			machine.DebugRegisters()
			machine.Mutex.Unlock()

		case "reg":
			if !machine.IsDebugMode() {
				fmt.Printf("Máquina não está em debug mode!\n")
				continue
			}
			machine.DebugRegisters()

		case "pcb":
			if !machine.IsDebugMode() {
				fmt.Printf("Máquina não está em debug mode!\n")
				continue
			}
			machine.DebugSystem()
		}
	}

}
