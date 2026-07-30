package machine

import (
	"slices"
	"sync"
)

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
	w6
	w7
	w8
	w9
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
	base
	limit
)

const (
	MV Opcode = iota
	AND
	OR
	XOR
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
	LDD
	STRD
	LDB
	LDSB
	STRB
	MRET
	SYSCALL
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
	bufferSize = 65536 // 64 x 1024
)

const (
	clockInt uint32 = iota
	inputInt
	killInt
	syscallExc
	faultExc
)

const TIMER_INTERVAL = 128

var operations = map[Opcode]Operation{
	MV:      (*Machine).mv,
	AND:     (*Machine).and,
	OR:      (*Machine).or,
	XOR:     (*Machine).xor,
	ADD:     (*Machine).add,
	SUB:     (*Machine).sub,
	MUL:     (*Machine).mul,
	UDIV:    (*Machine).udiv,
	SDIV:    (*Machine).sdiv,
	CMP:     (*Machine).cmp,
	JUMP:    (*Machine).jump,
	JMPR:    (*Machine).jmpr,
	BEQ:     (*Machine).beq,
	BLT:     (*Machine).blt,
	BGT:     (*Machine).bgt,
	LOAD:    (*Machine).load,
	STORE:   (*Machine).store,
	LDD:     (*Machine).ldd,
	STRD:    (*Machine).strd,
	LDB:     (*Machine).ldb,
	LDSB:    (*Machine).ldsb,
	STRB:    (*Machine).strb,
	MRET:    (*Machine).mret,
	SYSCALL: (*Machine).syscall,
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
	memory     []byte
	registers  map[Register]uint32
	mu         sync.RWMutex
	killFlag   bool
	inputFlag  bool
	debugFlag  bool
	StepChan   chan struct{}
	DoneChan   chan struct{}
	checkpoint *Machine
	quitChan   chan struct{}
}

func (m *Machine) mv(inst Instruction) {
	m.registers[inst.rd] = uint32(int16(inst.immediate))
}

func (m *Machine) and(inst Instruction) {
	m.registers[inst.rd] = m.registers[inst.rs1] & m.registers[inst.rs2]
}

func (m *Machine) or(inst Instruction) {
	m.registers[inst.rd] = m.registers[inst.rs1] | m.registers[inst.rs2]
}

func (m *Machine) xor(inst Instruction) {
	m.registers[inst.rd] = m.registers[inst.rs1] ^ m.registers[inst.rs2]
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
	if m.registers[inst.rs2] == 0 {
		m.exception(faultExc)
		return
	}

	m.registers[inst.rd] = m.registers[inst.rs1] / m.registers[inst.rs2]
}

func (m *Machine) sdiv(inst Instruction) {
	if m.registers[inst.rs2] == 0 {
		m.exception(faultExc)
		return
	}

	m.registers[inst.rd] = uint32(int32(m.registers[inst.rs1]) / int32(m.registers[inst.rs2]))
}

func (m *Machine) cmp(inst Instruction) {
	var cmpMask int32 = -8
	m.registers[sr] &= uint32(cmpMask)
	if m.registers[inst.rs1] == m.registers[inst.rs2] {
		m.registers[sr] |= 1
	} else if m.registers[inst.rs1] > m.registers[inst.rs2] {
		m.registers[sr] |= 4
	} else {
		m.registers[sr] |= 2
	}
}

func (m *Machine) jump(inst Instruction) {
	m.registers[pc] += (uint32(int16(inst.immediate)) - 1) * 4
}

func (m *Machine) jmpr(inst Instruction) {
	m.registers[pc] = m.registers[inst.rs1]
}

func (m *Machine) beq(inst Instruction) {
	if m.registers[sr]&1 != 0 {
		m.jump(inst)
	}
}

func (m *Machine) blt(inst Instruction) {
	if m.registers[sr]&2 != 0 {
		m.jump(inst)
	}
}

func (m *Machine) bgt(inst Instruction) {
	if m.registers[sr]&4 != 0 {
		m.jump(inst)
	}
}

func (m *Machine) load(inst Instruction) {
	addr, ok := m.translate(m.registers[inst.rs2], 3)
	if !ok {
		return
	}

	m.registers[inst.rs1] = 0

	for i := uint32(0); i < 4; i++ {
		m.registers[inst.rs1] |= (uint32(m.memory[addr+i]) << (i * 8))
	}
}

