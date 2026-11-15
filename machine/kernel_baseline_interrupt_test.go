package machine

import (
	"encoding/binary"
	"testing"

	assembler "zepa-machine/cross-assembler"
)

// This fixture configures a 4 KiB buffer, not the usual 64 KiB baseline or a
// host-provided interrupt buffer. It occupies the last virtual page; fixture
// addresses are offsets from the kernel boundary, including this wrapped address.
func interruptBuffer(f *kernelFixture, size, value uint32) {
	f.t.Helper()
	const buffer = uint32(0x3ffff000)
	f.mapKernel(buffer)
	f.put(0x2008, 4096)
	f.put(buffer, size)
	f.put(buffer+4, value)
}

// This probes only the sample's parent instructions after fork, not a full
// syscall lifecycle. wait reads saved W8 as status_addr (specs.md and wait).
func TestBaselineForkSampleWaitArgument(t *testing.T) {
	f := newKernelFixture(t)
	image, err := assembler.RunAssembler("../asm/samples/fork_test.asm")
	if err != nil {
		t.Fatal(err)
	}
	frame, ok := f.m.lookupPhysical(0, 4)
	if !ok {
		t.Fatal("sample user page is unmapped")
	}
	copy(f.m.memory[frame:], image)
	f.m.registers[pc] = 20         // parent fallthrough immediately after fork branch
	f.m.registers[w8] = 0x12345678 // inherited value, not set by this sample
	for i := 0; i < 4; i++ {
		if !f.m.fetch() {
			t.Fatal("sample fetch fault")
		}
		inst, ok := f.m.decode()
		if !ok || f.m.checkIllegalInstruction(inst) || f.m.checkIllegalRegisterAccess(inst) {
			t.Fatal("sample decode/access fault")
		}
		f.m.execute(inst)
	}
	if f.m.registers[pc] != 36 || f.m.registers[w9] != 1 || f.m.registers[w8] != 0x12345678 {
		t.Fatalf("before wait syscall: PC=%#x W9=%#x W8=%#x", f.m.registers[pc], f.m.registers[w9], f.m.registers[w8])
	}
	t.Log("fork_test parent reaches wait without setting W8; validity depends on inherited context; no wait syscall executed")
}

func TestBaselineKillInterruptRejected(t *testing.T) {
	for _, tc := range []struct {
		name       string
		pid, flags uint32
	}{
		{"out-of-range", 1, 9},
		{"unmapped", 0, 8},
		{"zombie", 0, 11},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newKernelFixture(t)
			interruptBuffer(f, 4, tc.pid)
			f.put(0x2000, 1)
			f.put(legacyPCB+72, tc.flags)
			f.put(legacyPCB+56, 0x12345678)
			f.run("kill_int", "schedule", 200)
			if got := f.get(legacyPCB + 56); got != 0x12345678 || f.get(legacyPCB+72) != tc.flags {
				t.Fatalf("rejected kill changed PCB: W9=%#x flags=%#x", got, f.get(legacyPCB+72))
			}
		})
	}
}

func TestBaselineKillInterruptOrphan(t *testing.T) {
	f := newKernelFixture(t)
	interruptBuffer(f, 4, 0)
	f.put(0x2000, 1)
	f.put(legacyPCB, ^uint32(0))
	f.put(legacyPCB+4, ^uint32(0))
	f.put(legacyPCB+72, 9)
	f.put(legacyPCB+76, 0)
	f.run("kill_int", "schedule", 500)
	if got := f.get(legacyPCB + 56); got != 2 || f.get(legacyPCB+72) != 8 {
		t.Fatalf("orphan kill: W9=%#x flags=%#x; want 2 and 8", got, f.get(legacyPCB+72))
	}
}

