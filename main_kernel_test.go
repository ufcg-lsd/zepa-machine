//go:build kernel

package main

import (
	"bytes"
	"testing"

	"zepa-machine/machine"
)

func TestInitializeKernelGlobalsUsesSharedLayout(t *testing.T) {
	const memorySize = 0x400C
	memory := make([]byte, memorySize)
	canary := memory[:0x4000]
	for i := range canary {
		canary[i] = byte(i*31 + 7)
	}
	canaryBefore := bytes.Clone(canary)

	const timeSlice = uint32(0x12345678)
	if err := initializeKernelGlobals(memory, timeSlice); err != nil {
		t.Fatalf("initializeKernelGlobals() error = %v", err)
	}

	if !bytes.Equal(canary, canaryBefore) {
		t.Fatal("kernel code region [0x0000,0x4000) changed")
	}

	globals := []struct {
		name    string
		offset  uint32
		address int
		want    []byte
	}{
		{
			name:    "MAX_PROCESSES",
			offset:  machine.GlobalMaxProcessesOffset,
			address: 0x4000,
			want:    []byte{0x00, 0x01, 0x00, 0x00},
		},
		{
			name:    "TIME_SLICE",
			offset:  machine.GlobalTimeSliceOffset,
			address: 0x4004,
			want:    []byte{0x78, 0x56, 0x34, 0x12},
		},
		{
			name:    "BUFFER_SIZE",
			offset:  machine.GlobalBufferSizeOffset,
			address: 0x4008,
			want:    []byte{0x00, 0x00, 0x01, 0x00},
		},
	}
	for _, global := range globals {
		t.Run(global.name, func(t *testing.T) {
			address := machine.GlobalAddress(global.offset)
			if int(address) != global.address {
				t.Fatalf("GlobalAddress(%#x) = %#x, want %#x", global.offset, address, global.address)
			}
			got := memory[global.address : global.address+4]
			if !bytes.Equal(got, global.want) {
				t.Fatalf("bytes at %#x = % x, want little-endian % x", global.address, got, global.want)
			}
		})
	}
}

func TestInitializeKernelGlobalsRejectsTruncatedBackingWithoutWrites(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{name: "before MAX_PROCESSES", size: 0x4000},
		{name: "before TIME_SLICE", size: 0x4004},
		{name: "before BUFFER_SIZE", size: 0x4008},
		{name: "partial BUFFER_SIZE", size: 0x400B},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memory := make([]byte, tt.size)
			for i := range memory {
				memory[i] = byte(i%255 + 1)
			}
			before := bytes.Clone(memory)

			if err := initializeKernelGlobals(memory, 0x12345678); err == nil {
				t.Fatal("initializeKernelGlobals() error = nil, want error")
			}
			if !bytes.Equal(memory, before) {
				t.Fatal("initializeKernelGlobals() modified truncated backing")
			}
		})
	}
}
