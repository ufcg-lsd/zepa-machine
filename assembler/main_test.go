package main

import (
	"bytes"
	"testing"
)

func TestRunAssemblerAndMemory(t *testing.T) {
	assemblyCode := `
		ADD W1, W2, W3
		MV W1, #5
	`
	file := bytes.NewBufferString(assemblyCode)

	memory = []byte{}

	err := RunAssemblerFromReader(file)
	if err != nil {
		t.Fatalf("Erro ao rodar o assembler: %v", err)
	}

	expectedMemory := []byte{
		0b00110100, 0b00100010, 0b00011000, 0b00000000, // ADD W1, W2, W3
		0b00110000, 0b00100000, 0b00000000, 0b10100000, // MV W1, #5
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Tamanho da memória incorreto. Esperado: %d, Obtido: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Memória incorreta no bloco %d. Esperado: 0b%08b, Obtido: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestSUBAndCMP(t *testing.T) {
	assemblyCode := `
		SUB W4, W5, W3
		CMP W0, W1
	`
	file := bytes.NewBufferString(assemblyCode)

	memory = []byte{}

	err := RunAssemblerFromReader(file)
	if err != nil {
		t.Fatalf("Erro ao rodar o assembler: %v", err)
	}

	expectedMemory := []byte{
		0b00111000, 0b10000101, 0b00011000, 0b00000000, // SUB W4, W5, W3
		0b00111100, 0b00000000, 0b00001000, 0b00000000, // CMP W0, W1
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Tamanho da memória incorreto. Esperado: %d, Obtido: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Memória incorreta no bloco %d. Esperado: 0b%08b, Obtido: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestJUMPAndLOAD(t *testing.T) {
	assemblyCode := `
		JUMP 0x0010
		LOAD W2, 0x0020
	`
	file := bytes.NewBufferString(assemblyCode)

	memory = []byte{}

	err := RunAssemblerFromReader(file)
	if err != nil {
		t.Fatalf("Erro ao rodar o assembler: %v", err)
	}

	expectedMemory := []byte{
		0b01000000, 0b00000000, 0b00000010, 0b00000000, // JUMP 0x0010
		0b01000100, 0b01000000, 0b00000100, 0b00000000, // LOAD W2, 0x0020
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Tamanho da memória incorreto. Esperado: %d, Obtido: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Memória incorreta no bloco %d. Esperado: 0b%08b, Obtido: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestSTORE(t *testing.T) {
	assemblyCode := `
		STORE W3, 0x0040
	`
	file := bytes.NewBufferString(assemblyCode)

	memory = []byte{}

	err := RunAssemblerFromReader(file)
	if err != nil {
		t.Fatalf("Erro ao rodar o assembler: %v", err)
	}

	expectedMemory := []byte{
		0b01001000, 0b01100000, 0b00001000, 0b00000000, // STORE W3, 0x0040
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Tamanho da memória incorreto. Esperado: %d, Obtido: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Memória incorreta no bloco %d. Esperado: 0b%08b, Obtido: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestAddTwo(t *testing.T) {
	assemblyCode := `
		MV W1, #5
		MV W2, #3
		ADD W0, W1, W2
	`
	file := bytes.NewBufferString(assemblyCode)

	memory = []byte{}

	err := RunAssemblerFromReader(file)
	if err != nil {
		t.Fatalf("Erro ao rodar o assembler: %v", err)
	}

	expectedMemory := []byte{
		0b00110000, 0b00100000, 0b00000000, 0b10100000, // MV W1, #5
		0b00110000, 0b01000000, 0b00000000, 0b01100000, // MV W2, #3
		0b00110100, 0b00000001, 0b00010000, 0b00000000, // ADD W0, W1, W2
	}

	if len(memory) != len(expectedMemory) {
		t.Fatalf("Tamanho da memória incorreto. Esperado: %d, Obtido: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Memória incorreta no bloco %d. Esperado: 0b%08b, Obtido: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestMultiply(t *testing.T) {
	assemblyCode := `
		MV W1, #4
		MV W2, #6
		MV W0, #0
		ADD W0, W0, W1
		ADD W0, W0, W1
		ADD W0, W0, W1
		ADD W0, W0, W1
		ADD W0, W0, W1
		ADD W0, W0, W1
	`
	file := bytes.NewBufferString(assemblyCode)

	memory = []byte{}

	err := RunAssemblerFromReader(file)
	if err != nil {
		t.Fatalf("Erro ao rodar o assembler: %v", err)
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
		t.Fatalf("Tamanho da memória incorreto. Esperado: %d, Obtido: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Memória incorreta no bloco %d. Esperado: 0b%08b, Obtido: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestSumList(t *testing.T) {
	assemblyCode := `
		MV W0, #0
		MV W1, #1
		MV W2, #2
		MV W3, #3
		ADD W0, W0, W1
		ADD W0, W0, W2
		ADD W0, W0, W3
	`
	file := bytes.NewBufferString(assemblyCode)

	memory = []byte{}

	err := RunAssemblerFromReader(file)
	if err != nil {
		t.Fatalf("Erro ao rodar o assembler: %v", err)
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
		t.Fatalf("Tamanho da memória incorreto. Esperado: %d, Obtido: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Memória incorreta no bloco %d. Esperado: 0b%08b, Obtido: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

func TestInstructions(t *testing.T) {
	assemblyCode := `
		LOAD W1, 0x1000
		STORE W2, 0x2000
		ADD W3, W4, W5
		SUB W5, W1, W2
		CMP W3, W4
		JUMP 0x3000
	`
	file := bytes.NewBufferString(assemblyCode)

	memory = []byte{}

	err := RunAssemblerFromReader(file)
	if err != nil {
		t.Fatalf("Erro ao rodar o assembler: %v", err)
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
		t.Fatalf("Tamanho da memória incorreto. Esperado: %d, Obtido: %d", len(expectedMemory), len(memory))
	}

	for i, byteVal := range memory {
		if byteVal != expectedMemory[i] {
			t.Errorf("Memória incorreta no bloco %d. Esperado: 0b%08b, Obtido: 0b%08b", i, expectedMemory[i], byteVal)
		}
	}
}

// Função auxiliar para rodar o assembler usando um leitor ao invés de um arquivo real
func RunAssemblerFromReader(reader *bytes.Buffer) error {
	instrs, err := LoadAssemblyFromReader(reader)
	if err != nil {
		return err
	}
	return ConvertInstructionsToBinary(instrs)
}
