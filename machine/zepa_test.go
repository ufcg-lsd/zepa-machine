package machine

import (
	"testing"
)

// To-do: refact to avoid duplicate code

func TestFetch(t *testing.T) {

	machine := NewMachine(2048, false)
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
	machine := NewMachine(2048, false)
	machine.memory[0] = 0b01001000
	machine.memory[1] = 0b01000011
	machine.memory[2] = 0b00001000
	machine.memory[3] = 0b00000000
	machine.fetch()

	decodedInstruction, _ := machine.decode()

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
	machine := NewMachine(2048, false)
	inst := Instruction{opcode: MV, rd: w0, immediate: 0xFF}
	machine.execute(inst)

	if machine.registers[w0] != 0xFF {
		t.Errorf("Expected w0 to be 255, got %d", machine.registers[w0])
	}
}

func TestAND(t *testing.T) {
	machine := NewMachine(2048, false)

	machine.registers[w1] = 12
	machine.registers[w2] = 10

	inst := Instruction{opcode: AND, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w0] != 8 {
		t.Errorf("Expected w0 to be 8, got %d", machine.registers[w0])
	}
}

func TestOR(t *testing.T) {
	machine := NewMachine(2048, false)

	machine.registers[w1] = 12
	machine.registers[w2] = 10

	inst := Instruction{opcode: OR, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w0] != 14 {
		t.Errorf("Expected w0 to be 14, got %d", machine.registers[w0])
	}
}

func TestXOR(t *testing.T) {
	machine := NewMachine(2048, false)

	machine.registers[w1] = 12
	machine.registers[w2] = 10

	inst := Instruction{opcode: XOR, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w0] != 6 {
		t.Errorf("Expected w0 to be 6, got %d", machine.registers[w0])
	}
}

