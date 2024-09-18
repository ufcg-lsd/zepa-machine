package main

import (
	"reflect"
	"testing"
)

func TestMultiplyTwoNumbers(t *testing.T) {
	asmFile := "../asm/samples/multiply_two_numbers.asm"

	err := RunAssembler(asmFile)
	if err != nil {
		t.Errorf("Error in RunAssembler: %v", err)
	}

	expectedInstructions := map[int][]string{
		1: {"MV", "W1", "#4"},
		2: {"MV", "W2", "#6"},
		3: {"MUL", "W0", "W1", "W2"},
	}

	if !reflect.DeepEqual(instructions, expectedInstructions) {
		t.Errorf("Expected %v, but got %v", expectedInstructions, instructions)
	}
}

func TestGCD(t *testing.T) {
	asmFile := "../asm/samples/gcd.asm"

	err := RunAssembler(asmFile)
	if err != nil {
		t.Errorf("Error in RunAssembler: %v", err)
	}

	expectedInstructions := map[int][]string{
		1: {"MV", "W0", "#48"},
		2: {"MV", "W1", "#18"},
		3: {"CMP", "W0", "W1"},
		4: {"JZ", "_end"},
		5: {"JG", "_subtract_W0"},
		6: {"SUB", "W1", "W1", "W0"},
		7: {"JUMP", "_loop"},
		8: {"SUB", "W0", "W0", "W1"},
		9: {"JUMP", "_loop"},
	}

	if !reflect.DeepEqual(instructions, expectedInstructions) {
		t.Errorf("Expected %v, but got %v", expectedInstructions, instructions)
	}
}

func TestAddThreeNumbers(t *testing.T) {
	asmFile := "../asm/samples/sum_list.asm"

	err := RunAssembler(asmFile)
	if err != nil {
		t.Errorf("Error in RunAssembler: %v", err)
	}

	expectedInstructions := map[int][]string{
		1: {"MV", "W0", "#0"},
		2: {"MV", "W1", "#1"},
		3: {"MV", "W2", "#2"},
		4: {"MV", "W3", "#3"},
		5: {"ADD", "W0", "W0", "W1"},
		6: {"ADD", "W0", "W0", "W2"},
		7: {"ADD", "W0", "W0", "W3"},
	}

	if !reflect.DeepEqual(instructions, expectedInstructions) {
		t.Errorf("Expected %v, but got %v", expectedInstructions, instructions)
	}
}

func TestFindMaximum(t *testing.T) {
	asmFile := "../asm/samples/find_maximum.asm"

	err := RunAssembler(asmFile)
	if err != nil {
		t.Errorf("Error in RunAssembler: %v", err)
	}

	expectedInstructions := map[int][]string{
		1:  {"MV", "W0", "#3"},
		2:  {"MV", "W1", "#7"},
		3:  {"MV", "W2", "#2"},
		4:  {"MV", "W3", "#9"},
		5:  {"MV", "W4", "W0"},
		6:  {"CMP", "W1", "W4"},
		7:  {"JG", "_update_W1"},
		8:  {"JUMP", "_check_W2"},
		9:  {"MV", "W4", "W1"},
		10: {"CMP", "W2", "W4"},
		11: {"JG", "_update_W2"},
		12: {"JUMP", "_check_W3"},
		13: {"MV", "W4", "W2"},
		14: {"CMP", "W3", "W4"},
		15: {"JG", "_update_W3"},
		16: {"JUMP", "_end"},
		17: {"MV", "W4", "W3"},
	}

	if !reflect.DeepEqual(instructions, expectedInstructions) {
		t.Errorf("Expected %v, but got %v", expectedInstructions, instructions)
	}
}

func TestFactorial(t *testing.T) {
	asmFile := "../asm/samples/factorial.asm"

	err := RunAssembler(asmFile)
	if err != nil {
		t.Errorf("Error in RunAssembler: %v", err)
	}

	expectedInstructions := map[int][]string{
		1: {"MV", "W0", "#1"},
		2: {"MV", "W1", "#5"},
		3: {"CMP", "W1", "#1"},
		4: {"JZ", "_end"},
		5: {"MUL", "W0", "W0", "W1"},
		6: {"SUB", "W1", "W1", "#1"},
		7: {"JUMP", "_loop"},
	}

	if !reflect.DeepEqual(instructions, expectedInstructions) {
		t.Errorf("Expected %v, but got %v", expectedInstructions, instructions)
	}
}

func TestAddTwoNumbers(t *testing.T) {
	asmFile := "../asm/samples/add_two_number.asm"

	err := RunAssembler(asmFile)
	if err != nil {
		t.Errorf("Error in RunAssembler: %v", err)
	}

	expectedInstructions := map[int][]string{
		1: {"MV", "W1", "#5"},
		2: {"MV", "W2", "#3"},
		3: {"ADD", "W0", "W1", "W2"},
	}

	if !reflect.DeepEqual(instructions, expectedInstructions) {
		t.Errorf("Expected %v, but got %v", expectedInstructions, instructions)
	}
}
