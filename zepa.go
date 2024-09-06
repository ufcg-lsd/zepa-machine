package machine

type GeneralRegisters struct {
	w0, w1, w2, w3, w4, w5 byte
}

type SpecialRegisters struct {
	pc  byte
	sp  byte
	ir  byte
	sr  byte
	mdr byte
	mar byte
}

type Memory struct {
	data [1024]byte
}

type Machine struct {
	genRegs  GeneralRegisters
	specRegs SpecialRegisters
	memory   Memory
}
