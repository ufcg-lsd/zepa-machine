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
	if decodedInstruction.funct5 != 0 {
		t.Errorf("Expected funct5 to be 0, but got %d", decodedInstruction.funct5)
	}
	if decodedInstruction.funct6 != 0 {
		t.Errorf("Expected funct6 to be 0, but got %d", decodedInstruction.funct6)
	}
}

func TestMV(t *testing.T) {
	machine := NewMachine(2048)
	inst := Instruction{opcode: (*Machine).mv, rd: w0, immediate: 0xFF}
	machine.execute(inst)

	if machine.registers[w0] != 0xFF {
		t.Errorf("Expected w0 to be 255, got %d", machine.registers[w0])
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

func TestMUL(t *testing.T) {
	machine := NewMachine(2048)
	machine.registers[w1] = 12
	machine.registers[w2] = 4
	inst := Instruction{opcode: (*Machine).mul, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w0] != 48 {
		t.Errorf("Expected w0 to be 48, got %d", machine.registers[w0])
	}
}

func TestUDIV(t *testing.T) {
	machine := NewMachine(2048)
	machine.registers[w1] = 20
	machine.registers[w2] = 3
	inst := Instruction{opcode: (*Machine).udiv, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w0] != 6 {
		t.Errorf("Expected w0 to be 6, got %d", machine.registers[w0])
	}
}

func TestSDIV(t *testing.T) {
	machine := NewMachine(2048)

	var numerator int32 = -15
	var denominator int32 = -5
	machine.registers[w1] = uint32(numerator)
	machine.registers[w2] = 3
	inst := Instruction{opcode: (*Machine).sdiv, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	expected := uint32(denominator)

	if machine.registers[w0] != expected {
		t.Errorf("Expected w0 to be %d (-5), got %d", expected, machine.registers[w0])
	}
}

func TestJUMP(t *testing.T) {
	machine := NewMachine(2048)
	inst := Instruction{opcode: (*Machine).jump, immediate: 0xA}
	machine.execute(inst)

	if machine.registers[pc] != 36 {
		t.Errorf("Expected pc to be 36, got %d", machine.registers[pc])
	}
}

func TestJMPR(t *testing.T) {
	machine := NewMachine(2048)

	expectedAddress := uint32(128)
	machine.registers[w5] = expectedAddress

	inst := Instruction{opcode: (*Machine).jmpr, rs1: w5}
	machine.execute(inst)

	if machine.registers[pc] != expectedAddress {
		t.Errorf("Expected pc to be %d, got %d", expectedAddress, machine.registers[pc])
	}
}

func TestBEQ(t *testing.T) {
	machine := NewMachine(2048)
	inst := Instruction{opcode: (*Machine).beq, immediate: 5} // Jump offset of 5 ((5-1) * 4 = 16 bytes)

	machine.registers[pc] = 100
	machine.registers[sr] = 1
	machine.execute(inst)

	if machine.registers[pc] != 116 {
		t.Errorf("Branch Taken: Expected pc to be 116, got %d", machine.registers[pc])
	}

	machine.registers[pc] = 100
	machine.registers[sr] = 10
	machine.execute(inst)

	if machine.registers[pc] != 100 {
		t.Errorf("Branch Not Taken: Expected pc to remain 100, got %d", machine.registers[pc])
	}
}

func TestBLT(t *testing.T) {
	machine := NewMachine(2048)
	inst := Instruction{opcode: (*Machine).blt, immediate: 3} // Jump offset of 3 ((3-1) * 4 = 8 bytes)

	machine.registers[pc] = 50
	machine.registers[sr] = 2
	machine.execute(inst)

	if machine.registers[pc] != 58 {
		t.Errorf("Branch Taken: Expected pc to be 58, got %d", machine.registers[pc])
	}

	machine.registers[pc] = 50
	machine.registers[sr] = 10
	machine.execute(inst)

	if machine.registers[pc] != 50 {
		t.Errorf("Branch Not Taken: Expected pc to remain 50, got %d", machine.registers[pc])
	}
}

func TestBGT(t *testing.T) {
	machine := NewMachine(2048)
	offset := -4
	inst := Instruction{opcode: (*Machine).bgt, immediate: uint16(offset)} // Offset of -4 ((-4-1) * 4 = -20 bytes)

	machine.registers[pc] = 200
	machine.registers[sr] = 4
	machine.execute(inst)

	if machine.registers[pc] != 180 {
		t.Errorf("Branch Taken: Expected pc to be 180, got %d", machine.registers[pc])
	}

	machine.registers[pc] = 200
	machine.registers[sr] = 10
	machine.execute(inst)

	if machine.registers[pc] != 200 {
		t.Errorf("Branch Not Taken: Expected pc to remain 200, got %d", machine.registers[pc])
	}
}

func TestLOAD(t *testing.T) {
	machine := NewMachine(2048)
	machine.memory[256] = 42 // Definindo um valor na memória para ser carregado
	machine.registers[w2] = 256
	inst := Instruction{opcode: (*Machine).load, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w1] != 42 {
		t.Errorf("Expected w1 to be 42, got %d", machine.registers[w1])
	}
}

func TestSTORE(t *testing.T) {
	machine := NewMachine(2048)
	machine.registers[w1] = 65
	machine.registers[w2] = 100
	inst := Instruction{opcode: (*Machine).store, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.memory[100] != 65 {
		t.Errorf("Expected memory value to be 65, got %d", machine.memory[100])
	}
}

func TestLDB(t *testing.T) {
	machine := NewMachine(2048)
	machine.memory[256] = 200
	machine.registers[w2] = 256
	inst := Instruction{opcode: (*Machine).ldb, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w1] != 200 {
		t.Errorf("Expected w1 to be 200, got %d", machine.registers[w1])
	}
}

func TestLDSB(t *testing.T) {
	machine := NewMachine(2048)
	value := -66
	machine.memory[256] = byte(value)
	machine.registers[w2] = 256
	inst := Instruction{opcode: (*Machine).ldsb, rs1: w1, rs2: w2}
	machine.execute(inst)

	// We expect the CPU to sign-extend the byte to a 32-bit integer (-66)
	expected := uint32(value)

	if machine.registers[w1] != expected {
		t.Errorf("Expected w1 to be %d (-66 sign-extended), got %d", uint32(expected), machine.registers[w1])
	}
}

func TestSTRB(t *testing.T) {
	machine := NewMachine(2048)
	// The register contains a 32-bit value.
	// STRB should only extract and store the lowest byte (0xDD = 221).
	machine.registers[w1] = 0xAABBCCDD
	machine.registers[w2] = 100
	inst := Instruction{opcode: (*Machine).strb, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.memory[100] != 0xDD {
		t.Errorf("Expected memory value to be 221 (0xDD), got %d", machine.memory[100])
	}

	// Ensure the CPU didn't accidentally write a full word and overwrite adjacent memory
	if machine.memory[101] != 0 {
		t.Errorf("Expected adjacent memory at block 101 to remain 0, got %d", machine.memory[101])
	}
}
