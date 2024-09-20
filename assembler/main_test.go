package main

import (
	"bytes"
	"fmt"
	"testing"
)

// Example programs
var addTwoNumbersAssembly = `
MV W1, #5       ; Load the number 5 into W1
MV W2, #3       ; Load the number 3 into W2
ADD W0, W1, W2  ; Add W1 and W2, result in W0
`

var sumListAssembly = `
MV W0, #0       ; Initialize sum as 0
MV W1, #1       ; Example number 1
MV W2, #2       ; Example number 2
MV W3, #3       ; Example number 3

ADD W0, W0, W1  ; Add W1
ADD W0, W0, W2  ; Add W2
ADD W0, W0, W3  ; Add W3
`

var multiplyAssembly = `
MV W1, #4       ; Load the number 4 into W1
MV W2, #6       ; Load the number 6 into W2
MV W0, #0       ; Initialize W0 with 0

ADD W0, W0, W1  ; W0 = W0 + W1 (1st time)
ADD W0, W0, W1  ; W0 = W0 + W1 (2nd time)
ADD W0, W0, W1  ; W0 = W0 + W1 (3rd time)
ADD W0, W0, W1  ; W0 = W0 + W1 (4th time)
ADD W0, W0, W1  ; W0 = W0 + W1 (5th time)
ADD W0, W0, W1  ; W0 = W0 + W1 (6th time)
`

// Helper function to assemble the code and get the binary output in bit strings
func assembleCode(t *testing.T, assemblyCode string) []string {
	// Simulate file reading
	reader := bytes.NewBufferString(assemblyCode)
	instrs, err := LoadAssemblyFromReader(reader)
	if err != nil {
		t.Fatalf("Error loading assembly code: %v", err)
	}

	instructionMemory, err := ConvertInstructionsToBinary(instrs)
	if err != nil {
		t.Fatalf("Error executing assembler: %v", err)
	}

	// Convert instructions to binary strings
	var output []string
	for _, instr := range instructionMemory {
		output = append(output, fmt.Sprintf("%032b", instr))
	}

	return output
}

// Test for the add_two_numbers program
func TestAddTwoNumbers(t *testing.T) {
	expectedOutput := []string{
		"00011000001000000000000010100000", // MV W1, #5
		"00011000010000000000000001100000", // MV W2, #3
		"00011100000000010001000000000000", // ADD W0, W1, W2
	}

	output := assembleCode(t, addTwoNumbersAssembly)

	if len(output) != len(expectedOutput) {
		t.Fatalf("Incorrect number of instructions. Expected %d, got %d", len(expectedOutput), len(output))
	}

	for i, instr := range output {
		if instr != expectedOutput[i] {
			t.Errorf("Incorrect instruction %d.\nExpected: %s\nGot:  %s", i, expectedOutput[i], instr)
		}
	}
}

// Test for the sum_list program
func TestSumList(t *testing.T) {
	expectedOutput := []string{
		"00011000000000000000000000000000", // MV W0, #0
		"00011000001000000000000000100000", // MV W1, #1
		"00011000010000000000000001000000", // MV W2, #2
		"00011000011000000000000001100000", // MV W3, #3
		"00011100000000000000100000000000", // ADD W0, W0, W1
		"00011100000000000001000000000000", // ADD W0, W0, W2
		"00011100000000000001100000000000", // ADD W0, W0, W3
	}

	output := assembleCode(t, sumListAssembly)

	if len(output) != len(expectedOutput) {
		t.Fatalf("Incorrect number of instructions. Expected %d, got %d", len(expectedOutput), len(output))
	}

	for i, instr := range output {
		if instr != expectedOutput[i] {
			t.Errorf("Incorrect instruction %d.\nExpected: %s\nGot:  %s", i, expectedOutput[i], instr)
		}
	}
}

// Test for the multiply program
func TestMultiply(t *testing.T) {
	expectedOutput := []string{
		"00011000001000000000000010000000", // MV W1, #4
		"00011000010000000000000011000000", // MV W2, #6
		"00011000000000000000000000000000", // MV W0, #0
		"00011100000000000000100000000000", // ADD W0, W0, W1 (1st time)
		"00011100000000000000100000000000", // ADD W0, W0, W1 (2nd time)
		"00011100000000000000100000000000", // ADD W0, W0, W1 (3rd time)
		"00011100000000000000100000000000", // ADD W0, W0, W1 (4th time)
		"00011100000000000000100000000000", // ADD W0, W0, W1 (5th time)
		"00011100000000000000100000000000", // ADD W0, W0, W1 (6th time)
	}

	output := assembleCode(t, multiplyAssembly)

	if len(output) != len(expectedOutput) {
		t.Fatalf("Incorrect number of instructions. Expected %d, got %d", len(expectedOutput), len(output))
	}

	for i, instr := range output {
		if instr != expectedOutput[i] {
			t.Errorf("Incorrect instruction %d.\nExpected: %s\nGot:  %s", i, expectedOutput[i], instr)
		}
	}
}
