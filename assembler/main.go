package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Register uint8
type Opcode uint8

const (
	W0 Register = iota
	W1
	W2
	W3
	W4
	W5
	_
	_
	_
	_
	_
	_

	MV_OPCODE Opcode = iota
	ADD_OPCODE
	SUB_OPCODE
	CMP_OPCODE
	JUMP_OPCODE
	LOAD_OPCODE
	STORE_OPCODE
)

var registerMap = map[string]Register{
	"W0": W0,
	"W1": W1,
	"W2": W2,
	"W3": W3,
	"W4": W4,
	"W5": W5,
}

var opcodeMap = map[string]Opcode{
	"ADD":   ADD_OPCODE,
	"SUB":   SUB_OPCODE,
	"CMP":   CMP_OPCODE,
	"MV":    MV_OPCODE,
	"JUMP":  JUMP_OPCODE,
	"LOAD":  LOAD_OPCODE,
	"STORE": STORE_OPCODE,
}

type InstructionSpec struct {
	Format string
	Opcode Opcode
	Funct3 byte
	Funct7 byte
}

func newInstructionSpec(format string, opcode Opcode) InstructionSpec {
	return InstructionSpec{
		Format: format,
		Opcode: opcode,
		Funct3: 0b00000,
		Funct7: 0b000000,
	}
}

var instructionSpecs = map[Opcode]InstructionSpec{
	ADD_OPCODE:   newInstructionSpec("R-Type", ADD_OPCODE),
	SUB_OPCODE:   newInstructionSpec("R-Type", SUB_OPCODE),
	CMP_OPCODE:   newInstructionSpec("R-Type", CMP_OPCODE),
	MV_OPCODE:    newInstructionSpec("I-Type", MV_OPCODE),
	JUMP_OPCODE:  newInstructionSpec("I-Type", JUMP_OPCODE),
	LOAD_OPCODE:  newInstructionSpec("I-Type", LOAD_OPCODE),
	STORE_OPCODE: newInstructionSpec("I-Type", STORE_OPCODE),
}

const (
	// Common Instruction Fields
	OPCODE_SHIFT = 26
	OPCODE_MASK  = 0x3F
)

const (
	// R-Type Instructions
	RD_SHIFT       = 21
	RD_MASK        = 0x1F
	RS1_SHIFT      = 16
	RS1_MASK       = 0x1F
	RS2_SHIFT      = 11
	RS2_MASK       = 0x1F
	FUNCT3_SHIFT_R = 6
	FUNCT3_MASK_R  = 0x1F
	FUNCT7_SHIFT   = 0
	FUNCT7_MASK    = 0x3F
)

const (
	// R-Type Instructions
	RD_RS1_SHIFT_I = 21
	RD_RS1_MASK_I  = 0x1F
	FUNCT3_SHIFT_I = 6
	FUNCT3_MASK_I  = 0x1F
	IMM_SHIFT_I    = 5
	IMM_MASK_I     = 0xFFFF
)

var memory []byte

func main() {
	if len(os.Args) < 2 {
		return
	}

	sourceFile := os.Args[1]
	if err := RunAssembler(sourceFile); err != nil {
		return
	}
}

func RunAssembler(filePath string) error {
	instrs, err := LoadAssemblyFile(filePath)
	if err != nil {
		return err
	}

	return ConvertInstructionsToBinary(instrs)
}

func ConvertInstructionsToBinary(instructions [][]string) error {
	for _, instr := range instructions {
		bytes, err := ConvertInstructionToBinary(instr)
		if err != nil {
			return fmt.Errorf("Error converting instruction '%v': %v", instr, err)
		}
		memory = append(memory, bytes...)
	}
	return nil
}

func LoadAssemblyFile(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error opening file '%s': %v", filePath, err)
	}
	defer file.Close()

	return LoadAssemblyFromReader(file)
}

func LoadAssemblyFromReader(reader io.Reader) ([][]string, error) {
	var instructions [][]string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := processLine(scanner.Text())
		if line != "" {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				instructions = append(instructions, parts)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error reading assembly code: %v", err)
	}

	return instructions, nil
}

func processLine(line string) string {
	line = strings.TrimSpace(line)
	if idx := strings.Index(line, ";"); idx != -1 {
		line = strings.TrimSpace(line[:idx])
	}
	if strings.HasSuffix(line, ":") {
		return ""
	}
	return strings.ReplaceAll(line, ",", "")
}

