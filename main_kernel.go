//go:build kernel

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

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run -tags kernel . <memory_size> <partition_size> <time_slice> [--tui] [--no-debug]")
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

	partitionSize, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	timeSlice, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	// Parse optional flags
	useTUI := false
	debugMode := true
	for _, arg := range os.Args[4:] {
		switch arg {
		case "--tui":
			useTUI = true
		case "--no-debug":
			debugMode = false
		}
	}

	// TUI always requires debug mode
	if useTUI {
		debugMode = true
	}

	m := machine.NewMachine(memorySize, debugMode)
	m.LoadProgram(binaryCode)

	mem := m.GetMemory()
	binary.LittleEndian.PutUint32(mem[0x1000:0x1004], uint32(partitionSize))
	binary.LittleEndian.PutUint32(mem[0x1004:0x1008], uint32(timeSlice))
	binary.LittleEndian.PutUint32(mem[0x1008:0x100C], uint32(0x10000))

	go m.Boot()

	if useTUI {
		runTUIWithMachine(m)
		return
	}

	// CLI loop
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Comandos: d [n] (step), reg (registradores), pcb (processos), kill <pid>, input <path>")
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
			var buf [4]byte
			binary.LittleEndian.PutUint32(buf[:], uint32(pid))
			m.LoadBuffer(buf[:])
			m.SetKillFlag()
			fmt.Printf("kill %d enviado\n", pid)

		case "input":
			if len(parts) < 2 {
				fmt.Println("uso: input <path>")
				continue
			}
			code, err := assembler.RunAssembler(parts[1])
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			m.LoadBuffer(code)
			m.SetInputFlag()
			fmt.Printf("input enviado\n")

		case "d":
			if !m.IsDebugMode() {
				fmt.Printf("Máquina não está em debug mode!\n")
				continue
			}

			steps := 1
			if len(parts) > 1 {
				parsedSteps, err := strconv.Atoi(parts[1])
				if err != nil || parsedSteps <= 0 {
					fmt.Println("Número de passos inválido. Executando 1 passo.")
				} else {
					steps = parsedSteps
				}
			}

			for i := 0; i < steps; i++ {
				m.StepChan <- struct{}{}
				<-m.DoneChan
			}

			m.DebugRegisters()

		case "reg":
			if !m.IsDebugMode() {
				fmt.Printf("Máquina não está em debug mode!\n")
				continue
			}
			m.DebugRegisters()

		case "pcb":
			if !m.IsDebugMode() {
				fmt.Printf("Máquina não está em debug mode!\n")
				continue
			}
			m.DebugSystem()
		}
	}
}
