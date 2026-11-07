package machine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
)

func putDebugLayoutWord(memory []byte, addr uint32, value uint32) {
	binary.LittleEndian.PutUint32(memory[int(addr):int(addr)+4], value)
}

func putDebugLayoutByte(memory []byte, addr uint32, value byte) {
	memory[int(addr)] = value
}

func TestDebugUsesRelocatedGlobalsAndFirstPCBReadOnly(t *testing.T) {
	memory := make([]byte, int(PCBVectorStart)+int(PCBPageTableOffset)+4)

	// Put unmistakable sentinels at the predecessor global and PCB locations.
	for offset, value := range map[uint32]uint32{
		0x2000: 9,
		0x2004: 0x11111111,
		0x2008: 0x22222222,
		0x200C: 0x33333333,
		0x2010: 0x44444444,
		0x2014: 0x55555555,
		0x2018: 0x66666666,
		0x201C: 0x77777777,
		0x2020: 0x88888888,
	} {
		putDebugLayoutWord(memory, offset, value)
	}
	legacyPCB := uint32(0x2024)
	putDebugLayoutWord(memory, legacyPCB+PCBParentOffset, 901)
	putDebugLayoutWord(memory, legacyPCB+PCBW0Offset, 902)
	putDebugLayoutWord(memory, legacyPCB+PCBProgramCounterOffset, 0xDEADBEEF)
	putDebugLayoutByte(memory, legacyPCB+PCBFlagsOffset, 1)

	for offset, value := range map[uint32]uint32{
		GlobalMaxProcessesOffset:       1,
		GlobalTimeSliceOffset:          0x12345678,
		GlobalBufferSizeOffset:         0xA1A2A3A4,
		GlobalMemorySizeOffset:         0x00004000,
		GlobalRunningPIDOffset:         0,
		GlobalClockCountOffset:         17,
		GlobalKernelStackPointerOffset: 0xFFFD0000,
		GlobalScratch0Offset:           0x01020304,
		GlobalScratch1Offset:           0x05060708,
	} {
		putDebugLayoutWord(memory, uint32(GlobalAddress(offset)), value)
	}

	pcb, ok := PCBAddress(0)
	if !ok || pcb != PCBVectorStart {
		t.Fatalf("PCBAddress(0) = %#x, valid=%t; want PCBVectorStart %#x", pcb, ok, PCBVectorStart)
	}
	putDebugLayoutWord(memory, uint32(pcb)+PCBParentOffset, 42)
	putDebugLayoutWord(memory, uint32(pcb)+PCBChildOffset, 43)
	putDebugLayoutWord(memory, uint32(pcb)+PCBPreviousSiblingOffset, 44)
	putDebugLayoutWord(memory, uint32(pcb)+PCBNextSiblingOffset, 45)
	putDebugLayoutWord(memory, uint32(pcb)+PCBStatusAddressOffset, 0x123)
	for register := uint32(0); register <= 9; register++ {
		putDebugLayoutWord(memory, uint32(pcb)+PCBW0Offset+register*4, 101+register)
	}
	putDebugLayoutWord(memory, uint32(pcb)+PCBProgramCounterOffset, 0x00401234)
	putDebugLayoutWord(memory, uint32(pcb)+PCBStackPointerOffset, 0x7FFFEFFC)
	putDebugLayoutWord(memory, uint32(pcb)+PCBStatusRegisterOffset, 0x5A)
	putDebugLayoutByte(memory, uint32(pcb)+PCBFlagsOffset, 1)
	putDebugLayoutWord(memory, uint32(pcb)+PCBPagesUsedOffset, 0)

	before := append([]byte(nil), memory...)
	m := &Machine{memory: memory}

	vars := m.GetKernelVarsString()
	for _, want := range []string{
		"0x00000001",
		"0x12345678",
		"0xA1A2A3A4",
		"0x00004000",
		"0x00000011",
		"0xFFFD0000",
		"0x01020304",
		"0x05060708",
		"0x00004024",
	} {
		if !strings.Contains(vars, want) {
			t.Errorf("kernel vars do not display relocated value %q:\n%s", want, vars)
		}
	}
	for _, oldValue := range []string{
		"0x00000009", "0x11111111", "0x22222222", "0x33333333", "0x44444444",
		"0x55555555", "0x66666666", "0x77777777", "0x88888888",
	} {
		if strings.Contains(vars, oldValue) {
			t.Errorf("kernel vars used predecessor global sentinel %s:\n%s", oldValue, vars)
		}
	}

	processes := m.GetProcessTableString()
	for _, want := range []string{"42", "101", "4198964", "2147479548"} {
		if !strings.Contains(processes, want) {
			t.Errorf("process table does not display relocated PCB value %q:\n%s", want, processes)
		}
	}
	for _, oldValue := range []string{"901", "902", "-559038737"} {
		if strings.Contains(processes, oldValue) {
			t.Errorf("process table used predecessor PCB sentinel %s:\n%s", oldValue, processes)
		}
	}

	pcbView := m.GetPCBVectorString()
	if !strings.Contains(pcbView, "42") || !strings.Contains(pcbView, "4198964") || !strings.Contains(pcbView, "2147479548") {
		t.Errorf("PCB vector does not preserve relocated PCB values and user virtual addresses:\n%s", pcbView)
	}
	if strings.Contains(pcbView, "901") || strings.Contains(pcbView, "902") {
		t.Errorf("PCB vector used predecessor PCB sentinels:\n%s", pcbView)
	}

	pageTable := m.GetProcessPageTableString(0)
	if !strings.Contains(pageTable, "No valid entries") {
		t.Errorf("process page table did not read relocated PCB0:\n%s", pageTable)
	}
	if !bytes.Equal(memory, before) {
		t.Fatal("debug reads modified the crafted memory")
	}
}

