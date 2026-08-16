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

const (
	buffer        = 64 * 1024       // 64KB
	minKernelSize = 8 * 1024 * 1024 // 8MB
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run -tags kernel . <memory_size> <time_slice> [--tui] [--no-debug]")
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

	if memorySize%4 != 0 {
		log.Fatalf("memory_size must be a multiple of 4, got %d", memorySize)
	}
	if memorySize < (1<<30)+(1<<12) {
		log.Fatalf("memory_size must be at least 1GB + 4KB, got %d", memorySize)
	}

	timeSlice, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	// Parse optional flags
	useTUI := false
	debugMode := true
	for _, arg := range os.Args[3:] {
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

	machine := machine.NewMachine(memorySize, debugMode)
	machine.LoadProgram(binaryCode)

	memSlice := machine.GetMemory()[0x2000:0x2004]
	binary.LittleEndian.PutUint32(memSlice, uint32(256)) // max processes

	memSlice = machine.GetMemory()[0x2004:0x2008]
	binary.LittleEndian.PutUint32(memSlice, uint32(timeSlice))

	memSlice = machine.GetMemory()[0x2008:0x200C]
	binary.LittleEndian.PutUint32(memSlice, uint32(0x10000)) // buffer size

	go machine.Boot()

	if useTUI {
		runTUIWithMachine(machine)
		return
	}

	// CLI loop
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Commands: d (step), r (restore), reg (registers), pcb (processes), b (breakpoint), c (instruction counter), kill <pid>, input <path>")
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "kill":
			if len(parts) < 2 {
				fmt.Println("usage: kill <pid>")
				continue
			}
			pid, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("invalid PID")
				continue
			}
			buffer := make([]byte, 4)
			binary.LittleEndian.PutUint32(buffer, uint32(pid))
			machine.LoadBuffer(buffer)
			machine.SetKillFlag()
			fmt.Printf("[sys] kill %d sent\n", pid)

		case "input":
			if len(parts) < 2 {
				fmt.Println("usage: input <path>")
				continue
			}
			code, err := assembler.RunAssembler(parts[1])
			if err != nil {
				fmt.Printf("[err] %v\n", err)
				continue
			}
			machine.LoadBuffer(code)
			machine.SetInputFlag()
			fmt.Printf("[sys] input sent (%d bytes)\n", len(code))

		case "d":
			if !machine.IsDebugMode() {
				fmt.Printf("Machine is not in debug mode!\n")
				continue
			}

			steps := 1
			if len(parts) > 1 {
				parsedSteps, err := strconv.Atoi(parts[1])
				if err != nil || parsedSteps <= 0 {
					fmt.Println("Invalid number of steps. Running 1 step.")
				} else {
					steps = parsedSteps
				}
			}

			machine.SaveCheckpoint()
			for i := 0; i < steps; i++ {
				machine.StepChan <- struct{}{}
				<-machine.DoneChan
			}

			machine.DebugRegisters()

		case "b":
			if !machine.IsDebugMode() {
				fmt.Printf("Machine is not in debug mode!\n")
				continue
			}

			if len(parts) > 1 {
				pcValue, err := strconv.Atoi(parts[1])
				if err != nil || pcValue < 0 {
					fmt.Println("Invalid PC.")
					continue
				}

				machine.SaveCheckpoint()
				instructionCount := 0
				for {
					machine.StepChan <- struct{}{}
					<-machine.DoneChan
					instructionCount++

					if machine.GetRegisters()[10] == uint32(pcValue) {
						break
					}
				}

				fmt.Printf("Breakpoint reached at pc=%d after %d instructions\n", pcValue, instructionCount)
				machine.DebugRegisters()
			}

		case "c":
			if !machine.IsDebugMode() {
				fmt.Printf("Machine is not in debug mode!\n")
				continue
			}
			fmt.Printf("Instructions executed since boot: %d\n", machine.GetInstructionsExecuted())

		case "r":
			if !machine.IsDebugMode() {
				fmt.Printf("Machine is not in debug mode!\n")
				continue
			}
			if !machine.RestoreCheckpoint() {
				fmt.Println("No checkpoint available.")
			} else {
				fmt.Println("Checkpoint restored!")
				machine.DebugRegisters()
			}

		case "reg":
			if !machine.IsDebugMode() {
				fmt.Printf("Machine is not in debug mode!\n")
				continue
			}
			machine.DebugRegisters()

		case "pcb":
			if !machine.IsDebugMode() {
				fmt.Printf("Machine is not in debug mode!\n")
				continue
			}
			machine.DebugSystem()
		}
	}
}
