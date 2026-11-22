package machine

import (
	"encoding/binary"
	"errors"
	"runtime"
	"testing"
	"time"
)

const (
	safePointHandlerAddr uint32 = 0x100
	srInterruptEnabled   uint32 = 16

	// JUMP with immediate 0 is PC += -4 after fetch: a jump to itself.
	jumpSelfWord uint32 = uint32(JUMP) << 26
	mretWord     uint32 = uint32(MRET) << 26

	safePointTimeout = 2 * time.Second
)

// newSafePointMachine builds a kernel-mode machine with a self-loop at 0 and
// an MRET handler at safePointHandlerAddr. The backing covers the transfer
// buffer and control region, which writeTransferBufferLocked requires.
func newSafePointMachine(t *testing.T, debug bool, status uint32) *Machine {
	t.Helper()
	m := NewMachine(int(RunnerMinimumBackingBytes), debug)
	binary.BigEndian.PutUint32(m.memory[0:4], jumpSelfWord)
	binary.BigEndian.PutUint32(m.memory[safePointHandlerAddr:safePointHandlerAddr+4], mretWord)
	m.registers[esa] = safePointHandlerAddr
	m.registers[sr] = status
	return m
}

func transferBufferSnapshot(m *Machine, n int) []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	start := int(TransferBufferStart)
	return append([]byte(nil), m.memory[start:start+4+n]...)
}

func expectTransferBuffer(t *testing.T, m *Machine, payload []byte) {
	t.Helper()
	got := transferBufferSnapshot(m, len(payload))
	if length := binary.LittleEndian.Uint32(got[:4]); length != uint32(len(payload)) {
		t.Fatalf("transfer length = %d, want %d", length, len(payload))
	}
	for i, b := range payload {
		if got[4+i] != b {
			t.Fatalf("transfer byte %d = %#x, want %#x", i, got[4+i], b)
		}
	}
}

func expectTransferBufferUntouched(t *testing.T, m *Machine) {
	t.Helper()
	for i, b := range transferBufferSnapshot(m, 4) {
		if b != 0 {
			t.Fatalf("transfer buffer byte %d = %#x before safe point, want 0", i, b)
		}
	}
}

func pidPayload(pid uint32) []byte {
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, pid)
	return payload
}

// waitFor yields until cond holds. It bounds the wait with a deadline and is
// used only where the machine offers no completion signal (control servicing
// at a stopped CPU deliberately sends nothing on DoneChan).
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(safePointTimeout)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		runtime.Gosched()
	}
}

func step(t *testing.T, m *Machine) {
	t.Helper()
	select {
	case m.StepChan <- struct{}{}:
	case <-time.After(safePointTimeout):
		t.Fatal("step not accepted: CPU owner deadlocked")
	}
	select {
	case <-m.DoneChan:
	case <-time.After(safePointTimeout):
		t.Fatal("step not completed: CPU owner deadlocked")
	}
}

func TestControlRequestDuringBootstrapLeavesRAMUntouched(t *testing.T) {
	m := newSafePointMachine(t, false, 0) // interrupts still disabled

	if err := m.RequestKill(7); err != nil {
		t.Fatalf("RequestKill: %v", err)
	}
	m.mu.Lock()
	m.serviceControlLocked()
	m.mu.Unlock()

	expectTransferBufferUntouched(t, m)
	if !m.pendingKill.active {
		t.Fatal("kill request dropped outside a safe point")
	}

	m.registers[sr] = srInterruptEnabled
	m.mu.Lock()
	m.serviceControlLocked()
	m.mu.Unlock()

	expectTransferBuffer(t, m, pidPayload(7))
	if m.registers[ecr] != killInt || m.registers[pc] != safePointHandlerAddr {
		t.Fatalf("ecr=%d pc=%#x, want killInt at handler", m.registers[ecr], m.registers[pc])
	}
	if m.pendingKill.active {
		t.Fatal("kill request still pending after delivery")
	}
}

