package machine

import (
	"testing"
	"zepa-machine/core"
)

// To-do: refact to avoid duplicate code

func TestFetch(t *testing.T) {

	machine := NewMachine(2048)
	machine.memory[0] = 0b00110100
	machine.memory[1] = 0b01000011
	machine.memory[2] = 0b00001000
	machine.memory[3] = 0b00000000
	machine.fetch()

	expectedInstruction := uint32(0b00110100010000110000100000000000)
	if machine.registers[core.IR] != expectedInstruction {
		t.Errorf("Expected 0b%032b, but got 0b%032b", expectedInstruction, machine.registers[core.IR])
	}

	if machine.registers[core.PC] != 4 {
		t.Errorf("Expected PC to be 4, but got %d", machine.registers[core.PC])
	}
}

func TestDecode(t *testing.T) {
	machine := NewMachine(2048)
	machine.memory[0] = 0b00110100
	machine.memory[1] = 0b01000011
	machine.memory[2] = 0b00001000
	machine.memory[3] = 0b00000000
	machine.fetch()

	decodedInstruction := machine.decode()

	if decodedInstruction.rd != core.Register(2) {
		t.Errorf("Expected rd to be 2, but got %d", decodedInstruction.rd)
	}
	if decodedInstruction.rs1 != core.Register(3) {
		t.Errorf("Expected rs1 to be 3, but got %d", decodedInstruction.rs1)
	}
	if decodedInstruction.rs2 != core.Register(1) {
		t.Errorf("Expected rs2 to be 1, but got %d", decodedInstruction.rs2)
	}
	if decodedInstruction.funct5 != 0 {
		t.Errorf("Expected funct5 to be 0, but got %d", decodedInstruction.funct5)
	}
	if decodedInstruction.funct6 != 0 {
		t.Errorf("Expected funct6 to be 0, but got %d", decodedInstruction.funct6)
	}
}

func TestMV(t *testing.T) {
	machine := NewMachine(2048)
	inst := Instruction{opcode: (*Machine).mv, rd: core.W0, immediate: 0xFF}
	machine.execute(inst)

	if machine.registers[core.W0] != 0xFF {
		t.Errorf("Expected w0 to be 255, got %d", machine.registers[core.W0])
	}
}

func TestADD(t *testing.T) {
	machine := NewMachine(2048)
	machine.registers[core.W1] = 66
	machine.registers[core.W2] = 3000
	inst := Instruction{opcode: (*Machine).add, rd: core.W0, rs1: core.W1, rs2: core.W2}
	machine.execute(inst)

	if machine.registers[core.W0] != 3066 {
		t.Errorf("Expected w0 to be 3066, got %d", machine.registers[core.W0])
	}
}

func TestSUB(t *testing.T) {
	machine := NewMachine(2048)
	machine.registers[core.W1] = 30
	machine.registers[core.W2] = 10
	inst := Instruction{opcode: (*Machine).sub, rd: core.W0, rs1: core.W1, rs2: core.W2}
	machine.execute(inst)

	if machine.registers[core.W0] != 20 {
		t.Errorf("Expected w0 to be 20, got %d", machine.registers[core.W0])
	}
}

func TestJUMP(t *testing.T) {
	machine := NewMachine(2048)
	inst := Instruction{opcode: (*Machine).jump, immediate: 0xA}
	machine.execute(inst)

	if machine.registers[core.PC] != 0xA {
		t.Errorf("Expected pc to be 10, got %d", machine.registers[core.PC])
	}
}

func TestLOAD(t *testing.T) {
	machine := NewMachine(2048)
	machine.memory[256] = 42 // Definindo um valor na memória para ser carregado
	inst := Instruction{opcode: (*Machine).load, rd: core.W1, immediate: 256}
	machine.execute(inst)

	if machine.registers[core.W1] != 42 {
		t.Errorf("Expected w1 to be 42, got %d", machine.registers[core.W1])
	}
}

func TestSTORE(t *testing.T) {
	machine := NewMachine(2048)
	machine.registers[core.W1] = 65
	inst := Instruction{opcode: (*Machine).store, rd: core.W1, immediate: 100}
	machine.execute(inst)

	if machine.memory[100] != 65 {
		t.Errorf("Expected memory value to be 65, got %d", machine.memory[100])
	}
}