func ConvertInstructionToBinary(instruction []string) ([]byte, error) {
	if len(instruction) == 0 {
		return nil, fmt.Errorf("Empty instruction")
	}

	opcodeStr := strings.ToUpper(instruction[0])
	opcode, ok := opcodeMap[opcodeStr]
	if !ok {
		return nil, fmt.Errorf("Invalid opcode: %s", opcodeStr)
	}

	spec, exists := instructionSpecs[opcode]
	if !exists {
		return nil, fmt.Errorf("No specification available for opcode: %s", opcodeStr)
	}

	var binaryInstruction uint32
	var err error

	switch spec.Format {
	case "R-Type":
		binaryInstruction, err = encodeRType(spec, instruction[1:])
	case "I-Type":
		binaryInstruction, err = encodeIType(spec, instruction[1:])
	default:
		return nil, fmt.Errorf("Unknown instruction format: %s", spec.Format)
	}

	if err != nil {
		return nil, fmt.Errorf("Error encoding instruction '%v': %v", instruction, err)
	}

	bytes := []byte{
		byte((binaryInstruction >> 24) & 0xFF),
		byte((binaryInstruction >> 16) & 0xFF),
		byte((binaryInstruction >> 8) & 0xFF),
		byte(binaryInstruction & 0xFF),
	}

	return bytes, nil
}

func encodeRType(spec InstructionSpec, operands []string) (uint32, error) {
	var rd, rs1, rs2 byte
	var err error

	switch len(operands) {
	case 3:
		rd, err = parseRegister(operands[0])
		if err != nil {
			return 0, err
		}
		rs1, err = parseRegister(operands[1])
		if err != nil {
			return 0, err
		}
		rs2, err = parseRegister(operands[2])
		if err != nil {
			return 0, err
		}
	case 2:
		rd = 0
		rs1, err = parseRegister(operands[0])
		if err != nil {
			return 0, err
		}
		rs2, err = parseRegister(operands[1])
		if err != nil {
			return 0, err
		}
	default:
		return 0, fmt.Errorf("R-Type instruction expects 2 or 3 operands, got %d", len(operands))
	}

	binaryInstruction := uint32(spec.Opcode&OPCODE_MASK) << OPCODE_SHIFT
	binaryInstruction |= uint32(rd&RD_MASK) << RD_SHIFT
	binaryInstruction |= uint32(rs1&RS1_MASK) << RS1_SHIFT
	binaryInstruction |= uint32(rs2&RS2_MASK) << RS2_SHIFT
	binaryInstruction |= uint32(spec.Funct3&FUNCT3_MASK_R) << FUNCT3_SHIFT_R
	binaryInstruction |= uint32(spec.Funct7&FUNCT7_MASK) << FUNCT7_SHIFT

	return binaryInstruction, nil
}

func encodeIType(spec InstructionSpec, operands []string) (uint32, error) {
	var rd_rs1 byte
	var immediate uint16
	var err error

	switch len(operands) {
	case 2:
		rd_rs1, err = parseRegister(operands[0])
		if err != nil {
			return 0, err
		}
		immediate, err = parseImmediate(operands[1])
		if err != nil {
			return 0, err
		}
	case 1:
		rd_rs1 = 0
		immediate, err = parseImmediate(operands[0])
		if err != nil {
			return 0, err
		}
	default:
		return 0, fmt.Errorf("I-Type instruction expects 1 or 2 operands, got %d", len(operands))
	}

	binaryInstruction := uint32(spec.Opcode&OPCODE_MASK) << OPCODE_SHIFT
	binaryInstruction |= uint32(rd_rs1&RD_RS1_MASK_I) << RD_RS1_SHIFT_I
	binaryInstruction |= uint32(spec.Funct3&FUNCT3_MASK_I) << FUNCT3_SHIFT_I
	binaryInstruction |= uint32(immediate&IMM_MASK_I) << IMM_SHIFT_I

	return binaryInstruction, nil
}

func parseRegister(register string) (byte, error) {
	reg, ok := registerMap[strings.ToUpper(register)]
	if !ok {
		return 0, fmt.Errorf("Invalid register: %s", register)
	}
	return byte(reg), nil
}

func parseImmediate(immediate string) (uint16, error) {
	immediate = strings.TrimPrefix(immediate, "#")

	var uintValue uint64
	var err error

	if strings.HasPrefix(immediate, "0x") || strings.HasPrefix(immediate, "0X") {
		uintValue, err = strconv.ParseUint(immediate, 0, 16)
	} else {
		uintValue, err = strconv.ParseUint(immediate, 10, 16)
	}

	if err != nil {
		return 0, fmt.Errorf("Invalid immediate value: %s", immediate)
	}

	return uint16(uintValue), nil
}