func TestControlWakeAtStoppedSafePointDoesNotConsumeStep(t *testing.T) {
	m := newSafePointMachine(t, true, srInterruptEnabled)
	go m.Boot()

	if err := m.RequestKill(5); err != nil {
		t.Fatalf("RequestKill: %v", err)
	}
	waitFor(t, "kill delivery at stopped safe point", func() bool {
		return m.GetRegisters()[ecr] == killInt
	})

	if got := m.GetInstructionsExecuted(); got != 0 {
		t.Fatalf("instructions executed = %d, want 0", got)
	}
	if pc := m.GetRegisters()[pc]; pc != safePointHandlerAddr {
		t.Fatalf("pc = %#x, want handler %#x", pc, safePointHandlerAddr)
	}
	expectTransferBuffer(t, m, pidPayload(5))

	// A normal step still completes, so the control path left no stray
	// DoneChan send behind.
	step(t, m)
	if got := m.GetInstructionsExecuted(); got != 1 {
		t.Fatalf("instructions executed after step = %d, want 1", got)
	}
}

func TestControlStaysPendingInStoppedHandlerUntilMRET(t *testing.T) {
	// Stopped inside a handler: interrupts disabled, MRET at PC, and the
	// interrupted context (loop at 0 with interrupts enabled) saved.
	m := newSafePointMachine(t, true, 0)
	m.registers[pc] = safePointHandlerAddr
	m.registers[epc] = 0
	m.registers[esr] = srInterruptEnabled
	go m.Boot()

	if err := m.RequestKill(9); err != nil {
		t.Fatalf("RequestKill: %v", err)
	}
	waitFor(t, "owner to consume the control wakeup", func() bool {
		return len(m.controlWake) == 0
	})

	expectTransferBufferUntouched(t, m)
	if got := m.GetInstructionsExecuted(); got != 0 {
		t.Fatalf("instructions executed = %d, want 0", got)
	}

	step(t, m) // executes MRET; the boundary after it is a safe point

	expectTransferBuffer(t, m, pidPayload(9))
	regs := m.GetRegisters()
	if regs[ecr] != killInt || regs[pc] != safePointHandlerAddr || regs[epc] != 0 {
		t.Fatalf("ecr=%d pc=%#x epc=%#x, want killInt delivered after MRET", regs[ecr], regs[pc], regs[epc])
	}
	if got := m.GetInstructionsExecuted(); got != 1 {
		t.Fatalf("instructions executed = %d, want 1", got)
	}
}

func TestControlKillAndInputUseSeparateSafePoints(t *testing.T) {
	m := newSafePointMachine(t, false, srInterruptEnabled)
	code := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x11, 0x22}

	if err := m.RequestKill(3); err != nil {
		t.Fatalf("RequestKill: %v", err)
	}
	if err := m.RequestInput(code); err != nil {
		t.Fatalf("RequestInput: %v", err)
	}

	m.serviceControlLocked()
	expectTransferBuffer(t, m, pidPayload(3))
	if m.registers[ecr] != killInt || !m.pendingInput.active {
		t.Fatal("kill must be delivered first and input must stay pending")
	}

	m.serviceControlLocked() // inside the kill handler: not a safe point
	expectTransferBuffer(t, m, pidPayload(3))

	m.mret(Instruction{})
	m.serviceControlLocked()
	expectTransferBuffer(t, m, code)
	if m.registers[ecr] != inputInt || m.pendingInput.active {
		t.Fatal("input not delivered at the safe point after MRET")
	}
}

