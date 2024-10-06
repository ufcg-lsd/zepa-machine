package assembler

import (
	"bytes"
	"os"
	"testing"
)

func TestAddAndMv(t *testing.T) {
	assemblyFilePath := "../asm/samples/add_and_mv.asm"

	assemblyCode, err := os.ReadFile(assemblyFilePath)
	if err != nil {
		t.Fatalf("Error reading the assembly file: %v", err)
	}

	file := bytes.NewBuffer(assemblyCode)

	memory, err := RunAssemblerFromReader(file)
	if err != nil {
		t.Fatalf("Error running the assembler: %v", err)
	}

	expectedMemory := []byte{
		0b00110100, 0b00100010, 0b00011000, 0b00000000, // ADD W1, W2, W3
		0b00110000, 0b00100000, 0b00000000, 0b10100000, // MV W1, #5
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Incorrect memory size. Expected: %d, saiu: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Incorrect memory at block %d. Expected: 0b%08b, saiu: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestProgram2(t *testing.T) {
	assemblyFilePath := "../asm/samples/add_two_number.asm"

	assemblyCode, err := os.ReadFile(assemblyFilePath)
	if err != nil {
		t.Fatalf("Error reading the assembly file: %v", err)
	}

	file := bytes.NewBuffer(assemblyCode)

	memory, err := RunAssemblerFromReader(file)
	if err != nil {
		t.Fatalf("Error running the assembler: %v", err)
	}

	expectedMemory := []byte{
		0b00110000, 0b00100000, 0b00000000, 0b10100000, // MV W1, #5
		0b00110000, 0b01000000, 0b00000000, 0b01100000, // MV W2, #3
		0b00110100, 0b00000001, 0b00010000, 0b00000000, // ADD W0, W1, W2
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Incorrect memory size. Expected: %d, saiu: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Incorrect memory at block %d. Expected: 0b%08b, saiu: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestProgram3(t *testing.T) {
	assemblyFilePath := "../asm/samples/multiply_two_numbers.asm"

	assemblyCode, err := os.ReadFile(assemblyFilePath)
	if err != nil {
		t.Fatalf("Error reading the assembly file: %v", err)
	}

	file := bytes.NewBuffer(assemblyCode)

	memory, err := RunAssemblerFromReader(file)
	if err != nil {
		t.Fatalf("Error running the assembler: %v", err)
	}

	expectedMemory := []byte{
		0b00110000, 0b00100000, 0b00000000, 0b10000000, // MV W1, #4
		0b00110000, 0b01000000, 0b00000000, 0b11000000, // MV W2, #6
		0b00110000, 0b00000000, 0b00000000, 0b00000000, // MV W0, #0
		0b00110100, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (1st time)
		0b00110100, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (2nd time)
		0b00110100, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (3rd time)
		0b00110100, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (4th time)
		0b00110100, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (5th time)
		0b00110100, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (6th time)
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Incorrect memory size. Expected: %d, saiu: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Incorrect memory at block %d. Expected: 0b%08b, saiu: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestProgram4(t *testing.T) {
	assemblyFilePath := "../asm/samples/simple_jump.asm"

	assemblyCode, err := os.ReadFile(assemblyFilePath)
	if err != nil {
		t.Fatalf("Error reading the assembly file: %v", err)
	}

	file := bytes.NewBuffer(assemblyCode)

	memory, err := RunAssemblerFromReader(file)
	if err != nil {
		t.Fatalf("Error running the assembler: %v", err)
	}

	expectedMemory := []byte{
		0b00110000, 0b00100000, 0b00000000, 0b01000000, // MV W1, #2
		0b00110000, 0b01000000, 0b00000000, 0b10100000, // MV W2, #5
		0b01000000, 0b00000000, 0b00000010, 0b10000000, // JUMP 0x14
		0b00110000, 0b00100000, 0b00000011, 0b11000000, // MV W1, #30 (this instruction is skipped due to jump)
		0b00110000, 0b01000000, 0b00000101, 0b00000000, // MV W2, #40 (this instruction is skipped due to jump)
		0b00110100, 0b00000001, 0b00010000, 0b00000000, // ADD W0, W1, W2
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Incorrect memory size. Expected: %d, saiu: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Incorrect memory at block %d. Expected: 0b%08b, saiu: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestStoreAndLoad(t *testing.T) {
	assemblyFilePath := "../asm/samples/store_load.asm"

	assemblyCode, err := os.ReadFile(assemblyFilePath)
	if err != nil {
		t.Fatalf("Error reading the assembly file: %v", err)
	}

	file := bytes.NewBuffer(assemblyCode)

	memory, err := RunAssemblerFromReader(file)
	if err != nil {
		t.Fatalf("Error running the assembler: %v", err)
	}

	expectedMemory := []byte{
		0b00110000, 0b01100000, 0b00001000, 0b01000000, // MV W3, #66
		0b01001000, 0b01100000, 0b00000100, 0b00000000, // STORE W3, 0x020
		0b01000100, 0b01000000, 0b00000100, 0b00000000, // LOAD W2, 0x020
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Incorrect memory size. Expected: %d, saiu: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Incorrect memory at block %d. Expected: 0b%08b, saiu: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestSubAndCmp(t *testing.T) {
	assemblyFilePath := "../asm/samples/sub_cmp_bigger.asm"

	assemblyCode, err := os.ReadFile(assemblyFilePath)
	if err != nil {
		t.Fatalf("Error reading the assembly file: %v", err)
	}

	file := bytes.NewBuffer(assemblyCode)

	memory, err := RunAssemblerFromReader(file)
	if err != nil {
		t.Fatalf("Error running the assembler: %v", err)
	}

	expectedMemory := []byte{
		0b00110000, 0b01100000, 0b00000111, 0b10000000, // MV W3, #60
		0b00110000, 0b10100000, 0b00000100, 0b01100000, // MV W5, #35
		0b00111000, 0b01100011, 0b00101000, 0b00000000, // SUB W3, W3, W5
		0b00111100, 0b00000101, 0b00011000, 0b00000000, // CMP W5, W3
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Incorrect memory size. Expected: %d, saiu: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Incorrect memory at block %d. Expected: 0b%08b, saiu: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

// Auxiliary function for running the assembler using a reader
func RunAssemblerFromReader(reader *bytes.Buffer) ([]byte, error) {
	instrs, err := LoadAssemblyFromReader(reader)
	if err != nil {
		return nil, err
	}
	return ConvertInstructionsToBinary(instrs)
}
