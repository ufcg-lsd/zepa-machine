package machine

import (
	"encoding/binary"
	"errors"
)

// Host control requests (kill, input) are recorded as bounded pending state
// and applied only by the CPU owner at a safe point. The requester never
// writes guest RAM. This mirrors pending-work flags such as xv6 p->killed
// (applied in usertrap) and Linux TIF_SIGPENDING (applied in
// exit_to_user_mode_loop): mark under a lock, wake the owner, and let the
// owner apply the effect at a precise instruction boundary.

var (
	// ErrControlBusy reports that a request of the same kind is still pending.
	// The slot is bounded (one per kind) and is not a queue.
	ErrControlBusy = errors.New("machine: control request already pending")
	// ErrControlPayloadTooLarge reports a payload that cannot fit the transfer
	// buffer together with its 4-byte length prefix.
	ErrControlPayloadTooLarge = errors.New("machine: control payload exceeds transfer buffer")
)

// pendingControl is one bounded request slot holding a private payload copy.
type pendingControl struct {
	active  bool
	payload []byte
}

// RequestKill records a kill request for pid. It returns ErrControlBusy if a
// kill request is already pending.
func (m *Machine) RequestKill(pid uint32) error {
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, pid)
	return m.requestControl(&m.pendingKill, payload)
}

// RequestInput records an input request carrying already-assembled code. The
// caller assembles before calling, so the machine mutex is never held during
// assembly. It returns ErrControlPayloadTooLarge or ErrControlBusy.
func (m *Machine) RequestInput(code []byte) error {
	if len(code) > bufferSize-4 {
		return ErrControlPayloadTooLarge
	}
	return m.requestControl(&m.pendingInput, append([]byte(nil), code...))
}

func (m *Machine) requestControl(slot *pendingControl, payload []byte) error {
	m.mu.Lock()
	if slot.active {
		m.mu.Unlock()
		return ErrControlBusy
	}
	slot.active = true
	slot.payload = payload
	m.mu.Unlock()

	// State is published before the signal, so the wakeup cannot be lost.
	// The send never blocks: a token already waiting covers this request.
	select {
	case m.controlWake <- struct{}{}:
	default:
	}
	return nil
}

// serviceControlLocked applies at most one pending request if the current
// instruction boundary is a safe point. The transfer buffer is shared, so one
// request per safe point keeps payloads from overwriting each other before
// the kernel handler reads them. Priority follows the existing order:
// kill before input. The caller must hold m.mu.
func (m *Machine) serviceControlLocked() {
	if !m.atSafePointLocked() {
		return
	}

	switch {
	case m.pendingKill.active:
		m.writeTransferBufferLocked(m.pendingKill.payload)
		m.pendingKill = pendingControl{}
		m.exception(killInt)
	case m.pendingInput.active:
		m.writeTransferBufferLocked(m.pendingInput.payload)
		m.pendingInput = pendingControl{}
		m.exception(inputInt)
	}
}

// atSafePointLocked reports whether host control may change machine state at
// the current instruction boundary. It requires enabled interrupts, which also
// excludes the window between exception entry and MRET. When the host has
// published the control bootstrap, it additionally requires kernel readiness
// and no admitted guest operation awaiting its terminal result and ack. A
// legacy boot without the bootstrap has no control operations, and the kernel
// enables interrupts only after its own setup. The caller must hold m.mu.
func (m *Machine) atSafePointLocked() bool {
	if !m.isInterruptEnabled() {
		return false
	}
	if !m.controlProtocolActiveLocked() {
		return true
	}

	if m.controlWordLocked(ControlStatusOffset+StatusReadyOffset) != 1 {
		return false
	}
	switch ControlPhase(m.controlWordLocked(ControlStatusOffset + StatusPhaseOffset)) {
	case PhaseLoading, PhaseInitializing, PhaseExecuting:
		return false
	}
	resultValid := m.controlWordLocked(ControlResultOffset + ResultValidTokenOffset)
	resultAck := m.controlWordLocked(ControlResultOffset + ResultAckTokenOffset)
	return resultValid == resultAck
}

func (m *Machine) controlProtocolActiveLocked() bool {
	if uint64(len(m.memory)) < uint64(ControlRegionStart)+uint64(ControlUsedBytes) {
		return false
	}
	return m.controlWordLocked(ControlBootstrapOffset+BootstrapMagicOffset) == ControlMagic
}

// controlWordLocked reads a little-endian control word. Callers must check
// controlProtocolActiveLocked first so the access is in bounds.
func (m *Machine) controlWordLocked(offset uint32) uint32 {
	addr := uint32(ControlRegionStart) + offset
	return binary.LittleEndian.Uint32(m.memory[addr : addr+4])
}

// writeTransferBufferLocked clears the transfer buffer and writes a
// little-endian length prefix followed by payload. The caller must hold m.mu.
func (m *Machine) writeTransferBufferLocked(payload []byte) {
	const oneGB = 1 << 30
	bufferIndex := oneGB - bufferSize
	clear(m.memory[bufferIndex:oneGB])
	binary.LittleEndian.PutUint32(m.memory[bufferIndex:bufferIndex+4], uint32(len(payload)))
	copy(m.memory[bufferIndex+4:oneGB], payload)
}