func TestControlRequestsAreBoundedAndNonBlocking(t *testing.T) {
	m := newSafePointMachine(t, false, srInterruptEnabled)

	if err := m.RequestKill(1); err != nil {
		t.Fatalf("first RequestKill: %v", err)
	}
	if err := m.RequestKill(2); !errors.Is(err, ErrControlBusy) {
		t.Fatalf("second RequestKill error = %v, want ErrControlBusy", err)
	}
	for i := 0; i < 100; i++ {
		_ = m.RequestKill(uint32(i)) // nothing serves the wakeup; must not block
	}
	if n := len(m.controlWake); n != 1 {
		t.Fatalf("pending wakeups = %d, want 1 (coalesced)", n)
	}

	tooLarge := make([]byte, bufferSize-3)
	if err := m.RequestInput(tooLarge); !errors.Is(err, ErrControlPayloadTooLarge) {
		t.Fatalf("oversized RequestInput error = %v, want ErrControlPayloadTooLarge", err)
	}
	if m.pendingInput.active {
		t.Fatal("rejected input left a pending request")
	}

	code := []byte{1, 2, 3, 4}
	if err := m.RequestInput(code); err != nil {
		t.Fatalf("RequestInput: %v", err)
	}
	code[0] = 0xFF // the machine keeps its own copy
	if m.pendingInput.payload[0] != 1 {
		t.Fatal("RequestInput kept a reference to the caller's slice")
	}
}

func TestSafePointHonorsPublishedControlProtocol(t *testing.T) {
	m := newSafePointMachine(t, false, srInterruptEnabled)
	put := func(offset, value uint32) {
		addr := uint32(ControlRegionStart) + offset
		binary.LittleEndian.PutUint32(m.memory[addr:addr+4], value)
	}

	if !m.atSafePointLocked() {
		t.Fatal("legacy boot with interrupts enabled must be a safe point")
	}

	put(ControlBootstrapOffset+BootstrapMagicOffset, ControlMagic)
	if m.atSafePointLocked() {
		t.Fatal("safe point before kernel status.ready")
	}

	put(ControlStatusOffset+StatusReadyOffset, 1)
	for _, phase := range []ControlPhase{PhaseLoading, PhaseInitializing, PhaseExecuting} {
		put(ControlStatusOffset+StatusPhaseOffset, uint32(phase))
		if m.atSafePointLocked() {
			t.Fatalf("safe point while guest operation is in phase %d", phase)
		}
	}

	put(ControlStatusOffset+StatusPhaseOffset, uint32(PhaseReady))
	put(ControlResultOffset+ResultValidTokenOffset, 2)
	put(ControlResultOffset+ResultAckTokenOffset, 1)
	if m.atSafePointLocked() {
		t.Fatal("safe point while terminal result awaits ack")
	}

	put(ControlResultOffset+ResultAckTokenOffset, 2)
	if !m.atSafePointLocked() {
		t.Fatal("ready protocol with acknowledged result must be a safe point")
	}

	m.registers[sr] = 0
	if m.atSafePointLocked() {
		t.Fatal("safe point with interrupts disabled")
	}
}

func TestGetRegistersReturnsDetachedCopy(t *testing.T) {
	m := NewMachine(2048, false)
	m.registers[w1] = 42

	regs := m.GetRegisters()
	regs[w1] = 7
	regs[pc] = 0x400

	if m.registers[w1] != 42 || m.registers[pc] != 0 {
		t.Fatal("GetRegisters exposed live machine state")
	}
}

func TestConcurrentStepsControlAndSnapshots(t *testing.T) {
	m := newSafePointMachine(t, true, srInterruptEnabled)
	go m.Boot()

	stop := make(chan struct{})
	done := make(chan struct{}, 2)
	go func() {
		defer func() { done <- struct{}{} }()
		for pid := uint32(0); ; pid++ {
			select {
			case <-stop:
				return
			default:
			}
			_ = m.RequestKill(pid)
			_ = m.RequestInput([]byte{byte(pid)})
		}
	}()
	go func() {
		defer func() { done <- struct{}{} }()
		for {
			select {
			case <-stop:
				return
			default:
			}
			_ = m.GetRegisters()
			_ = m.GetInstructionsExecuted()
		}
	}()

	const steps = 300
	for i := 0; i < steps; i++ {
		step(t, m)
	}
	close(stop)
	<-done
	<-done

	if got := m.GetInstructionsExecuted(); got != steps {
		t.Fatalf("instructions executed = %d, want %d (control must not consume steps)", got, steps)
	}
}
