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
	err := processFile(sourceFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(instructions)

}

func processFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening the file '%s': %v", filename, err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	linePos := 1
	for scanner.Scan() {
		line := processLine(scanner.Text())
		if line != "" {
			parts := strings.Fields(line)
			instructions[linePos] = parts
			linePos++
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading the file '%s': %v", filename, err)
	}

	return nil
}

func processLine(line string) string {
	line = strings.TrimSpace(line)

	if idx := strings.Index(line, ";"); idx != -1 {
		line = strings.TrimSpace(line[:idx])
	}

	if line == "" {
		return ""
	}

	return line
}
