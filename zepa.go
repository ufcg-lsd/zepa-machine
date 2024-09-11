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

type Machine struct {
	memory    []byte
	registers map[Register]byte
}

func (m *Machine) fetch() {
	m.registers[ir] = m.memory[m.registers[pc]]
	m.registers[pc] += 1
}

func (m *Machine) decode() {}

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
