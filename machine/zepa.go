package machine

import (
	"fmt"
	"os"
	"time"
	"zepa-machine/core"
	assembler "zepa-machine/cross-assembler"
)

type Operation func(m *Machine, inst Instruction)

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

var operations = map[byte]Operation{
	byte(core.MV_OPCODE):    (*Machine).mv,
	byte(core.ADD_OPCODE):   (*Machine).add,
	byte(core.SUB_OPCODE):   (*Machine).sub,
	byte(core.CMP_OPCODE):   (*Machine).cmp,
	byte(core.JUMP_OPCODE):  (*Machine).jump,
	byte(core.LOAD_OPCODE):  (*Machine).load,
	byte(core.STORE_OPCODE): (*Machine).store,
	byte(core.HALT_OPCODE):  (*Machine).halt,
	byte(core.RET_OPCODE):   (*Machine).ret,
	byte(core.BEQ_OPCODE):   (*Machine).beq,
	byte(core.BLT_OPCODE):   (*Machine).blt,
	byte(core.BGT_OPCODE):   (*Machine).bgt,
	byte(core.UDF_OPCODE):   (*Machine).udf,
}

type Instruction struct {
	opcode    func(m *Machine, inst Instruction)
	rd        core.Register
	rs1       core.Register
	rs2       core.Register
	funct5    byte
	funct6    byte
	immediate uint16
}

type Machine struct {
	memory         []byte
	registers      map[core.Register]uint32
	evt            map[core.Exception]int
	ivt            map[core.Interrupt]func(m *Machine)
	interruptQueue []core.Interrupt
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
		m.registers[core.SR] = 0
	} else if m.registers[inst.rs1] > m.registers[inst.rs2] {
		m.registers[core.SR] = 2
	} else {
		m.registers[core.SR] = 1
	}
}

func (m *Machine) jump(inst Instruction) {
	m.registers[core.PC] = uint32(inst.immediate)
}

func (m *Machine) beq(inst Instruction) {
	if m.registers[core.SR] == 0 {
		m.registers[core.PC] = uint32(inst.immediate)
	}
}

func (m *Machine) blt(inst Instruction) {
	if m.registers[core.SR] == 1 {
		m.registers[core.PC] = uint32(inst.immediate)
	}
}

func (m *Machine) bgt(inst Instruction) {
	if m.registers[core.SR] == 2 {
		m.registers[core.PC] = uint32(inst.immediate)
	}
}

func (m *Machine) load(inst Instruction) {
	// Verify invalid address
	if int(inst.immediate) >= len(m.memory) {
		m.exception(core.EXC_MEMORY_VIOLATION)
		return
	}

	m.registers[inst.rd] = uint32(m.memory[inst.immediate])
}

func (m *Machine) store(inst Instruction) {
	// Verify invalid address
	if int(inst.immediate) >= len(m.memory) {
		m.exception(core.EXC_MEMORY_VIOLATION)
		return
	}

	m.memory[inst.immediate] = byte(m.registers[inst.rd])
}

func (m *Machine) fetch() {
	var completeInstruction uint32 = 0
	for i := 0; i < 4; i++ {
		currentInstructionAddress := m.registers[core.PC]
		currentInstruction := m.memory[currentInstructionAddress]
		completeInstruction = completeInstruction | uint32(currentInstruction)<<(24-8*i)
		m.registers[core.PC] += 1
	}
	m.registers[core.IR] = completeInstruction
}

func (m *Machine) decodeRTypeInst(instruction uint32) Instruction {
	offsetOpcode := word - opcodeLength
	offSetRd := offsetOpcode - rdLength
	offSetRs1 := offSetRd - rs1Length
	offSetRs2 := offSetRs1 - rs2Length
	offSetFunct5 := offSetRs2 - funct5Length
	offSetFunct6 := offSetFunct5 - funct6Length

	opcode := instruction >> (uint32(offsetOpcode)) & core.OpCodeBitMask
	rd := (instruction >> uint32(offSetRd)) & core.RegisterBitMask
	rs1 := (instruction >> (uint32(offSetRs1))) & core.RegisterBitMask
	rs2 := (instruction >> (uint32(offSetRs2))) & core.RegisterBitMask
	funct5 := (instruction >> (uint32(offSetFunct5))) & core.Funct5BitMask
	funct6 := (instruction >> (uint32(offSetFunct6))) & core.Funct6BitMask

	operation := operations[byte(opcode)]

	return Instruction{
		opcode: operation,
		rd:     core.Register(rd),
		rs1:    core.Register(rs1),
		rs2:    core.Register(rs2),
		funct5: byte(funct5),
		funct6: byte(funct6),
	}
}

func (m *Machine) decodeITypeInst(instruction uint32) Instruction {
	offsetOpcode := word - opcodeLength
	offSetRdRs1 := offsetOpcode - rdLength
	offSetImmediate := offSetRdRs1 - immediateLen
	offSetFunct5 := offSetImmediate - funct5Length

	opcode := instruction >> (uint32(offsetOpcode)) & core.OpCodeBitMask
	rdRs1 := (instruction >> uint32(offSetRdRs1)) & core.RegisterBitMask
	immediate := (instruction >> (uint32(offSetImmediate))) & core.ImmediateBitMask
	funct5 := (instruction >> (uint32(offSetFunct5))) & core.Funct5BitMask

	operation := operations[byte(opcode)]

	return Instruction{
		opcode:    operation,
		rd:        core.Register(rdRs1),
		immediate: uint16(immediate),
		funct5:    byte(funct5),
	}
}

func (m *Machine) isEndOfProgram() bool {
	if (m.registers[core.IR]) == 0 {
		m.registers[core.PC] -= 4
		return true
	}
	return false
}

