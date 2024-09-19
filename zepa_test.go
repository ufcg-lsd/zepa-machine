package machine

import (
	"testing"
)

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

// To-do: refact to avoid code duplication
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
