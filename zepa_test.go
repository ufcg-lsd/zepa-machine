package machine

import (
	"testing"
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
	if machine.registers[ir] != expectedInstruction {
		t.Errorf("Expected 0b%032b, but got 0b%032b", expectedInstruction, machine.registers[ir])
	}

	if machine.registers[pc] != 4 {
		t.Errorf("Expected PC to be 4, but got %d", machine.registers[pc])
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

	if decodedInstruction.rd != Register(2) {
		t.Errorf("Expected rd to be 2, but got %d", decodedInstruction.rd)
	}
	if decodedInstruction.rs1 != Register(3) {
		t.Errorf("Expected rs1 to be 3, but got %d", decodedInstruction.rs1)
	}
	if decodedInstruction.rs2 != Register(1) {
		t.Errorf("Expected rs2 to be 1, but got %d", decodedInstruction.rs2)
	}
	if decodedInstruction.funct3 != 0 {
		t.Errorf("Expected funct3 to be 0, but got %d", decodedInstruction.funct3)
	}
	if decodedInstruction.funct7 != 0 {
		t.Errorf("Expected funct7 to be 0, but got %d", decodedInstruction.funct7)
	}
}

func TestMV(t *testing.T) {
	machine := NewMachine(2048)
	inst := Instruction{opcode: (*Machine).mv, rd: w0, addrConst: 0xFF}
	machine.execute(inst)

	if machine.registers[w0] != 0xFF {
		t.Errorf("Expected w0 to be 42, got %d", machine.registers[w0])
	}
}

func TestADD(t *testing.T) {
	machine := NewMachine(2048)
	machine.registers[w1] = 66
	machine.registers[w2] = 3000
	inst := Instruction{opcode: (*Machine).add, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w0] != 3066 {
		t.Errorf("Expected w0 to be 3066, got %d", machine.registers[w0])
	}
}

func TestSUB(t *testing.T) {
	machine := NewMachine(2048)
	machine.registers[w1] = 30
	machine.registers[w2] = 10
	inst := Instruction{opcode: (*Machine).sub, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w0] != 20 {
		t.Errorf("Expected w0 to be 20, got %d", machine.registers[w0])
	}
}

func TestJUMP(t *testing.T) {
	machine := NewMachine(2048)
	inst := Instruction{opcode: (*Machine).jump, addrConst: 0xA}
	machine.execute(inst)

	if machine.registers[pc] != 0xA {
		t.Errorf("Expected pc to be 10, got %d", machine.registers[pc])
	}
}

func TestLOAD(t *testing.T) {
	machine := NewMachine(2048)
	inst := Instruction{opcode: (*Machine).load, rs1: w1, addrConst: 256}
	machine.execute(inst)

	if machine.registers[w1] != 256 {
		t.Errorf("Expected w1 to be 256, got %d", machine.registers[w1])
	}
}

func TestSTORE(t *testing.T) {
	machine := NewMachine(2048)
	machine.registers[w1] = 65
	inst := Instruction{opcode: (*Machine).store, rs1: w1, addrConst: 100}
	machine.execute(inst)

	if machine.memory[100] != 65 {
		t.Errorf("Expected memory value to be 65, got %d", machine.memory[100])
	}
}
