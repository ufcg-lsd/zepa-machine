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

	numberOfRegisters
)

type Machine struct {
	memory    []byte
	registers [numberOfRegisters]Register
}

func NewMachine(memoryBytes int) *Machine {
	machine := &Machine{
		memory: make([]byte, memoryBytes),
	}

	return machine
}