func TestDebugRejectsPIDOutsidePCBVector(t *testing.T) {
	memory := make([]byte, int(PCBVectorStart)+int(PCBFlagsOffset)+1)
	putDebugLayoutWord(memory, uint32(GlobalAddress(GlobalMaxProcessesOffset)), PCBCount+1)

	got := (&Machine{memory: memory}).GetProcessPageTableString(PCBCount)
	want := fmt.Sprintf("PID %d out of range (MAX_PROCESSES = %d).", PCBCount, PCBCount+1)
	if got != want {
		t.Fatalf("out-of-range PID response = %q, want %q", got, want)
	}
}

func TestDebugLayoutKPTRAndBitmapAddressSelection(t *testing.T) {
	if got := uint32(KernelPageTableStart); got != 0x30009024 {
		t.Fatalf("kernel page table starts at %#x, want shared KPTR base %#x", got, 0x30009024)
	}
	if got := uint32(KernelPageTableStart) + debugKernelPTECount*debugKernelPTEStride; got != uint32(KernelPageTableEnd) {
		t.Fatalf("kernel page table end by PTE arithmetic = %#x, want %#x", got, KernelPageTableEnd)
	}
	if got := uint32(BitmapStart); got != 0x30109024 {
		t.Fatalf("bitmap starts at %#x, want shared bitmap base %#x", got, 0x30109024)
	}
	if got := bitmapWordAddress(0); got != BitmapStart {
		t.Fatalf("bitmap word for frame 0 = %#x, want shared bitmap base %#x", got, BitmapStart)
	}
	if got, want := bitmapWordAddress(32), BitmapStart+4; got != want {
		t.Fatalf("bitmap word for frame 32 = %#x, want %#x", got, want)
	}
	lastFrame := PhysicalMemoryBytes/pageSize - 1
	lastWord := BitmapStart + PhysicalAddress((lastFrame>>5)*4)
	if got := bitmapWordAddress(lastFrame); got != lastWord || got >= BitmapEnd {
		t.Fatalf("bitmap word for last physical frame = %#x, want %#x within bitmap ending %#x", got, lastWord, BitmapEnd)
	}
}
