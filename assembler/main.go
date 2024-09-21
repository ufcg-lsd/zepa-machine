package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Register uint32
type Opcode uint8

const (
	W0 Register = iota
	W1
	W2
	W3
	W4
	W5
	// Reservado para futuros registradores
	_
	_
	_
	_
	_
)

const (
	// R-Type Opcodes
	ADD_OPCODE Opcode = 0b001101
	SUB_OPCODE Opcode = 0b001110
	CMP_OPCODE Opcode = 0b001111

	// I-Type Opcodes
	MV_OPCODE    Opcode = 0b001100
	JUMP_OPCODE  Opcode = 0b010000
	LOAD_OPCODE  Opcode = 0b010001
	STORE_OPCODE Opcode = 0b010010
	FETCH_OPCODE Opcode = 0b010011
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
	"FETCH": FETCH_OPCODE,
}

type InstructionSpec struct {
	Format string
	Opcode Opcode
	Funct3 byte
	Funct7 byte
}

var instructionSpecs = map[Opcode]InstructionSpec{
	ADD_OPCODE: {
		Format: "R-Type",
		Opcode: ADD_OPCODE,
		Funct3: 0b00000,
		Funct7: 0b000000,
	},
	SUB_OPCODE: {
		Format: "R-Type",
		Opcode: SUB_OPCODE,
		Funct3: 0b00000,
		Funct7: 0b000000,
	},
	CMP_OPCODE: {
		Format: "R-Type",
		Opcode: CMP_OPCODE,
		Funct3: 0b00000,
		Funct7: 0b000000,
	},
	MV_OPCODE: {
		Format: "I-Type",
		Opcode: MV_OPCODE,
		Funct3: 0b00000,
		Funct7: 0b000000, // Não usado em I-Type, mas mantido para consistência
	},
	JUMP_OPCODE: {
		Format: "I-Type",
		Opcode: JUMP_OPCODE,
		Funct3: 0b00000,
		Funct7: 0b000000,
	},
	LOAD_OPCODE: {
		Format: "I-Type",
		Opcode: LOAD_OPCODE,
		Funct3: 0b00000,
		Funct7: 0b000000,
	},
	STORE_OPCODE: {
		Format: "I-Type",
		Opcode: STORE_OPCODE,
		Funct3: 0b00000,
		Funct7: 0b000000,
	},
	FETCH_OPCODE: {
		Format: "I-Type",
		Opcode: FETCH_OPCODE,
		Funct3: 0b00000,
		Funct7: 0b000000,
	},
}

var encoderMap = map[Opcode]func(InstructionSpec, []string) (uint32, error){
	ADD_OPCODE:   encodeRType,
	SUB_OPCODE:   encodeRType,
	CMP_OPCODE:   encodeRType,
	MV_OPCODE:    encodeIType,
	LOAD_OPCODE:  encodeIType,
	STORE_OPCODE: encodeIType,
	JUMP_OPCODE:  encodeIType,
	FETCH_OPCODE: encodeIType,
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: assembler <input_file.asm>")
		return
	}

	sourceFile := os.Args[1]
	instructionMemory, err := RunAssembler(sourceFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for i, instruction := range instructionMemory {
		instructionBits := fmt.Sprintf("%032b", instruction)
		fmt.Printf("Instruction %d: %s\n", i, instructionBits)
	}
}

func RunAssembler(filePath string) ([]uint32, error) {
	instrs, err := LoadAssemblyFile(filePath)
	if err != nil {
		return nil, err
	}
	return ConvertInstructionsToBinary(instrs)
}

func ConvertInstructionsToBinary(instructions [][]string) ([]uint32, error) {
	var instructionMemory []uint32
	for _, instr := range instructions {
		bytes, err := ConvertInstructionToBinary(instr)
		if err != nil {
			return nil, fmt.Errorf("Error converting instruction '%v': %v", instr, err)
		}
		instruction := uint32(bytes[0])<<24 | uint32(bytes[1])<<16 | uint32(bytes[2])<<8 | uint32(bytes[3])
		instructionMemory = append(instructionMemory, instruction)
	}
	return instructionMemory, nil
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
	line = strings.ReplaceAll(line, ",", "")
	return line
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

	encoder, ok := encoderMap[opcode]
	if !ok {
		return nil, fmt.Errorf("No encoder available for opcode: %s", opcodeStr)
	}

	binaryInstruction, err := encoder(spec, instruction)
	if err != nil {
		return nil, fmt.Errorf("Error encoding instruction: %v", err)
	}

	bytes := make([]byte, 4)
	bytes[0] = byte((binaryInstruction >> 24) & 0xFF)
	bytes[1] = byte((binaryInstruction >> 16) & 0xFF)
	bytes[2] = byte((binaryInstruction >> 8) & 0xFF)
	bytes[3] = byte(binaryInstruction & 0xFF)

	return bytes, nil
}

