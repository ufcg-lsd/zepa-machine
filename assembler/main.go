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
type Opcode byte

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

	MV Opcode = iota
	ADD
	SUB
	CMP
	JUMP
	LOAD
	STORE
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
	"MV":    MV,
	"ADD":   ADD,
	"SUB":   SUB,
	"CMP":   CMP,
	"JUMP":  JUMP,
	"LOAD":  LOAD,
	"STORE": STORE,
}

var encoderMap = map[Opcode]func(Opcode, []string) (uint32, error){
	ADD:   encodeRType,
	SUB:   encodeRType,
	CMP:   encodeRType,
	MV:    encodeIType,
	LOAD:  encodeIType,
	STORE: encodeIType,
	JUMP:  encodeIType,
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
		fmt.Println(instructionBits)

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
		bytes := ConvertInstructionToBinary(instr)
		if bytes == nil {
			return nil, fmt.Errorf("Error converting instruction: %v", instr)
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

func ConvertInstructionToBinary(instruction []string) []byte {
	if len(instruction) == 0 {
		return nil
	}

	opcodeStr := instruction[0]
	opcode, ok := opcodeMap[opcodeStr]
	if !ok {
		fmt.Printf("Invalid opcode: %s\n", opcodeStr)
		return nil
	}

	encoder, ok := encoderMap[opcode]
	if !ok {
		fmt.Printf("No encoder available for opcode: %s\n", opcodeStr)
		return nil
	}

	binaryInstruction, err := encoder(opcode, instruction)
	if err != nil {
		fmt.Printf("Error encoding instruction: %v\n", err)
		return nil
	}

	bytes := make([]byte, 4)
	bytes[0] = byte((binaryInstruction >> 24) & 0xFF)
	bytes[1] = byte((binaryInstruction >> 16) & 0xFF)
	bytes[2] = byte((binaryInstruction >> 8) & 0xFF)
	bytes[3] = byte(binaryInstruction & 0xFF)

	return bytes
}

func encodeRType(opcode Opcode, instruction []string) (uint32, error) {
	if len(instruction) != 4 {
		return 0, fmt.Errorf("Invalid number of operands for R-type instruction: %v", instruction)
	}

	rd := RegisterToBinary(instruction[1])
	rs1 := RegisterToBinary(instruction[2])
	rs2 := RegisterToBinary(instruction[3])

	funct3 := byte(0)
	funct7 := byte(0)

	binaryInstruction := uint32(opcode&0x3F) << 26
	binaryInstruction |= uint32(rd&0x1F) << 21
	binaryInstruction |= uint32(rs1&0x1F) << 16
	binaryInstruction |= uint32(rs2&0x1F) << 11
	binaryInstruction |= uint32(funct3&0x07) << 8
	binaryInstruction |= uint32(funct7 & 0x7F)

	return binaryInstruction, nil
}

func encodeIType(opcode Opcode, instruction []string) (uint32, error) {
	if len(instruction) != 3 {
		return 0, fmt.Errorf("Invalid number of operands for I-type instruction: %v", instruction)
	}

	rd_rs1 := RegisterToBinary(instruction[1])
	immediate := ImmediateToBinary(instruction[2])

	funct3 := byte(0)

	binaryInstruction := uint32(opcode&0x3F) << 26
	binaryInstruction |= uint32(rd_rs1&0x1F) << 21
	binaryInstruction |= uint32(immediate&0xFFFF) << 5
	binaryInstruction |= uint32(funct3 & 0x1F)

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
	value := strings.TrimPrefix(immediate, "#")
	intValue, err := strconv.Atoi(value)
	if err != nil {
		fmt.Printf("Invalid immediate value: %s\n", immediate)
		return 0
	}
	return uint16(intValue)
}