func (m *Machine) store(inst Instruction) {
	addr, ok := m.translate(m.registers[inst.rs2], 3)
	if !ok {
		return
	}

	for i := uint32(0); i < 4; i++ {
		m.memory[addr+i] = byte(m.registers[inst.rs1] >> (i * 8))
	}
}

func (m *Machine) ldd(inst Instruction) {
	addr, ok := m.translate(uint32(inst.immediate), 3)
	if !ok {
		return
	}

	m.registers[inst.rd] = 0

	for i := uint32(0); i < 4; i++ {
		m.registers[inst.rd] |= (uint32(m.memory[addr+i]) << (i * 8))
	}
}

func (m *Machine) strd(inst Instruction) {
	addr, ok := m.translate(uint32(inst.immediate), 3)
	if !ok {
		return
	}

	for i := uint32(0); i < 4; i++ {
		m.memory[addr+i] = byte(m.registers[inst.rd] >> (i * 8))
	}
}

func (m *Machine) ldb(inst Instruction) {
	addr, ok := m.translate(m.registers[inst.rs2], 0)
	if !ok {
		return
	}

	m.registers[inst.rs1] = uint32(m.memory[addr])
}

func (m *Machine) ldsb(inst Instruction) {
	addr, ok := m.translate(m.registers[inst.rs2], 0)
	if !ok {
		return
	}

	m.registers[inst.rs1] = uint32(int8(m.memory[addr]))
}

func (m *Machine) strb(inst Instruction) {
	addr, ok := m.translate(m.registers[inst.rs2], 0)
	if !ok {
		return
	}

	m.memory[addr] = byte(m.registers[inst.rs1])
}

func (m *Machine) mret(inst Instruction) {
	m.registers[sr] = m.registers[esr]

	m.registers[pc] = m.registers[epc]
}

func (m *Machine) syscall(inst Instruction) {
	m.registers[w9] = uint32(inst.immediate)
	m.exception(syscallExc)
}

func (m *Machine) translate(addr uint32, addrOffset uint32) (uint32, bool) {
	if !m.isKernelMode() {
		addr = addr + m.registers[base]
		if addr+addrOffset >= m.registers[limit] {
			m.exception(faultExc)
			return addr, false
		}
	}

	return addr, true
}

func (m *Machine) exception(cause uint32) {
	m.registers[ecr] = cause

	m.registers[esr] = m.registers[sr]
	m.registers[sr] = 0

	m.registers[epc] = m.registers[pc]
	m.registers[pc] = m.registers[esa]
}

func (m *Machine) checkIllegalRegisterAccess(inst Instruction) bool {
	priviligedRegisters := []Register{ecr, esa, esr, epc, base, limit}
	return !m.isKernelMode() && (slices.Contains(priviligedRegisters, inst.rd) || slices.Contains(priviligedRegisters, inst.rs1) || slices.Contains(priviligedRegisters, inst.rs2))
}

func (m *Machine) checkIllegalInstruction(inst Instruction) bool {
	priviligedInstructions := []Opcode{MRET}
	return !m.isKernelMode() && slices.Contains(priviligedInstructions, inst.opcode)
}

func (m *Machine) isKernelMode() bool {
	return m.registers[sr]&8 == 0
}

func (m *Machine) isInterruptEnabled() bool {
	return m.registers[sr]&16 != 0
}

