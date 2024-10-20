package assembler

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"zepa-machine/core"
)

// Map register names to Register values
var registerMap = map[string]core.Register{
	"W0": core.W0,
	"W1": core.W1,
	"W2": core.W2,
	"W3": core.W3,
	"W4": core.W4,
	"W5": core.W5,
}

// Map instruction names to Opcode values
var opcodeMap = map[string]core.Opcode{
	"ADD":   core.ADD_OPCODE,
	"SUB":   core.SUB_OPCODE,
	"CMP":   core.CMP_OPCODE,
	"MV":    core.MV_OPCODE,
	"JUMP":  core.JUMP_OPCODE,
	"LOAD":  core.LOAD_OPCODE,
	"STORE": core.STORE_OPCODE,
}

// Define instruction format and function codes for each type
type InstructionSpec struct {
	Format string
	Opcode core.Opcode
	Funct5 byte
	Funct6 byte
}

// Helper function to create a new InstructionSpec
func newInstructionSpec(format string, opcode core.Opcode) InstructionSpec {
	return InstructionSpec{
		Format: format,
		Opcode: opcode,
		Funct5: 0b00000,  // Default value
		Funct6: 0b000000, // Default value
	}
}

// Define the specifications for different instructions (R-Type and I-Type)
var instructionSpecs = map[core.Opcode]InstructionSpec{
	core.ADD_OPCODE:   newInstructionSpec("R-Type", core.ADD_OPCODE),
	core.SUB_OPCODE:   newInstructionSpec("R-Type", core.SUB_OPCODE),
	core.CMP_OPCODE:   newInstructionSpec("R-Type", core.CMP_OPCODE),
	core.MV_OPCODE:    newInstructionSpec("I-Type", core.MV_OPCODE),
	core.JUMP_OPCODE:  newInstructionSpec("I-Type", core.JUMP_OPCODE),
	core.LOAD_OPCODE:  newInstructionSpec("I-Type", core.LOAD_OPCODE),
	core.STORE_OPCODE: newInstructionSpec("I-Type", core.STORE_OPCODE),
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
	FUNCT5_SHIFT_R = 6 // Shift funct3 by 6 bits for R-Type
	FUNCT5_MASK_R  = 0x1F
	FUNCT6_SHIFT_R = 0 // Shift funct7 by 0 bits for R-Type
	FUNCT6_MASK_R  = 0x3F
)

// I-Type instruction constants (shift and mask values)
const (
	// Destination/source register (rd/rs1)
	RD_RS1_SHIFT_I = 21 // Shift rd/rs1 by 21 bits for I-Type
	RD_RS1_MASK_I  = 0x1F

	// Function code (funct3)
	FUNCT5_SHIFT_I = 6 // Shift funct3 by 6 bits for I-Type
	FUNCT5_MASK_I  = 0x1F

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
	binaryInstruction |= uint32(spec.Funct5&FUNCT5_MASK_R) << FUNCT5_SHIFT_R
	binaryInstruction |= uint32(spec.Funct6&FUNCT6_MASK_R) << FUNCT6_SHIFT_R

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
	binaryInstruction |= uint32(spec.Funct5&FUNCT5_MASK_I) << FUNCT5_SHIFT_I
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
