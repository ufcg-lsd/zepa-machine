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

	memorySize, err := strconv.Atoi(os.Args[1])

	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	if memorySize%4 != 0 {
		log.Fatalf("memory_size must be a multiple of 4, got %d", memorySize)
	}

	partitionSize, err := strconv.Atoi(os.Args[2])

	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	if partitionSize%4 != 0 {
		log.Fatalf("partition_size must be a multiple of 4, got %d", partitionSize)
	}

	timeSlice, err := strconv.Atoi(os.Args[3])

	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	machine := machine.NewMachine(memorySize, debugMode)
	machine.LoadProgram(binaryCode)

	memSlice := machine.GetMemory()[4096:4100]
	binary.LittleEndian.PutUint32(memSlice, uint32(partitionSize))

	memSlice = machine.GetMemory()[4100:4104]
	binary.LittleEndian.PutUint32(memSlice, uint32(timeSlice))

	memSlice = machine.GetMemory()[4108:4112]
	binary.LittleEndian.PutUint32(memSlice, uint32(0x10000))

	go machine.Boot()

	// IO loop
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Commands: d (step), r (restore), reg (registers), pcb (processes), kill <pid>, input <path>")
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
			fmt.Printf("kill %d sent\n", pid)

		case "input":
			if len(parts) < 2 {
				fmt.Println("usage: input <path>")
				continue
			}
			binaryCode, err := assembler.RunAssembler(parts[1])
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			machine.LoadBuffer(binaryCode)
			machine.SetInputFlag()
			fmt.Printf("input sent\n")

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
				for {
					machine.StepChan <- struct{}{}
					<-machine.DoneChan

					if machine.GetRegisters()[10] == uint32(pcValue) {
						break
					}
				}

				machine.DebugRegisters()
			}

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