func TestSHL(t *testing.T) {
	machine := NewMachine(2048, false)

	machine.registers[w1] = 15
	machine.registers[w2] = 2

	inst := Instruction{opcode: SHL, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w0] != 60 { // 15 << 2 = 60
		t.Errorf("Expected w0 to be 60, got %d", machine.registers[w0])
	}

	machine.registers[w1] = 0xFFFFFFF0 // -16 in two's complement
	machine.registers[w2] = 0xFFFFFFFE // -2 in two's complement

	inst = Instruction{opcode: SHL, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	expected := uint32(0x3FFFFFFC)
	if machine.registers[w0] != expected {
		t.Errorf("Expected w0 to be 0x%X, got 0x%X", expected, machine.registers[w0])
	}
}

func TestSHA(t *testing.T) {
	machine := NewMachine(2048, false)

	machine.registers[w1] = 15
	machine.registers[w2] = 2

	inst := Instruction{opcode: SHA, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w0] != 60 { // 15 << 2 = 60
		t.Errorf("Expected w0 to be 60, got %d", machine.registers[w0])
	}

	machine.registers[w1] = 0xFFFFFFF0 // -16 in two's complement
	machine.registers[w2] = 0xFFFFFFFE // -2 in two's complement

	inst = Instruction{opcode: SHA, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	expected := uint32(0xFFFFFFFC) // -4 in two's complement
	if machine.registers[w0] != expected {
		t.Errorf("Expected w0 to be 0x%X, got 0x%X", expected, machine.registers[w0])
	}
}

func TestADD(t *testing.T) {
	machine := NewMachine(2048, false)
	machine.registers[w1] = 66
	machine.registers[w2] = 3000
	inst := Instruction{opcode: ADD, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w0] != 3066 {
		t.Errorf("Expected w0 to be 3066, got %d", machine.registers[w0])
	}
}

func TestSUB(t *testing.T) {
	machine := NewMachine(2048, false)
	machine.registers[w1] = 30
	machine.registers[w2] = 10
	inst := Instruction{opcode: SUB, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w0] != 20 {
		t.Errorf("Expected w0 to be 20, got %d", machine.registers[w0])
	}
}

func TestMUL(t *testing.T) {
	machine := NewMachine(2048, false)
	machine.registers[w1] = 12
	machine.registers[w2] = 4
	inst := Instruction{opcode: MUL, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w0] != 48 {
		t.Errorf("Expected w0 to be 48, got %d", machine.registers[w0])
	}
}

func TestUDIV(t *testing.T) {
	machine := NewMachine(2048, false)
	machine.registers[w1] = 20
	machine.registers[w2] = 3
	inst := Instruction{opcode: UDIV, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w0] != 6 {
		t.Errorf("Expected w0 to be 6, got %d", machine.registers[w0])
	}
}

func TestSDIV(t *testing.T) {
	machine := NewMachine(2048, false)

	var numerator int32 = -15
	var denominator int32 = -5
	machine.registers[w1] = uint32(numerator)
	machine.registers[w2] = 3
	inst := Instruction{opcode: SDIV, rd: w0, rs1: w1, rs2: w2}
	machine.execute(inst)

	expected := uint32(denominator)

	if machine.registers[w0] != expected {
		t.Errorf("Expected w0 to be %d (-5), got %d", expected, machine.registers[w0])
	}
}

func TestJUMP(t *testing.T) {
	machine := NewMachine(2048, false)
	inst := Instruction{opcode: JUMP, immediate: 0xA}
	machine.execute(inst)

	if machine.registers[pc] != 36 {
		t.Errorf("Expected pc to be 36, got %d", machine.registers[pc])
	}
}

func TestJMPR(t *testing.T) {
	machine := NewMachine(2048, false)

	expectedAddress := uint32(128)
	machine.registers[w5] = expectedAddress

	inst := Instruction{opcode: JMPR, rs1: w5}
	machine.execute(inst)

	if machine.registers[pc] != expectedAddress {
		t.Errorf("Expected pc to be %d, got %d", expectedAddress, machine.registers[pc])
	}
}

func TestBEQ(t *testing.T) {
	machine := NewMachine(2048, false)
	inst := Instruction{opcode: BEQ, immediate: 5} // Jump offset of 5 ((5-1) * 4 = 16 bytes)

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
	machine := NewMachine(2048, false)
	inst := Instruction{opcode: BLT, immediate: 3} // Jump offset of 3 ((3-1) * 4 = 8 bytes)

	machine.registers[pc] = 50
	machine.registers[sr] = 2
	machine.execute(inst)

	if machine.registers[pc] != 58 {
		t.Errorf("Branch Taken: Expected pc to be 58, got %d", machine.registers[pc])
	}

	machine.registers[pc] = 50
	machine.registers[sr] = 4
	machine.execute(inst)

	if machine.registers[pc] != 50 {
		t.Errorf("Branch Not Taken: Expected pc to remain 50, got %d", machine.registers[pc])
	}
}

func TestBGT(t *testing.T) {
	machine := NewMachine(2048, false)
	offset := -4
	inst := Instruction{opcode: BGT, immediate: uint16(offset)} // Offset of -4 ((-4-1) * 4 = -20 bytes)

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
	machine := NewMachine(2048, false)

	machine.memory[256] = 0x78
	machine.memory[257] = 0x56
	machine.memory[258] = 0x34
	machine.memory[259] = 0x12

	machine.registers[w2] = 256
	inst := Instruction{opcode: LOAD, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w1] != 0x12345678 {
		t.Errorf("Expected w1 to be 0x12345678, got 0x%x", machine.registers[w1])
	}
}

func TestSTORE(t *testing.T) {
	machine := NewMachine(2048, false)

	machine.registers[w1] = 0x12345678
	machine.registers[w2] = 100

	inst := Instruction{opcode: STORE, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.memory[100] != 0x78 {
		t.Errorf("Expected memory[100] to be 0x78, got 0x%x", machine.memory[100])
	}
	if machine.memory[101] != 0x56 {
		t.Errorf("Expected memory[101] to be 0x56, got 0x%x", machine.memory[101])
	}
	if machine.memory[102] != 0x34 {
		t.Errorf("Expected memory[102] to be 0x34, got 0x%x", machine.memory[102])
	}
	if machine.memory[103] != 0x12 {
		t.Errorf("Expected memory[103] to be 0x12, got 0x%x", machine.memory[103])
	}
}

func TestLDD(t *testing.T) {
	machine := NewMachine(2048, false)

	machine.memory[256] = 0x78
	machine.memory[257] = 0x56
	machine.memory[258] = 0x34
	machine.memory[259] = 0x12

	inst := Instruction{opcode: LDD, rd: w1, immediate: 256}
	machine.execute(inst)

	if machine.registers[w1] != 0x12345678 {
		t.Errorf("Expected w1 to be 0x12345678, got 0x%x", machine.registers[w1])
	}
}

func TestSTRD(t *testing.T) {
	machine := NewMachine(2048, false)

	machine.registers[w1] = 0x12345678

	inst := Instruction{opcode: STRD, rd: w1, immediate: 100}
	machine.execute(inst)

	if machine.memory[100] != 0x78 {
		t.Errorf("Expected memory[100] to be 0x78, got 0x%x", machine.memory[100])
	}
	if machine.memory[101] != 0x56 {
		t.Errorf("Expected memory[101] to be 0x56, got 0x%x", machine.memory[101])
	}
	if machine.memory[102] != 0x34 {
		t.Errorf("Expected memory[102] to be 0x34, got 0x%x", machine.memory[102])
	}
	if machine.memory[103] != 0x12 {
		t.Errorf("Expected memory[103] to be 0x12, got 0x%x", machine.memory[103])
	}
}

func TestLDB(t *testing.T) {
	machine := NewMachine(2048, false)
	machine.memory[256] = 200
	machine.registers[w2] = 256
	inst := Instruction{opcode: LDB, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.registers[w1] != 200 {
		t.Errorf("Expected w1 to be 200, got %d", machine.registers[w1])
	}
}

func TestLDSB(t *testing.T) {
	machine := NewMachine(2048, false)
	value := -66
	machine.memory[256] = byte(value)
	machine.registers[w2] = 256
	inst := Instruction{opcode: LDSB, rs1: w1, rs2: w2}
	machine.execute(inst)

	// We expect the CPU to sign-extend the byte to a 32-bit integer (-66)
	expected := uint32(value)

	if machine.registers[w1] != expected {
		t.Errorf("Expected w1 to be %d (-66 sign-extended), got %d", uint32(expected), machine.registers[w1])
	}
}

func TestSTRB(t *testing.T) {
	machine := NewMachine(2048, false)
	// The register contains a 32-bit value.
	// STRB should only extract and store the lowest byte (0xDD = 221).
	machine.registers[w1] = 0xAABBCCDD
	machine.registers[w2] = 100
	inst := Instruction{opcode: STRB, rs1: w1, rs2: w2}
	machine.execute(inst)

	if machine.memory[100] != 0xDD {
		t.Errorf("Expected memory value to be 221 (0xDD), got %d", machine.memory[100])
	}

	// Ensure the CPU didn't accidentally write a full word and overwrite adjacent memory
	if machine.memory[101] != 0 {
		t.Errorf("Expected adjacent memory at block 101 to remain 0, got %d", machine.memory[101])
	}
}

func TestMRET(t *testing.T) {
	machine := NewMachine(2048, false)

	// Set up the exception state that we expect to be restored
	expectedSR := uint32(1)    // Example status register state
	expectedPC := uint32(1024) // Example return address

	machine.registers[esr] = expectedSR
	machine.registers[epc] = expectedPC

	// Set current state to something different to ensure it gets overwritten
	machine.registers[sr] = 0
	machine.registers[pc] = 512

	inst := Instruction{opcode: MRET}
	machine.execute(inst)

	if machine.registers[sr] != expectedSR {
		t.Errorf("Expected sr to be restored to %d, got %d", expectedSR, machine.registers[sr])
	}

	if machine.registers[pc] != expectedPC {
		t.Errorf("Expected pc to be restored to %d, got %d", expectedPC, machine.registers[pc])
	}
}

func TestSYSCALL(t *testing.T) {
	machine := NewMachine(2048, false)

	// Set up the initial state before the syscall
	initialPC := uint32(256)
	initialSR := uint32(1)
	expectedESA := uint32(512)

	machine.registers[pc] = initialPC
	machine.registers[sr] = initialSR
	machine.registers[esa] = expectedESA

	syscallCode := uint16(10)
	inst := Instruction{opcode: SYSCALL, immediate: syscallCode}

	machine.execute(inst)

	if machine.registers[w9] != uint32(syscallCode) {
		t.Errorf("Expected w5 to hold syscall code %d, got %d", syscallCode, machine.registers[w5])
	}

	if machine.registers[ecr] != syscallExc {
		t.Errorf("Expected ecr to be %d (syscallInt), got %d", syscallExc, machine.registers[ecr])
	}

	if machine.registers[esr] != initialSR {
		t.Errorf("Expected esr to back up initial sr %d, got %d", initialSR, machine.registers[esr])
	}

	if machine.registers[sr] != 0 {
		t.Errorf("Expected sr to be set to 0, got %d", machine.registers[sr])
	}

	if machine.registers[epc] != initialPC {
		t.Errorf("Expected epc to back up initial pc %d, got %d", initialPC, machine.registers[epc])
	}

	if machine.registers[pc] != expectedESA {
		t.Errorf("Expected pc to jump to esa %d, got %d", expectedESA, machine.registers[pc])
	}
}
