package assembler

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
	// Define registers
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

	// Define opcodes for different instructions
	MV_OPCODE Opcode = iota
	ADD_OPCODE
	SUB_OPCODE
	CMP_OPCODE
	JUMP_OPCODE
	LOAD_OPCODE
	STORE_OPCODE
)

// Map register names to Register values
var registerMap = map[string]Register{
	"W0": W0,
	"W1": W1,
	"W2": W2,
	"W3": W3,
	"W4": W4,
	"W5": W5,
}

// Map instruction names to Opcode values
var opcodeMap = map[string]Opcode{
	"ADD":   ADD_OPCODE,
	"SUB":   SUB_OPCODE,
	"CMP":   CMP_OPCODE,
	"MV":    MV_OPCODE,
	"JUMP":  JUMP_OPCODE,
	"LOAD":  LOAD_OPCODE,
	"STORE": STORE_OPCODE,
}

// Define instruction format and function codes for each type
type InstructionSpec struct {
	Format string
	Opcode Opcode
	Funct3 byte
	Funct7 byte
}

// Helper function to create a new InstructionSpec
func newInstructionSpec(format string, opcode Opcode) InstructionSpec {
	return InstructionSpec{
		Format: format,
		Opcode: opcode,
		Funct3: 0b00000,  // Default value
		Funct7: 0b000000, // Default value
	}
}

// Define the specifications for different instructions (R-Type and I-Type)
var instructionSpecs = map[Opcode]InstructionSpec{
	ADD_OPCODE:   newInstructionSpec("R-Type", ADD_OPCODE),
	SUB_OPCODE:   newInstructionSpec("R-Type", SUB_OPCODE),
	CMP_OPCODE:   newInstructionSpec("R-Type", CMP_OPCODE),
	MV_OPCODE:    newInstructionSpec("I-Type", MV_OPCODE),
	JUMP_OPCODE:  newInstructionSpec("I-Type", JUMP_OPCODE),
	LOAD_OPCODE:  newInstructionSpec("I-Type", LOAD_OPCODE),
	STORE_OPCODE: newInstructionSpec("I-Type", STORE_OPCODE),
}

// Common fields used across all instruction types
const (
	OPCODE_SHIFT = 26 // Shift opcode by 26 bits for all instructions
	OPCODE_MASK  = 0x3F
)

// R-Type instruction constants (shift and mask values)
const (
	// Destination register (rd)
	RD_SHIFT_R = 21 // Shift rd by 21 bits for R-Type
	RD_MASK_R  = 0x1F

	// Source register 1 (rs1)
	RS1_SHIFT_R = 16 // Shift rs1 by 16 bits for R-Type
	RS1_MASK_R  = 0x1F

	// Source register 2 (rs2)
	RS2_SHIFT_R = 11 // Shift rs2 by 11 bits for R-Type
	RS2_MASK_R  = 0x1F

	// Function codes (funct3 and funct7)
	FUNCT3_SHIFT_R = 6 // Shift funct3 by 6 bits for R-Type
	FUNCT3_MASK_R  = 0x1F
	FUNCT7_SHIFT_R = 0 // Shift funct7 by 0 bits for R-Type
	FUNCT7_MASK_R  = 0x3F
)

// I-Type instruction constants (shift and mask values)
const (
	// Destination/source register (rd/rs1)
	RD_RS1_SHIFT_I = 21 // Shift rd/rs1 by 21 bits for I-Type
	RD_RS1_MASK_I  = 0x1F

	// Function code (funct3)
	FUNCT3_SHIFT_I = 6 // Shift funct3 by 6 bits for I-Type
	FUNCT3_MASK_I  = 0x1F

	// Immediate value
	IMM_SHIFT_I = 5 // Shift immediate value by 5 bits for I-Type
	IMM_MASK_I  = 0xFFFF
)

// RunAssembler takes a file path to an assembly file, processes it, and returns the binary instructions as bytes
func RunAssembler(filePath string) ([]byte, error) {
	// Load assembly instructions from the file
	instrs, err := LoadAssemblyFile(filePath)
	if err != nil {
		return nil, err
	}

	// Convert instructions to binary format
	return ConvertInstructionsToBinary(instrs)
}

// Takes parsed assembly instructions and converts them to binary
func ConvertInstructionsToBinary(instructions [][]string) ([]byte, error) {
	var memory []byte

	// Loop through each instruction and convert it to binary
	for _, instr := range instructions {
		bytes, err := ConvertInstructionToBinary(instr)
		if err != nil {
			return nil, fmt.Errorf("Error converting instruction '%v': %v", instr, err)
		}

		memory = append(memory, bytes...)
	}
	return memory, nil
}