func encodeRType(spec InstructionSpec, instruction []string) (uint32, error) {
	if spec.Opcode == CMP_OPCODE {
		if len(instruction) != 3 {
			return 0, fmt.Errorf("CMP instruction expects 2 operands, got %d: %v", len(instruction)-1, instruction)
		}
		// Para CMP, rd é fixo em 00000
		rs1 := RegisterToBinary(instruction[1])
		rs2 := RegisterToBinary(instruction[2])

		fmt.Printf("Encoding CMP: rs1=%s(%d), rs2=%s(%d)\n", instruction[1], rs1, instruction[2], rs2)

		binaryInstruction := uint32(spec.Opcode&0x3F) << 26
		binaryInstruction |= uint32(0&0x1F) << 21 // rd = 00000
		binaryInstruction |= uint32(rs1&0x1F) << 16
		binaryInstruction |= uint32(rs2&0x1F) << 11
		binaryInstruction |= uint32(spec.Funct3&0x1F) << 6
		binaryInstruction |= uint32(spec.Funct7 & 0x3F)

		return binaryInstruction, nil
	}

	if len(instruction) != 4 {
		return 0, fmt.Errorf("R-type instruction expects 3 operands, got %d: %v", len(instruction)-1, instruction)
	}

	rd := RegisterToBinary(instruction[1])
	rs1 := RegisterToBinary(instruction[2])
	rs2 := RegisterToBinary(instruction[3])

	fmt.Printf("Encoding R-Type: rd=%s(%d), rs1=%s(%d), rs2=%s(%d)\n", instruction[1], rd, instruction[2], rs1, instruction[3], rs2)

	binaryInstruction := uint32(spec.Opcode&0x3F) << 26
	binaryInstruction |= uint32(rd&0x1F) << 21
	binaryInstruction |= uint32(rs1&0x1F) << 16
	binaryInstruction |= uint32(rs2&0x1F) << 11
	binaryInstruction |= uint32(spec.Funct3&0x1F) << 6 // Shift correctly to bits 6-10
	binaryInstruction |= uint32(spec.Funct7 & 0x3F)

	return binaryInstruction, nil
}

func encodeIType(spec InstructionSpec, instruction []string) (uint32, error) {
	var rd_rs1 byte
	var immediate uint16

	if spec.Opcode == JUMP_OPCODE || spec.Opcode == FETCH_OPCODE {
		if len(instruction) != 2 {
			return 0, fmt.Errorf("%s instruction expects 1 operand, got %d: %v", instruction[0], len(instruction)-1, instruction)
		}
		// Para JUMP e FETCH, rs1/rd é fixo em 00000
		rd_rs1 = 0
		if spec.Opcode == JUMP_OPCODE {
			immediate = ImmediateToBinary(instruction[1])
		} else if spec.Opcode == FETCH_OPCODE {
			immediate = 0 // 0000000000000000
		}
		fmt.Printf("Encoding %s: rd_rs1=00000, immediate=%d\n", instruction[0], immediate)
	} else {
		if len(instruction) != 3 {
			return 0, fmt.Errorf("I-type instruction expects 2 operands, got %d: %v", len(instruction)-1, instruction)
		}
		rd_rs1 = RegisterToBinary(instruction[1])
		immediate = ImmediateToBinary(instruction[2])
		fmt.Printf("Encoding I-Type: rd_rs1=%s(%d), immediate=%d\n", instruction[1], rd_rs1, immediate)
	}

	binaryInstruction := uint32(spec.Opcode&0x3F) << 26
	binaryInstruction |= uint32(rd_rs1&0x1F) << 21
	binaryInstruction |= uint32(immediate&0xFFFF) << 5
	binaryInstruction |= uint32(spec.Funct3&0x1F) << 6 // Shift correctly to bits 6-10

	return binaryInstruction, nil
}

func RegisterToBinary(register string) byte {
	if reg, ok := registerMap[register]; ok {
		return byte(reg)
	}
	fmt.Printf("Invalid register: %s\n", register)
	return 0
}

func ImmediateToBinary(immediate string) uint16 {
	var uintValue uint64
	var err error

	// Verifica se o imediato está em formato hexadecimal
	if strings.HasPrefix(immediate, "0x") || strings.HasPrefix(immediate, "0X") {
		uintValue, err = strconv.ParseUint(immediate, 0, 16)
	} else {
		uintValue, err = strconv.ParseUint(immediate, 10, 16)
	}

	if err != nil {
		fmt.Printf("Invalid immediate value: %s\n", immediate)
		return 0
	}

	return uint16(uintValue)
}
