package main

type Register uint32
type Opcode byte
type Operation func(m *Machine, inst Instruction)

const (
	w0 Register = iota
	w1
	w2
	w3
	w4
	w5
	pc
	sp
	ir
	sr
	mdr
	mar

	MV Opcode = iota
	ADD
	SUB
	CMP
	JUMP
	LOAD
	STORE
	FETCH
)

const (
	opcodeLength = 6
	rdLength     = 5
	rs1Length    = 5
	rs2Length    = 5
	funct3Length = 3
	funct7Length = 7
	addrConstLen = 16
	word         = 32
)

var operations = map[byte]Operation{
	byte(ADD):   (*Machine).add,
	byte(SUB):   (*Machine).sub,
	byte(CMP):   (*Machine).cmp,
	byte(JUMP):  (*Machine).jump,
	byte(LOAD):  (*Machine).load,
	byte(STORE): (*Machine).store,
}

type Instruction struct {
	opcode    func(m *Machine, inst Instruction)
	rd        Register
	rs1       Register
	rs2       Register
	funct3    byte
	funct7    byte
	addrConst uint16
}

type Machine struct {
	memory    []byte
	registers map[Register]uint32
}

func (m *Machine) mv(inst Instruction) {
	m.registers[inst.rd] = uint32(inst.addrConst)
}

func (m *Machine) add(inst Instruction) {
	m.registers[inst.rd] = m.registers[inst.rs1] + m.registers[inst.rs2]
}

func (m *Machine) sub(inst Instruction) {
	m.registers[inst.rd] = m.registers[inst.rs1] - m.registers[inst.rs2]
}

func (m *Machine) cmp(inst Instruction) {
	if m.registers[inst.rs1] == m.registers[inst.rs2] {
		m.registers[sr] = 0
	} else if m.registers[inst.rs1] > m.registers[inst.rs2] {
		m.registers[sr] = 1
	} else {
		m.registers[sr] = 2
	}
}

func (m *Machine) jump(inst Instruction) {
	m.registers[pc] = uint32(inst.addrConst)
}

func (m *Machine) load(inst Instruction) {

}

func (m *Machine) store(inst Instruction) {}

func (m *Machine) fetch() {
	var completeInstruction uint32 = 0
	for i := 0; i < 4; i++ {
		currentInstructionAddress := m.registers[pc]
		currentInstruction := m.memory[currentInstructionAddress]
		completeInstruction = completeInstruction | uint32(currentInstruction)<<(24-8*i)
		m.registers[pc] += 1
	}
	m.registers[ir] = completeInstruction
}

func (m *Machine) decodeRTypeInst(instruction uint32) Instruction {
	offsetOpcode := word - opcodeLength
	offSetRd := offsetOpcode - rdLength
	offSetRs1 := offSetRd - rs1Length
	offSetRs2 := offSetRs1 - rs2Length
	offSetFunct3 := offSetRs2 - funct3Length
	offSetFunct7 := offSetFunct3 - funct7Length

	opcode := instruction >> (uint32(offsetOpcode))
	rd := (instruction >> uint32(offSetRd)) & 0b11111
	rs1 := (instruction >> (uint32(offSetRs1))) & 0b11111
	rs2 := (instruction >> (uint32(offSetRs2))) & 0b11111
	funct3 := (instruction >> (uint32(offSetFunct3))) & 0b111
	funct7 := (instruction >> (uint32(offSetFunct7))) & 0b1111111

	operation := operations[byte(opcode)]

	return Instruction{
		opcode: operation,
		rd:     Register(rd),
		rs1:    Register(rs1),
		rs2:    Register(rs2),
		funct3: byte(funct3),
		funct7: byte(funct7),
	}
}

func (m *Machine) decodeITypeInst(instruction uint32) Instruction {
	offsetOpcode := word - opcodeLength
	offSetRdRs1 := offsetOpcode - rdLength
	offSetImmediate := offSetRdRs1 - addrConstLen
	offSetFunct3 := offSetImmediate - funct3Length

	opcode := instruction >> (uint32(offsetOpcode))
	rdRs1 := (instruction >> uint32(offSetRdRs1)) & 0b11111
	immediate := (instruction >> (uint32(offSetImmediate))) & 0b111
	funct3 := (instruction >> (uint32(offSetFunct3))) & 0b1111111

	operation := operations[byte(opcode)]

	return Instruction{
		opcode:    operation,
		rd:        Register(rdRs1),
		addrConst: uint16(immediate),
		funct3:    byte(funct3),
	}
}

func (m *Machine) checkRType(instruction uint32) bool {
	offsetOpcode := word - opcodeLength
	opcode := instruction >> (uint32(offsetOpcode))

	rTypeInsts := []Opcode{ADD, SUB, CMP}

	for _, operation := range rTypeInsts {
		if opcode == uint32(operation) {
			return true
		}
	}

	return false
}

func (m *Machine) decode() Instruction {
	instruction := m.registers[ir]

	if m.checkRType(instruction) {
		return m.decodeRTypeInst(instruction)
	}
	return m.decodeITypeInst(instruction)
}

func (m *Machine) execute(inst Instruction) {
	inst.opcode(m, inst)
}

func (m *Machine) boot() {
	for {
		m.fetch()
		decodedInstruction := m.decode()
		m.execute(decodedInstruction)
		break
	}
}

func NewMachine(memoryBytes int) *Machine {
	machine := &Machine{
		memory:    make([]byte, memoryBytes),
		registers: make(map[Register]uint32),
	}

	return machine
}

func main() {
	machine := NewMachine(2048)
	machine.memory[0] = 0b00110100
	machine.memory[1] = 0b01000011
	machine.memory[2] = 0b00001000
	machine.memory[3] = 0b00000000

	machine.registers[w3] = 3
	machine.registers[w1] = 2
	machine.boot()
}
