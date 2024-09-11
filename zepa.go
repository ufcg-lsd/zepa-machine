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

func NewMachine(memoryBytes int) *Machine {

	machine := &Machine{
		memory:    make([]byte, memoryBytes),
		registers: make(map[Register]byte),
	}

	return machine
}
