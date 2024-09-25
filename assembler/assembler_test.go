package assembler

import (
	"bytes"
	"os"
	"testing"
)

func TestRunAssemblerAndMemory(t *testing.T) {
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
		t.Fatalf("Incorrect memory size. Expected: %d, Got: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Incorrect memory at block %d. Expected: 0b%08b, Got: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestSUBAndCMP(t *testing.T) {
	assemblyFilePath := "../asm/samples/sub_and_cmp.asm"

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
		0b00111000, 0b10000101, 0b00011000, 0b00000000, // SUB W4, W5, W3
		0b00111100, 0b00000000, 0b00001000, 0b00000000, // CMP W0, W1
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Incorrect memory size. Expected: %d, Got: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Incorrect memory at block %d. Expected: 0b%08b, Got: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestJUMPAndLOAD(t *testing.T) {
	assemblyFilePath := "../asm/samples/jump_and_load.asm"

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
		0b01000000, 0b00000000, 0b00000010, 0b00000000, // JUMP 0x0010
		0b01000100, 0b01000000, 0b00000100, 0b00000000, // LOAD W2, 0x0020
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Incorrect memory size. Expected: %d, Got: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Incorrect memory at block %d. Expected: 0b%08b, Got: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestSTORE(t *testing.T) {
	assemblyFilePath := "../asm/samples/store.asm"

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
		0b01001000, 0b01100000, 0b00001000, 0b00000000, // STORE W3, 0x0040
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Incorrect memory size. Expected: %d, Got: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Incorrect memory at block %d. Expected: 0b%08b, Got: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestAddTwo(t *testing.T) {
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
		t.Fatalf("Incorrect memory size. Expected: %d, Got: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Incorrect memory at block %d. Expected: 0b%08b, Got: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestMultiply(t *testing.T) {
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
		0b00110100, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (1ª vez)
		0b00110100, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (2ª vez)
		0b00110100, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (3ª vez)
		0b00110100, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (4ª vez)
		0b00110100, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (5ª vez)
		0b00110100, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (6ª vez)
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Incorrect memory size. Expected: %d, Got: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Incorrect memory at block %d. Expected: 0b%08b, Got: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestSumList(t *testing.T) {
	assemblyFilePath := "../asm/samples/sum_list.asm"

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
		0b00110000, 0b00000000, 0b00000000, 0b00000000, // MV W0, #0
		0b00110000, 0b00100000, 0b00000000, 0b00100000, // MV W1, #1
		0b00110000, 0b01000000, 0b00000000, 0b01000000, // MV W2, #2
		0b00110000, 0b01100000, 0b00000000, 0b01100000, // MV W3, #3
		0b00110100, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1
		0b00110100, 0b00000000, 0b00010000, 0b00000000, // ADD W0, W0, W2
		0b00110100, 0b00000000, 0b00011000, 0b00000000, // ADD W0, W0, W3
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Incorrect memory size. Expected: %d, Got: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Incorrect memory at block %d. Expected: 0b%08b, Got: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestInstructions(t *testing.T) {
	assemblyFilePath := "../asm/samples/all_instructions.asm"

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
		0b01000100, 0b00100010, 0b00000000, 0b00000000, // LOAD W1, 0x1000
		0b01001000, 0b01000100, 0b00000000, 0b00000000, // STORE W2, 0x2000
		0b00110100, 0b01100100, 0b00101000, 0b00000000, // ADD W3, W4, W5
		0b00111000, 0b10100001, 0b00010000, 0b00000000, // SUB W5, W1, W2
		0b00111100, 0b00000011, 0b00100000, 0b00000000, // CMP W3, W4
		0b01000000, 0b00000110, 0b00000000, 0b00000000, // JUMP 0x3000
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Incorrect memory size. Expected: %d, Got: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Incorrect memory at block %d. Expected: 0b%08b, Got: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

// auxiliary function for running the assembler using a reader
func RunAssemblerFromReader(reader *bytes.Buffer) ([]byte, error) {
	instrs, err := LoadAssemblyFromReader(reader)
	if err != nil {
		return nil, err
	}
	return ConvertInstructionsToBinary(instrs)
}
