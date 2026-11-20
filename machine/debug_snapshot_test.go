package machine

import "testing"

func TestGetRegistersSnapshotIsDetached(t *testing.T) {
	m := NewMachine(4, false)
	m.registers[w0] = 7

	snapshot := m.GetRegistersSnapshot()
	snapshot[w0] = 8
	m.registers[w0] = 9

	if got := snapshot[w0]; got != 8 {
		t.Fatalf("snapshot W0 = %d, want detached value 8", got)
	}
	if got := m.registers[w0]; got != 9 {
		t.Fatalf("machine W0 = %d, want 9", got)
	}
}
