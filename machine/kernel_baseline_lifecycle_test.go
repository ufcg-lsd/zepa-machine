package machine

import (
	"encoding/binary"
	"testing"
)

// These tests enter the real syscall dispatcher with synthetic saved PCB state;
// they are not boot, interrupt-delivery, or user-program integration tests.
func TestLifecycleWaitBaseline(t *testing.T) {
	for _, scenario := range []string{"no child", "live child", "zombie child"} {
		t.Run(scenario, func(t *testing.T) {
			f := newKernelFixture(t)
			parent, child := uint32(legacyPCB), uint32(legacyPCB+legacyStride)
			f.put(0x2010, 0)
			f.put(parent+52, 0x100) // saved W8: mapped, aligned status address
			f.put(parent+56, 0xabcdef01)
			f.put(parent+72, 9)
			f.put(parent+80, binary.LittleEndian.Uint32(f.m.memory[fixtureUPTR:]))
			f.put(parent+4, ^uint32(0))
			if scenario != "no child" {
				f.put(parent+4, 1)
				f.put(child, 0)
				f.put(child+8, ^uint32(0))
				f.put(child+12, ^uint32(0))
				f.put(child+56, 42)
				f.put(child+72, 9)
				if scenario == "zombie child" {
					f.put(child+72, 11)
				}
			}
			statusFrame := binary.LittleEndian.Uint32(f.m.memory[fixtureUPTR:]) & (isPteMappedMask - 1)
			status := statusFrame*pageSize + 0x100
			binary.LittleEndian.PutUint32(f.m.memory[status:], 0xfeedface)
			f.m.registers[w9] = 1
			f.run("syscall_int", "schedule", 500)
			switch scenario {
			case "no child":
				if got := f.get(parent + 56); got != ^uint32(0) {
					t.Errorf("W9=%#x, want -1", got)
				}
			case "live child":
				if got := f.get(parent + 16); got != 0x100 {
					t.Errorf("status_addr=%#x", got)
				}
				if got := f.get(parent + 72); got != 21 {
					t.Errorf("blocked/waiting flags=%#x, want 0x15", got)
				}
				if got := f.get(child + 72); got != 9 {
					t.Errorf("live child flags changed: %#x", got)
				}
			case "zombie child":
				if got := f.get(parent + 56); got != 1 {
					t.Errorf("W9=%d, want child PID 1", got)
				}
				if got := binary.LittleEndian.Uint32(f.m.memory[status:]); got != 42 {
					t.Errorf("status=%d, want 42", got)
				}
				if got := f.get(parent + 4); got != ^uint32(0) {
					t.Errorf("child link=%#x", got)
				}
				if got := f.get(child + 72); got != 10 {
					t.Errorf("reaped child flags=%#x", got)
				}
			}
			if scenario != "zombie child" && binary.LittleEndian.Uint32(f.m.memory[status:]) != 0xfeedface {
				t.Error("status modified before child exit")
			}
		})
	}
}

func TestLifecycleExitOrphanBaseline(t *testing.T) {
	f := newKernelFixture(t)
	p := uint32(legacyPCB)
	f.put(0x2010, 0)
	f.put(p, ^uint32(0))
	f.put(p+4, ^uint32(0))
	f.put(p+52, 42)
	f.put(p+56, 0xabcdef01)
	f.put(p+72, 9)
	f.put(p+76, 0) // no allocated user pages in this synthetic PCB
	f.m.registers[w9] = 2
	f.run("syscall_int", "schedule", 500)
	if got := f.get(p + 56); got != 42 {
		t.Errorf("exit status=%d, want 42", got)
	}
	if got := f.get(p + 72); got != 8 {
		t.Errorf("orphan flags=%#x, want unmapped 8", got)
	}
}

func TestLifecycleReleBaseline(t *testing.T) {
	f := newKernelFixture(t)
	f.identityPCBPages(0, 1)
	p := uint32(legacyPCB)
	next := uint32(legacyPCB + legacyStride)
	f.put(0x2000, 2)
	f.put(0x2010, 0)
	f.put(p+56, 0xabcdef01)
	f.put(p+72, 1) // mapped and running
	f.put(next+56, 0x12345678)
	f.put(next+60, 0x100)
	f.put(next+64, 0x800)
	f.put(next+68, 40)
	f.put(next+72, 9) // mapped and ready
	f.m.registers[w9] = 4
	f.run("syscall_int", "schedule", 100)
	if got := f.get(p + 56); got != 0xabcdef01 {
		t.Errorf("rele changed saved W9: %#x", got)
	}
	if got := f.get(p + 72); got != 1 {
		t.Errorf("rele changed running flags before scheduler: %#x", got)
	}
	f.runToMRET("schedule", 500)
	if got := f.get(p + 72); got != 9 {
		t.Errorf("yielding process flags=%#x, want ready", got)
	}
	if got := f.get(0x2010); got != 1 {
		t.Errorf("scheduled PID=%d, want 1", got)
	}
	if got := f.m.registers[pc]; got != 0x100 {
		t.Errorf("resumed PC=%#x", got)
	}
	if got := f.m.registers[w9]; got != 0x12345678 {
		t.Errorf("resumed W9=%#x", got)
	}
}
