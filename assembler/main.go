package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var instructions = make(map[int][]string)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: assembler <input_file.txt>")
		return
	}

	sourceFile := os.Args[1]
	err := RunAssembler(sourceFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(instructions)
}

func RunAssembler(filePath string) error {
	instrs, err := LoadAssemblyFile(filePath)
	if err != nil {
		return err
	}

	instructions = instrs
	return nil
}

func processFile(filename string) (map[int][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening the file '%s': %v", filename, err)
	}
	defer file.Close()

	instructions := make(map[int][]string)
	scanner := bufio.NewScanner(file)
	linePos := 1
	for scanner.Scan() {
		line := processLine(scanner.Text())
		if line != "" {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				instructions[linePos] = parts
				linePos++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading the file '%s': %v", filename, err)
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

func LoadAssemblyFile(filePath string) (map[int][]string, error) {
	return processFile(filePath)
}
