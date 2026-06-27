package machine

import "slices"

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
	ecr
	esa
	esr
	epc
)

const (
	MV Opcode = iota
	ADD
	SUB
	MUL
	UDIV
	SDIV
	CMP
	JUMP
	JMPR
	BEQ
	BLT
	BGT
	LOAD
	STORE
	LDB
	LDSB
	STRB
	MRET
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

const (
	clockInt uint32 = iota
	inputInt
	syscallInt
	faultInt
)

const TIMER_INTERVAL = 128

var operations = map[Opcode]Operation{
	MV:    (*Machine).mv,
	ADD:   (*Machine).add,
	SUB:   (*Machine).sub,
	MUL:   (*Machine).mul,
	UDIV:  (*Machine).udiv,
	SDIV:  (*Machine).sdiv,
	CMP:   (*Machine).cmp,
	JUMP:  (*Machine).jump,
	JMPR:  (*Machine).jmpr,
	BEQ:   (*Machine).beq,
	BLT:   (*Machine).blt,
	BGT:   (*Machine).bgt,
	LOAD:  (*Machine).load,
	STORE: (*Machine).store,
	LDB:   (*Machine).ldb,
	LDSB:  (*Machine).ldsb,
	STRB:  (*Machine).strb,
	MRET:  (*Machine).mret,
}

type Instruction struct {
	opcode    Opcode
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
	m.registers[inst.rd] = uint32(int16(inst.immediate))
}

func (m *Machine) add(inst Instruction) {
	m.registers[inst.rd] = m.registers[inst.rs1] + m.registers[inst.rs2]
}

func (m *Machine) sub(inst Instruction) {
	m.registers[inst.rd] = m.registers[inst.rs1] - m.registers[inst.rs2]
}

func (m *Machine) mul(inst Instruction) {
	m.registers[inst.rd] = m.registers[inst.rs1] * m.registers[inst.rs2]
}

func (m *Machine) udiv(inst Instruction) {
	m.registers[inst.rd] = m.registers[inst.rs1] / m.registers[inst.rs2]
}

func (m *Machine) sdiv(inst Instruction) {
	m.registers[inst.rd] = uint32(int32(m.registers[inst.rs1]) / int32(m.registers[inst.rs2]))
}

func (m *Machine) cmp(inst Instruction) {
	if m.registers[inst.rs1] == m.registers[inst.rs2] {
		m.registers[sr] = 1
	} else if m.registers[inst.rs1] > m.registers[inst.rs2] {
		m.registers[sr] = 4
	} else {
		m.registers[sr] = 2
	}
}

func (m *Machine) jump(inst Instruction) {
	m.registers[pc] += (uint32(int16(inst.immediate)) - 1) * 4
}

func (m *Machine) jmpr(inst Instruction) {
	m.registers[pc] = m.registers[inst.rs1]
}

func (m *Machine) beq(inst Instruction) {
	if m.registers[sr] == 1 {
		m.jump(inst)
	}
}

func (m *Machine) blt(inst Instruction) {
	if m.registers[sr] == 2 {
		m.jump(inst)
	}
}

func (m *Machine) bgt(inst Instruction) {
	if m.registers[sr] == 4 {
		m.jump(inst)
	}
}

func (m *Machine) load(inst Instruction) {
	addr := m.registers[inst.rs2]
	m.registers[inst.rs1] = 0

	for i := uint32(0); i < 4; i++ {
		m.registers[inst.rs1] |= (uint32(m.memory[addr+i]) << (i * 8))
	}
}

func (m *Machine) store(inst Instruction) {
	addr := m.registers[inst.rs2]

	for i := uint32(0); i < 4; i++ {
		m.memory[addr+i] = byte(m.registers[inst.rs1] >> (i * 8))
	}
}

func (m *Machine) ldb(inst Instruction) {
	m.registers[inst.rs1] = uint32(m.memory[m.registers[inst.rs2]])
}

func (m *Machine) ldsb(inst Instruction) {
	m.registers[inst.rs1] = uint32(int8(m.memory[m.registers[inst.rs2]]))
}

func (m *Machine) strb(inst Instruction) {
	m.memory[m.registers[inst.rs2]] = byte(m.registers[inst.rs1])
}

func (m *Machine) mret(inst Instruction) {
	m.registers[sr] = m.registers[esr]

	m.registers[pc] = m.registers[epc]
}

func (m *Machine) exception(cause uint32) {
	m.registers[ecr] = cause

	m.registers[esr] = m.registers[sr]
	m.registers[sr] = 0

	m.registers[epc] = m.registers[pc]
	m.registers[pc] = m.registers[esa]
}

func (m *Machine) checkIllegalRegisterAccess(inst Instruction) bool {
	priviligedRegisters := []Register{ecr, esa, esr, epc}
	return !m.isKernelMode() && (slices.Contains(priviligedRegisters, inst.rd) || slices.Contains(priviligedRegisters, inst.rs1) || slices.Contains(priviligedRegisters, inst.rs2))
}

func (m *Machine) checkIllegalInstruction(inst Instruction) bool {
	priviligedInstructions := []Opcode{MRET}
	return !m.isKernelMode() && slices.Contains(priviligedInstructions, inst.opcode)
}

func (m *Machine) isKernelMode() bool {
	return m.registers[sr]&0x8 == 0
}

func (m *Machine) isInterruptEnabled() bool {
	return m.registers[sr]&0x8 == 1
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

	return Instruction{
		opcode: Opcode(opcode),
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

	return Instruction{
		opcode:    Opcode(opcode),
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
	case ADD, SUB, MUL, UDIV, SDIV, CMP, JMPR, LOAD, STORE, LDB, LDSB, STRB:
		return m.decodeRTypeInst(instruction)
	case MV, JUMP, BEQ, BLT, BGT, MRET:
		fallthrough
	default:
		return m.decodeITypeInst(instruction)
	}
}

func (m *Machine) execute(inst Instruction) {
	operations[inst.opcode](m, inst)
}

func (m *Machine) Boot() {

	instructionsExcecuted := 0

	for {

		if m.isInterruptEnabled() && instructionsExcecuted%TIMER_INTERVAL == 0 {
			m.exception(clockInt)
		}

		m.fetch()
		if m.isEndOfProgram() {
			break
		}
		decodedInstruction := m.decode()

		if m.isInterruptEnabled() {
			if m.checkIllegalRegisterAccess(decodedInstruction) {
				m.exception(faultInt)
				continue
			}
		}

		m.execute(decodedInstruction)
		instructionsExcecuted++
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
