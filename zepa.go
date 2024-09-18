package machine

type Register byte

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
)

const (
	opcodeLength = 6
)

type Instruction struct {
	opcode    byte     // 6 bits
	rd        Register // 5 bits
	rs1       Register // 5 bits
	rs2       Register // 5 bits
	funct3    byte     // 5 bits
	funct7    byte     // 6 bits
	addrConst uint16   // 16 bits
}

type Machine struct {
	memory    []byte
	registers map[Register]byte
}

func (m *Machine) fetch() {
	currentInstructionAddress := m.registers[pc]
	currentInstruction := m.memory[currentInstructionAddress]
	// Update Instruction Register with current instruction
	m.registers[ir] = currentInstruction
	// Increment Program Counter
	m.registers[pc] += 1
}

func (m *Machine) decodeRTypeInst() Instruction {

}

func (m *Machine) decodeITypeInst() Instruction {

}

func (m *Machine) decode() Instruction {

}

func (m *Machine) execute() {}

func (m *Machine) boot() {
	for {
		m.fetch()
		m.decode()
		m.execute()
	}
}

func NewMachine(memoryBytes int) *Machine {
	machine := &Machine{
		memory:    make([]byte, memoryBytes),
		registers: make(map[Register]byte),
	}

	return machine
}
