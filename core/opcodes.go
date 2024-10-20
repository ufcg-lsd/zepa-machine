package core

type Opcode uint8
type Register uint8

const (
	// Define registers
	W0 Register = iota
	W1
	W2
	W3
	W4
	W5
	PC
	SP
	IR
	SR
	MDR
	MAR

	// Define opcodes for different instructions
	MV_OPCODE Opcode = iota
	ADD_OPCODE
	SUB_OPCODE
	CMP_OPCODE
	JUMP_OPCODE
	LOAD_OPCODE
	STORE_OPCODE
)
