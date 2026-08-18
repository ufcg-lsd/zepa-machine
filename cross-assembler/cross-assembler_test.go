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
		0b00011000, 0b00100010, 0b00011000, 0b00000000, // ADD W1, W2, W3
		0b00000000, 0b00100000, 0b00000000, 0b10100000, // MV W1, #5
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

func TestAddTwoNumber(t *testing.T) {
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
		0b00000000, 0b00100000, 0b00000000, 0b10100000, // MV W1, #5
		0b00000000, 0b01000000, 0b00000000, 0b01100000, // MV W2, #3
		0b00011000, 0b00000001, 0b00010000, 0b00000000, // ADD W0, W1, W2
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

func TestBitwiseOperations(t *testing.T) {
	assemblyFilePath := "../asm/samples/bitwise_operations.asm"

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
		0b00000000, 0b00000000, 0b00000000, 0b11100000, // MV W0, #7
		0b00000000, 0b00100000, 0b00000000, 0b10100000, // MV W1, #5
		0b00000100, 0b01000000, 0b00001000, 0b00000000, // AND W2, W0, W1
		0b00001000, 0b01100000, 0b00001000, 0b00000000, // OR W3, W0, W1
		0b00001100, 0b10000000, 0b00001000, 0b00000000, // XOR W4, W0, W1
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Incorrect memory size. Expected: %d, Got: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Incorrect memory at byte %d (Instruction %d). Expected: 0b%08b, Got: 0b%08b", i, (i/4)+1, expectedMemory[i], byteVal)
		}
	}
}