func TestBaselineKillInterruptZombie(t *testing.T) {
	f := newKernelFixture(t)
	interruptBuffer(f, 4, 1)
	f.put(0x2000, 2)
	child := uint32(legacyPCB + legacyStride)
	f.put(legacyPCB+72, 9)
	f.put(legacyPCB+4, 1)
	f.put(child, 0)
	f.put(child+4, ^uint32(0))
	f.put(child+72, 9)
	f.put(child+76, 0)
	f.run("kill_int", "schedule", 500)
	if f.get(child+56) != 2 || f.get(child+72) != 11 || f.get(legacyPCB+4) != 1 || f.get(legacyPCB+72) != 9 {
		t.Fatalf("zombie kill: W9=%#x flags=%#x parent child=%d flags=%#x", f.get(child+56), f.get(child+72), f.get(legacyPCB+4), f.get(legacyPCB+72))
	}
}

func TestBaselineInputInterruptNoSlot(t *testing.T) {
	f := newKernelFixture(t)
	interruptBuffer(f, 4, 0x12345678)
	f.put(0x2000, 1)
	f.put(legacyPCB+72, 9)
	f.put(legacyPCB+56, 0xabcdef01)
	f.run("input_int", "schedule", 200)
	if f.get(legacyPCB+72) != 9 || f.get(legacyPCB+56) != 0xabcdef01 || f.get(0x201c) != 4 {
		t.Fatalf("full input: flags=%#x W9=%#x size=%d", f.get(legacyPCB+72), f.get(legacyPCB+56), f.get(0x201c))
	}
}

func TestBaselineInputInterruptLoadsProgram(t *testing.T) {
	f := newKernelFixture(t)
	f.identityPCBPages(0, 1)
	interruptBuffer(f, 4, 0x12345678)
	const stack = uint32(0x1000000)
	f.mapKernel(stack)
	f.put(0x2018, kernelBoundary+stack+pageSize)
	f.put(0x2000, 2)
	f.put(legacyPCB+72, 9) // input must skip occupied PID 0
	f.put(legacyPCB+56, 0xabcdef01)
	child := uint32(legacyPCB + legacyStride)
	f.initBitmap()
	f.run("input_int", "schedule", 100000)
	for off := uint32(0); off <= 16; off += 4 {
		if got := f.get(child + off); got != ^uint32(0) {
			t.Fatalf("PCB link offset %d = %#x, want -1", off, got)
		}
	}
	for off := uint32(20); off <= 60; off += 4 {
		if got := f.get(child + off); got != 0 {
			t.Fatalf("PCB context offset %d = %#x, want 0", off, got)
		}
	}
	if f.get(child+64) != kernelBoundary || f.get(child+68) != 56 || f.get(child+72) != 9 || f.get(child+76) != 1 {
		t.Fatalf("input PCB sp=%#x sr=%#x flags=%#x pages=%d", f.get(child+64), f.get(child+68), f.get(child+72), f.get(child+76))
	}
	pte := f.get(child + 80)
	frame := (pte & (isPteMappedMask - 1)) * pageSize
	if pte&isPteMappedMask == 0 || frame+pageSize > uint32(len(f.m.memory)) {
		t.Fatalf("invalid input PTE %#x", pte)
	}
	frameID := frame / pageSize
	bitmapWord := f.get(legacyBitmap + (frameID/32)*4)
	if bitmapWord&(1<<(frameID%32)) == 0 {
		t.Fatalf("input frame %d not marked allocated in bitmap word %#x", frameID, bitmapWord)
	}
	if got := binary.LittleEndian.Uint32(f.m.memory[frame:]); got != 0x12345678 {
		t.Fatalf("input payload = %#x", got)
	}
	for i, b := range f.m.memory[frame+4 : frame+pageSize] {
		if b != 0 {
			t.Fatalf("input padding byte %d = %#x", i+4, b)
		}
	}
	if f.get(legacyPCB+56) != 0xabcdef01 || f.get(legacyPCB+72) != 9 {
		t.Fatal("input modified occupied PCB")
	}
}
