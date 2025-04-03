package core

type Opcode uint8
type Register uint8
type Exception uint8
type Interrupt uint8

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
	LR
	SSR

	// Define opcodes for different instructions
	MV_OPCODE Opcode = iota
	ADD_OPCODE
	SUB_OPCODE
	CMP_OPCODE
	JUMP_OPCODE
	LOAD_OPCODE
	STORE_OPCODE
	FETCH_OPCODE
	HALT_OPCODE
	RET_OPCODE
	BEQ_OPCODE
	BLT_OPCODE
	BGT_OPCODE
	UDF_OPCODE
)

const (
	OpCodeBitMask    = 0x3F
	RegisterBitMask  = 0x1F
	ImmediateBitMask = 0xFFFF
	Funct5BitMask    = 0x1F
	Funct6BitMask    = 0x3F
)

// Exception codes
const (
	EXC_DEFAULT Exception = iota
	EXC_MEMORY_VIOLATION

	INT_TIMER Interrupt = iota
)