// Reads assembly code from a file and returns a slice of instructions (split into components)
func LoadAssemblyFile(filePath string) ([][]string, error) {
	file, err := os.Open(filePath) // Open the file
	if err != nil {
		return nil, fmt.Errorf("Error opening file '%s': %v", filePath, err)
	}
	defer file.Close()

	// Parse the assembly code from the file
	return LoadAssemblyFromReader(file)
}

// LoadAssemblyFromReader reads and processes lines of assembly code from a reader
func LoadAssemblyFromReader(reader io.Reader) ([][]string, error) {
	var instructions [][]string
	scanner := bufio.NewScanner(reader)

	// Read each line
	for scanner.Scan() {
		line := processLine(scanner.Text())
		if line != "" {
			// Split the line into components (opcode, registers)
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

// Processes an individual line of assembly code, removing comments and trimming whitespace
func processLine(line string) string {
	line = strings.TrimSpace(line)
	// Remove comments
	if idx := strings.Index(line, ";"); idx != -1 {
		line = strings.TrimSpace(line[:idx])
	}
	// Ignore labels
	if strings.HasSuffix(line, ":") {
		return ""
	}

	return strings.ReplaceAll(line, ",", "")
}

// Takes a single parsed instruction and converts it to binary
func ConvertInstructionToBinary(instruction []string) ([]byte, error) {
	if len(instruction) == 0 {
		return nil, fmt.Errorf("Empty instruction")
	}

	// Retrieve the opcode for the instruction
	opcodeStr := strings.ToUpper(instruction[0])
	opcode, ok := opcodeMap[opcodeStr]
	if !ok {
		return nil, fmt.Errorf("Invalid opcode: %s", opcodeStr)
	}

	// Get the instruction specification (format, funct3, funct7)
	spec, exists := instructionSpecs[opcode]
	if !exists {
		return nil, fmt.Errorf("No specification available for opcode: %s", opcodeStr)
	}

	var binaryInstruction uint32
	var err error

	// Encode the instruction based on its format (R-Type or I-Type)
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

	// Convert the 32-bit instruction to 4 bytes
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

	// Expect either 2 or 3 operands for R-Type instructions
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

	// Encode the R-Type instruction by combining the opcode, registers, and function fields
	binaryInstruction := uint32(spec.Opcode&OPCODE_MASK) << OPCODE_SHIFT
	binaryInstruction |= uint32(rd&RD_MASK_R) << RD_SHIFT_R
	binaryInstruction |= uint32(rs1&RS1_MASK_R) << RS1_SHIFT_R
	binaryInstruction |= uint32(rs2&RS2_MASK_R) << RS2_SHIFT_R
	binaryInstruction |= uint32(spec.Funct3&FUNCT3_MASK_R) << FUNCT3_SHIFT_R
	binaryInstruction |= uint32(spec.Funct7&FUNCT7_MASK_R) << FUNCT7_SHIFT_R

	return binaryInstruction, nil
}

func encodeIType(spec InstructionSpec, operands []string) (uint32, error) {
	var rd_rs1 byte
	var immediate uint16
	var err error

	// Expect either 1 or 2 operands for I-Type instructions
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

	// Encode the I-Type instruction by combining the opcode, register, and immediate value
	binaryInstruction := uint32(spec.Opcode&OPCODE_MASK) << OPCODE_SHIFT
	binaryInstruction |= uint32(rd_rs1&RD_RS1_MASK_I) << RD_RS1_SHIFT_I
	binaryInstruction |= uint32(spec.Funct3&FUNCT3_MASK_I) << FUNCT3_SHIFT_I
	binaryInstruction |= uint32(immediate&IMM_MASK_I) << IMM_SHIFT_I

	return binaryInstruction, nil
}

// parses a register name and returns its corresponding byte value
func parseRegister(register string) (byte, error) {
	reg, ok := registerMap[strings.ToUpper(register)]
	if !ok {
		return 0, fmt.Errorf("Invalid register: %s", register)
	}
	return byte(reg), nil
}

// parses an immediate value and returns its corresponding uint16 value
func parseImmediate(immediate string) (uint16, error) {
	immediate = strings.TrimPrefix(immediate, "#")

	var uintValue uint64
	var err error

	// Handle hexadecimal and decimal immediate values
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
