package machine

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
	funct5Length = 5
	funct6Length = 6
	immediateLen = 16
	word         = 32
)

const (
	opcodeBitMask    = 0b111111
	registerBitMask  = 0b11111
	immediateBitMask = 0b1111111111111111
	funct5BitMask    = 0b11111
	funct6BitMask    = 0b111111
)

var operations = map[byte]Operation{
	byte(MV):    (*Machine).mv,
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
	funct5    byte
	funct6    byte
	immediate uint16
}

type Machine struct {
	memory    []byte
	registers map[Register]uint32
}

func (m *Machine) mv(inst Instruction) {
	m.registers[inst.rd] = uint32(inst.immediate)
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
		m.registers[sr] = 2
	} else {
		m.registers[sr] = 1
	}
}

func (m *Machine) jump(inst Instruction) {
	m.registers[pc] = uint32(inst.immediate)
}

func (m *Machine) load(inst Instruction) {
	m.registers[inst.rs1] = uint32(inst.immediate)
}

func (m *Machine) store(inst Instruction) {
	m.memory[inst.immediate] = byte(m.registers[inst.rs1])
}

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
	offSetFunct5 := offSetRs2 - funct5Length
	offSetFunct6 := offSetFunct5 - funct6Length

	opcode := instruction >> (uint32(offsetOpcode)) & opcodeBitMask
	rd := (instruction >> uint32(offSetRd)) & registerBitMask
	rs1 := (instruction >> (uint32(offSetRs1))) & registerBitMask
	rs2 := (instruction >> (uint32(offSetRs2))) & registerBitMask
	funct5 := (instruction >> (uint32(offSetFunct5))) & funct5BitMask
	funct6 := (instruction >> (uint32(offSetFunct6))) & funct6BitMask

	operation := operations[byte(opcode)]

	return Instruction{
		opcode: operation,
		rd:     Register(rd),
		rs1:    Register(rs1),
		rs2:    Register(rs2),
		funct5: byte(funct5),
		funct6: byte(funct6),
	}
}

func (m *Machine) decodeITypeInst(instruction uint32) Instruction {
	offsetOpcode := word - opcodeLength
	offSetRdRs1 := offsetOpcode - rdLength
	offSetImmediate := offSetRdRs1 - immediateLen
	offSetFunct5 := offSetImmediate - funct5Length

	opcode := instruction >> (uint32(offsetOpcode)) & opcodeBitMask
	rdRs1 := (instruction >> uint32(offSetRdRs1)) & registerBitMask
	immediate := (instruction >> (uint32(offSetImmediate))) & immediateBitMask
	funct5 := (instruction >> (uint32(offSetFunct5))) & funct5BitMask

	operation := operations[byte(opcode)]

	return Instruction{
		opcode:    operation,
		rd:        Register(rdRs1),
		immediate: uint16(immediate),
		funct5:    byte(funct5),
	}
}

func (m *Machine) isEndOfProgram() bool {
	if (m.registers[ir]) == 0 {
		m.registers[pc] -= 4
		return true
	}
	return false
}

func (m *Machine) getOpcode(instruction uint32) Opcode {
	offsetOpcode := word - opcodeLength
	opcode := instruction >> (uint32(offsetOpcode))

	return Opcode(opcode)
}

func (m *Machine) decode() Instruction {
	instruction := m.registers[ir]
	opcode := m.getOpcode(instruction)

	switch opcode {
	case ADD, SUB, CMP:
		return m.decodeRTypeInst(instruction)
	case MV, JUMP, LOAD, STORE:
		fallthrough
	default:
		return m.decodeITypeInst(instruction)
	}
}

func (m *Machine) execute(inst Instruction) {
	inst.opcode(m, inst)
}

func (m *Machine) Boot() {
	for {
		m.fetch()
		if m.isEndOfProgram() {
			break
		}
		decodedInstruction := m.decode()
		m.execute(decodedInstruction)
	}
}

func (m *Machine) LoadProgram(program []byte) {
	copy(m.memory, program)
}

func (m *Machine) GetMemory() []byte {
	return m.memory
}

func (m *Machine) GetRegisters() map[Register]uint32 {
	return m.registers
}

func NewMachine(memoryBytes int) *Machine {
	machine := &Machine{
		memory:    make([]byte, memoryBytes),
		registers: make(map[Register]uint32),
	}

	return machine
}