func TestBranches(t *testing.T) {
	assemblyFilePath := "../asm/samples/branches.asm"

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
		0b00000000, 0b00100000, 0b00000001, 0b11100000, // MV W1, #15
		0b00000000, 0b01000000, 0b00000010, 0b10000000, // MV W2, #20
		0b00000000, 0b10100000, 0b00000010, 0b00000000, // MV W5, #16
		0b00110000, 0b00000000, 0b00000000, 0b01000000, // JUMP COMPARE_FUNC
		0b00110000, 0b00000000, 0b00000001, 0b01100000, // JUMP _end
		0b00101100, 0b00000001, 0b00010000, 0b00000000, // CMP W1, W2
		0b00111000, 0b00000000, 0b00000000, 0b01100000, // BEQ SET_EQUAL
		0b00111100, 0b00000000, 0b00000000, 0b10000000, // BLT SET_LESS
		0b01000000, 0b00000000, 0b00000000, 0b10100000, // BGT SET_GREATER
		0b00000000, 0b00000000, 0b00000000, 0b00100000, // MV W0, #0
		0b00110100, 0b00000101, 0b00000000, 0b00000000, // JMPR W5
		0b00000000, 0b00000000, 0b00000000, 0b01000000, // MV W0, #1
		0b00110100, 0b00000101, 0b00000000, 0b00000000, // JMPR W5
		0b00000000, 0b00000000, 0b00000000, 0b10000000, // MV W0, #2
		0b00110100, 0b00000101, 0b00000000, 0b00000000, // JMPR W5
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

func TestDivideNegatives(t *testing.T) {
	assemblyFilePath := "../asm/samples/divide_negatives.asm"

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
		0b00000000, 0b00111111, 0b11111110, 0b11000000, // MV W1, #-10
		0b00000000, 0b01000000, 0b00000000, 0b01100000, // MV W2, #3
		0b00101000, 0b00000001, 0b00010000, 0b00000000, // SDIV W0, W1, W2
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

func TestDividePositives(t *testing.T) {
	assemblyFilePath := "../asm/samples/divide_positives.asm"

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
		0b00000000, 0b00100000, 0b00000001, 0b11100000, // MV W1, #15
		0b00000000, 0b01000000, 0b00000000, 0b01100000, // MV W2, #3
		0b00100100, 0b00000001, 0b00010000, 0b00000000, // UDIV W0, W1, W2
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

func TestMret(t *testing.T) {
	assemblyFilePath := "../asm/samples/mret.asm"

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
		0b01100000, 0b00000000, 0b00000000, 0b00000000, // MRET
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

func TestMultiplyTwoNumbers(t *testing.T) {
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
		0b00000000, 0b00100000, 0b00000000, 0b10000000, // MV W1, #4
		0b00000000, 0b01000000, 0b00000000, 0b11000000, // MV W2, #6
		0b00000000, 0b00000000, 0b00000000, 0b00000000, // MV W0, #0
		0b00011000, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (1st time)
		0b00011000, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (2nd time)
		0b00011000, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (3rd time)
		0b00011000, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (4th time)
		0b00011000, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (5th time)
		0b00011000, 0b00000000, 0b00001000, 0b00000000, // ADD W0, W0, W1 (6th time)
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

func TestMultiplyWithMul(t *testing.T) {
	assemblyFilePath := "../asm/samples/multiply_with_mul.asm"

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
		0b00000000, 0b00100000, 0b00000000, 0b10000000, // MV W1, #4
		0b00000000, 0b01000000, 0b00000000, 0b11000000, // MV W2, #6
		0b00100000, 0b00000001, 0b00010000, 0b00000000, // MUL W0, W1, W2
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

func TestShifts(t *testing.T) {
	assemblyFilePath := "../asm/samples/shifts.asm"

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
		0b00000000, 0b00111111, 0b11111111, 0b10000000, // MV W1, #-4
		0b00000000, 0b01000000, 0b00000000, 0b01000000, // MV W2, #2
		0b00000000, 0b01111111, 0b11111111, 0b11000000, // MV W3, #-2
		0b00010000, 0b10000001, 0b00010000, 0b00000000, // SHL W4, W1, W2
		0b00010000, 0b10100001, 0b00011000, 0b00000000, // SHL W5, W1, W3
		0b00010100, 0b11000001, 0b00010000, 0b00000000, // SHA W6, W1, W2
		0b00010100, 0b11100001, 0b00011000, 0b00000000, // SHA W7, W1, W3
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Incorrect memory size. Expected: %d, Got: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Incorrect memory at block %d.\nExpected: 0b%08b\nGot:      0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestSimpleJump(t *testing.T) {
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
		0b00000000, 0b00100000, 0b00000000, 0b01000000, // MV W1, #2
		0b00000000, 0b01000000, 0b00000000, 0b10100000, // MV W2, #5
		0b00110000, 0b00000000, 0b00000000, 0b01100000, // JUMP 0x14
		0b00000000, 0b00100000, 0b00000011, 0b11000000, // MV W1, #30 (this instruction is skipped due to jump)
		0b00000000, 0b01000000, 0b00000101, 0b00000000, // MV W2, #40 (this instruction is skipped due to jump)
		0b00011000, 0b00000001, 0b00010000, 0b00000000, // ADD W0, W1, W2
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

func TestStoreAndLoadByte(t *testing.T) {
	assemblyFilePath := "../asm/samples/store_load_byte.asm"

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
		0b00000000, 0b01110000, 0b00000000, 0b00000000, // MV W3, #-32768
		0b00000000, 0b00100000, 0b00000100, 0b00000000, // MV W1, #0x020
		0b01011100, 0b00000011, 0b00001000, 0b00000000, // STRB W3, W1
		0b01010100, 0b00000010, 0b00001000, 0b00000000, // LDB W2, W1
		0b01011000, 0b00000100, 0b00001000, 0b00000000, // LDSB W4, W1
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

func TestStoreAndLoadDWord(t *testing.T) {
	assemblyFilePath := "../asm/samples/store_load_direct.asm"

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
		0b00000000, 0b01100000, 0b00001000, 0b01000000, // MV W3, #66
		0b00000000, 0b00100000, 0b00000100, 0b00000000, // MV W1, #0x020
		0b01010000, 0b01100000, 0b00000100, 0b00000000, // STRD W3, #0x020
		0b01001100, 0b01000000, 0b00000100, 0b00000000, // LDD W2, #0x020
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
		0b00000000, 0b01100000, 0b00001000, 0b01000000, // MV W3, #66
		0b00000000, 0b00100000, 0b00000100, 0b00000000, // MV W1, #0x020
		0b01001000, 0b00000011, 0b00001000, 0b00000000, // STORE W3, W1
		0b01000100, 0b00000010, 0b00001000, 0b00000000, // LOAD W2, W1
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
		0b00000000, 0b01100000, 0b00000111, 0b10000000, // MV W3, #60
		0b00000000, 0b10100000, 0b00000100, 0b01100000, // MV W5, #35
		0b00011100, 0b01100011, 0b00101000, 0b00000000, // SUB W3, W3, W5
		0b00101100, 0b00000101, 0b00011000, 0b00000000, // CMP W5, W3
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

// Auxiliary function for running the assembler using a reader
func RunAssemblerFromReader(reader *bytes.Buffer) ([]byte, error) {
	instrs, err := LoadAssemblyFromReader(reader)
	if err != nil {
		return nil, err
	}
	return ConvertInstructionsToBinary(instrs)
}

func TestSyscall(t *testing.T) {
	assemblyFilePath := "../asm/samples/syscall.asm"

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
		0b01100100, 0b00000000, 0b00000000, 0b01000000, // SYSCALL #2
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