func (m *Machine) fetch() bool {
	addr, ok := m.translate(m.registers[pc], 3)
	if !ok {
		return false
	}

	var completeInstruction uint32 = 0

	for i := 0; i < 4; i++ {
		currentInstruction := m.memory[addr]
		completeInstruction = completeInstruction | uint32(currentInstruction)<<(24-8*i)
		addr += 1
		m.registers[pc] += 1
	}
	m.registers[ir] = completeInstruction
	return true
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

func (m *Machine) getOpcode(instruction uint32) Opcode {
	offsetOpcode := word - opcodeLength
	opcode := instruction >> (uint32(offsetOpcode))

	return Opcode(opcode)
}

func (m *Machine) decode() (Instruction, bool) {
	instruction := m.registers[ir]
	opcode := m.getOpcode(instruction)

	switch opcode {
	case AND, OR, XOR, ADD, SUB, MUL, UDIV, SDIV, CMP, JMPR, LOAD, STORE, LDB, LDSB, STRB:
		return m.decodeRTypeInst(instruction), true
	case MV, JUMP, BEQ, BLT, BGT, LDD, STRD, SYSCALL, MRET:
		return m.decodeITypeInst(instruction), true
	default:
		m.exception(faultExc)
		return Instruction{}, false
	}
}

func (m *Machine) execute(inst Instruction) {
	operations[inst.opcode](m, inst)
}

func (m *Machine) Boot() {
	m.registers[limit] = uint32(len(m.memory))
	instructionsExcecuted := 0

	var decodedInstruction Instruction
	var ok bool

	for {

		if m.debugFlag {
			select {
			case <-m.StepChan:
			case <-m.quitChan:
				return
			}
		}

		if !m.fetch() {
			goto endStep
		}

		decodedInstruction, ok = m.decode()
		if !ok {
			goto endStep
		}

		m.mu.Lock()
		if m.checkIllegalRegisterAccess(decodedInstruction) {
			m.exception(faultExc)
			m.mu.Unlock()
			goto endStep
		}
		if m.checkIllegalInstruction(decodedInstruction) {
			m.exception(faultExc)
			m.mu.Unlock()
			goto endStep
		}

		m.execute(decodedInstruction)

		if m.isInterruptEnabled() {
			instructionsExcecuted++

			if instructionsExcecuted%TIMER_INTERVAL == 0 {
				m.exception(clockInt)
			} else if m.killFlag {
				m.exception(killInt)
				m.killFlag = false
			} else if m.inputFlag {
				m.exception(inputInt)
				m.inputFlag = false
			}
		}
		m.mu.Unlock()

	endStep:
		if m.debugFlag {
			m.DoneChan <- struct{}{}
		}

	}
}

func (m *Machine) SaveCheckpoint() {
	cp := &Machine{
		memory:    make([]byte, len(m.memory)),
		registers: make(map[Register]uint32, len(m.registers)),
		debugFlag: m.debugFlag,
	}
	copy(cp.memory, m.memory)
	for k, v := range m.registers {
		cp.registers[k] = v
	}
	m.checkpoint = cp
}

func (m *Machine) RestoreCheckpoint() bool {
	if m.checkpoint == nil {
		return false
	}
	m.memory, m.checkpoint.memory = m.checkpoint.memory, m.memory
	m.registers, m.checkpoint.registers = m.checkpoint.registers, m.registers
	return true
}

func (m *Machine) LoadProgram(program []byte) {
	copy(m.memory, program)
}

func (m *Machine) GetMemory() []byte {
	return m.memory
}

func (m *Machine) ReadWord(addr uint32) uint32 {
	var word uint32
	for i := uint32(0); i < 4; i++ {
		word |= uint32(m.memory[addr+i]) << (24 - 8*i)
	}
	return word
}

func (m *Machine) LoadBuffer(buffer []byte) bool {
	if len(buffer) > bufferSize-4 {
		return false
	}

	bufferIndex := len(m.memory) - bufferSize
	clear(m.memory[bufferIndex:])

	for i := 0; i < 4; i++ {
		m.memory[bufferIndex+i] = byte(len(buffer) >> (i * 8))
	}

	copy(m.memory[bufferIndex+4:], buffer)

	return true
}

func (m *Machine) SetKillFlag() {
	m.killFlag = true
}

func (m *Machine) SetInputFlag() {
	m.inputFlag = true
}

func (m *Machine) IsDebugMode() bool {
	return m.debugFlag
}

func (m *Machine) Quit() {
	close(m.quitChan)
}

func (m *Machine) GetRegisters() map[Register]uint32 {

	return m.registers
}

func NewMachine(memoryBytes int, debugFlag bool) *Machine {
	machine := &Machine{
		memory:    make([]byte, memoryBytes),
		registers: make(map[Register]uint32),
		debugFlag: debugFlag,
		StepChan:  make(chan struct{}),
		DoneChan:  make(chan struct{}),
		quitChan:  make(chan struct{}),
	}

	return machine
}
