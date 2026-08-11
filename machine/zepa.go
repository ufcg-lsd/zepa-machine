package machine

import (
	"encoding/binary"
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
	uptr
	kptr
	efa
)

const (
	MV Opcode = iota
	AND
	OR
	XOR
	SHL
	SHA
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
	pageFaultExc
)

const TIMER_INTERVAL = 128

const (
	pageSize        = 4096       // 4KB pages (12 bits of offset)
	kernelBoundary  = 0xC0000000 // 3GB mark
	isPteMappedMask = 0x00100000 // V flag is bit 20 of the PTE
)

var operations = map[Opcode]Operation{
	MV:      (*Machine).mv,
	AND:     (*Machine).and,
	OR:      (*Machine).or,
	XOR:     (*Machine).xor,
	SHL:     (*Machine).shl,
	SHA:     (*Machine).sha,
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

func (m *Machine) shl(inst Instruction) {
	val := m.registers[inst.rs1]
	shiftAmount := int32(m.registers[inst.rs2])

	if shiftAmount > 0 {
		m.registers[inst.rd] = val << shiftAmount
	} else if shiftAmount < 0 {
		m.registers[inst.rd] = val >> (-shiftAmount)
	} else {
		m.registers[inst.rd] = val
	}
}

func (m *Machine) sha(inst Instruction) {
	val := m.registers[inst.rs1]
	shiftAmount := int32(m.registers[inst.rs2])

	if shiftAmount > 0 {
		m.registers[inst.rd] = val << shiftAmount
	} else if shiftAmount < 0 {
		m.registers[inst.rd] = uint32(int32(val) >> (-shiftAmount))
	} else {
		m.registers[inst.rd] = val
	}
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
	addr, ok := m.translate(m.registers[inst.rs2], 4)
	if !ok {
		return
	}

	m.registers[inst.rs1] = binary.LittleEndian.Uint32(m.memory[addr : addr+4])
}

func (m *Machine) store(inst Instruction) {
	addr, ok := m.translate(m.registers[inst.rs2], 4)
	if !ok {
		return
	}

	binary.LittleEndian.PutUint32(m.memory[addr:addr+4], m.registers[inst.rs1])
}

func (m *Machine) ldd(inst Instruction) {
	addr, ok := m.translate(uint32(inst.immediate), 4)
	if !ok {
		return
	}

	m.registers[inst.rd] = binary.LittleEndian.Uint32(m.memory[addr : addr+4])
}

func (m *Machine) strd(inst Instruction) {
	addr, ok := m.translate(uint32(inst.immediate), 4)
	if !ok {
		return
	}

	binary.LittleEndian.PutUint32(m.memory[addr:addr+4], m.registers[inst.rd])
}

func (m *Machine) ldb(inst Instruction) {
	addr, ok := m.translate(m.registers[inst.rs2], 1)
	if !ok {
		return
	}

	m.registers[inst.rs1] = uint32(m.memory[addr])
}

func (m *Machine) ldsb(inst Instruction) {
	addr, ok := m.translate(m.registers[inst.rs2], 1)
	if !ok {
		return
	}

	m.registers[inst.rs1] = uint32(int8(m.memory[addr]))
}

func (m *Machine) strb(inst Instruction) {
	addr, ok := m.translate(m.registers[inst.rs2], 1)
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

func (m *Machine) translate(addr uint32, byteCount uint32) (uint32, bool) {
	if !m.isMmuEnabled() {
		return addr, true
	}

	if byteCount == 4 && addr%4 != 0 {
		m.registers[efa] = addr
		m.exception(faultExc) // unaligned address
		return 0, false
	}

	if !m.isKernelMode() && addr >= kernelBoundary {
		m.exception(pageFaultExc)
		m.registers[efa] = addr
		return 0, false
	}

	var ptr uint32
	if addr >= kernelBoundary {
		ptr = uint32(m.registers[kptr])
	} else {
		ptr = uint32(m.registers[uptr])
	}

	pageNumber := addr / pageSize
	pteAddr := ptr + (pageNumber * 4)

	pte := binary.LittleEndian.Uint32(m.memory[pteAddr : pteAddr+4])

	if pte&isPteMappedMask == 0 {
		m.registers[efa] = addr
		m.exception(pageFaultExc)
		return 0, false
	}

	physicalFrame := pte & 0xFFFFF
	offset := addr % pageSize
	physicalAddr := physicalFrame + offset

	return physicalAddr, true
}

func (m *Machine) exception(cause uint32) {
	m.registers[ecr] = cause

	m.registers[esr] = m.registers[sr]
	m.registers[sr] &^= 0b11111 //sets the first 5 bits to 0, clears cmp tags, enters kernel mode and disables instructions

	m.registers[epc] = m.registers[pc]
	m.registers[pc] = m.registers[esa]
}

func (m *Machine) checkIllegalRegisterAccess(inst Instruction) bool {
	priviligedRegisters := []Register{ecr, esa, esr, epc, kptr, uptr, efa}
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

func (m *Machine) isMmuEnabled() bool {
	return m.registers[sr]&32 != 0
}

func (m *Machine) fetch() bool {
	addr, ok := m.translate(m.registers[pc], 4)
	if !ok {
		return false
	}

	m.registers[ir] = binary.BigEndian.Uint32(m.memory[addr : addr+4])
	m.registers[pc] += 4
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
	case AND, OR, XOR, ADD, SUB, MUL, UDIV, SDIV, CMP, JMPR, LOAD, STORE, LDB, LDSB, STRB, SHL, SHA:
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
	m.registers[w0] = uint32(len(m.memory))
	instructionsExcecuted := 0

	var decodedInstruction Instruction
	var ok bool

	for {

		if m.debugFlag {
			<-m.StepChan
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

	binary.LittleEndian.PutUint32(m.memory[bufferIndex:bufferIndex+4], uint32(len(buffer)))
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
	}

	return machine
}