func (m *Machine) getOpcode(instruction uint32) core.Opcode {
	offsetOpcode := word - opcodeLength
	opcode := instruction >> (uint32(offsetOpcode))

	return core.Opcode(opcode)
}

func (m *Machine) decode() Instruction {
	instruction := m.registers[core.IR]
	opcode := m.getOpcode(instruction)

	switch opcode {
	case core.ADD_OPCODE, core.SUB_OPCODE, core.CMP_OPCODE:
		return m.decodeRTypeInst(instruction)
	case core.MV_OPCODE, core.JUMP_OPCODE, core.LOAD_OPCODE, core.STORE_OPCODE,
		core.HALT_OPCODE, core.RET_OPCODE, core.BEQ_OPCODE, core.BGT_OPCODE, core.BLT_OPCODE, core.UDF_OPCODE:
		fallthrough
	default:
		return m.decodeITypeInst(instruction)
	}
}

func (m *Machine) execute(inst Instruction) {
	inst.opcode(m, inst)
}

func (m *Machine) Boot() {
	cycle := 0
	for {
		m.fetch()
		if m.isEndOfProgram() {
			break
		}
		decodedInstruction := m.decode()
		m.execute(decodedInstruction)

		cycle++

		if cycle >= 4 { // Timer interrupt example
			m.interrupt(core.INT_TIMER)
			cycle = 0
		}

		m.handleInterrupts()
	}
}

func (m *Machine) LoadProgram(program []byte) {
	copy(m.memory, program)
}

func (m *Machine) GetMemory() []byte {
	return m.memory
}

func (m *Machine) GetRegisters() map[core.Register]uint32 {
	return m.registers
}

func NewMachine(memoryBytes int) *Machine {
	// Define exception handler code
	handlerCode, err := assembler.ConvertInstructionsToBinary([][]string{
		{"HALT"}, //M + 0 : Default Handler
		{"RET"},  //M + 4 : Memory Violation Handler
	})

	if err != nil {
		fmt.Printf("%v\n", err)
	}

	// Setting space to exception handler and W registers backup
	qntRegisters := len(assembler.RegisterMap)
	exceptionHandlerSize := len(handlerCode) + qntRegisters
	handlerAddress := memoryBytes - exceptionHandlerSize // Set handler position in memory

	// Define the machine
	machine := &Machine{
		memory:         make([]byte, memoryBytes),
		registers:      make(map[core.Register]uint32),
		evt:            make(map[core.Exception]int),
		ivt:            make(map[core.Interrupt]func(m *Machine)),
		interruptQueue: []core.Interrupt{},
	}

	// Load exceptions
	machine.evt[core.EXC_DEFAULT] = handlerAddress              // Default handler location
	machine.evt[core.EXC_MEMORY_VIOLATION] = handlerAddress + 4 // Set memory violation handler location

	//Load interrupts
	machine.ivt[core.INT_TIMER] = TimerInterrupt // Set timer interrupt handler

	// Load exception handler to memory
	copy(machine.memory[handlerAddress:], handlerCode)

	return machine
}

func (m *Machine) exception(code core.Exception) {
	fmt.Printf("Exception raised: Code %d\n", code)

	// Save W registers into memory
	m.memory[len(m.memory)-6] = byte(m.registers[core.W0])
	m.memory[len(m.memory)-5] = byte(m.registers[core.W1])
	m.memory[len(m.memory)-4] = byte(m.registers[core.W2])
	m.memory[len(m.memory)-3] = byte(m.registers[core.W3])
	m.memory[len(m.memory)-2] = byte(m.registers[core.W4])
	m.memory[len(m.memory)-1] = byte(m.registers[core.W5])

	// Save information
	m.registers[core.LR] = m.registers[core.PC]  // Save instruction
	m.registers[core.SSR] = m.registers[core.SR] // Save Status

	// Redirect to exception Handler
	m.registers[core.PC] = uint32(m.evt[code])
}

func (m *Machine) halt(inst Instruction) {
	os.Exit(1)
}

func (m *Machine) ret(inst Instruction) {
	// Restore values
	m.registers[core.PC] = m.registers[core.LR]  // Restore PC (value before procedure) and go to next instruction
	m.registers[core.SSR] = m.registers[core.SR] // Restore status (value before procedure)

	// Restore W registers
	m.registers[core.W0] = uint32(m.memory[len(m.memory)-6])
	m.registers[core.W1] = uint32(m.memory[len(m.memory)-5])
	m.registers[core.W2] = uint32(m.memory[len(m.memory)-4])
	m.registers[core.W3] = uint32(m.memory[len(m.memory)-3])
	m.registers[core.W4] = uint32(m.memory[len(m.memory)-2])
	m.registers[core.W5] = uint32(m.memory[len(m.memory)-1])

	// Reset link register
	m.registers[core.LR] = 0
}

func (m *Machine) udf(inst Instruction) {
	m.exception(0)
}

func (m *Machine) interrupt(code core.Interrupt) {
	m.interruptQueue = append(m.interruptQueue, code)
}

func (m *Machine) handleInterrupts() {
	for len(m.interruptQueue) > 0 {
		interruptCode := m.interruptQueue[0]
		m.interruptQueue = m.interruptQueue[1:] // Remove from queue

		if handler, exists := m.ivt[interruptCode]; exists {
			handler(m)
		} else {
			fmt.Printf("Unhandled interrupt: Code %d\n", interruptCode)
		}
	}
}

func TimerInterrupt(m *Machine) {
	fmt.Printf("Timer Interrupt: 2s\n")
	time.Sleep(2 * time.Second)
}
