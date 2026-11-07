package machine

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestDecodeControlObservationStatesAndBounds(t *testing.T) {
	t.Run("null and truncated backing", func(t *testing.T) {
		for name, backing := range map[string][]byte{
			"nil":       nil,
			"bootstrap": make([]byte, ControlBootstrapSize-1),
			"event":     make([]byte, int(controlObservationBytes)-1),
		} {
			t.Run(name, func(t *testing.T) {
				got := decodeControlObservation(backing)
				if got.State != ControlObservationInsufficientBacking {
					t.Fatalf("state = %s, want insufficient backing", got.State)
				}
				if got.Bootstrap.Magic != 0 || got.Status.Version != 0 || got.Descriptor.Valid != 0 {
					t.Fatalf("short read returned a partial observation: %+v", got)
				}
			})
		}
	})

	t.Run("uninitialized", func(t *testing.T) {
		got := decodeControlObservation(make([]byte, int(controlObservationBytes)))
		if got.State != ControlObservationUninitialized {
			t.Fatalf("state = %s, want uninitialized", got.State)
		}
		if got.State == ControlObservationReady || got.ResultAvailable || got.EventAvailable {
			t.Fatalf("zero words were treated as initialized/published: %+v", got)
		}
	})

	t.Run("host unready", func(t *testing.T) {
		fixture := newControlObservationFixture()
		putControlWord(fixture, ControlBootstrapOffset+BootstrapReadyOffset, 0)
		got := decodeControlObservation(fixture)
		if got.State != ControlObservationUnready {
			t.Fatalf("state = %s, want unready", got.State)
		}
	})

	t.Run("kernel unready", func(t *testing.T) {
		fixture := newControlObservationFixture()
		putControlWord(fixture, ControlStatusOffset+StatusReadyOffset, 0)
		putControlWord(fixture, ControlStatusOffset+StatusPhaseOffset, uint32(PhaseUninitialized))
		got := decodeControlObservation(fixture)
		if got.State != ControlObservationUnready {
			t.Fatalf("state = %s, want unready", got.State)
		}
	})

	t.Run("descriptor invalid does not validate stale metadata", func(t *testing.T) {
		fixture := newControlObservationFixture()
		// Stale descriptor payload is ignored while Valid=0, as required by the
		// independent descriptor lifecycle-validity contract.
		putControlWord(fixture, ControlImageDescriptorOffset+DescriptorLengthOffset, 0xFFFFFFFF)
		putControlWord(fixture, ControlImageDescriptorOffset+DescriptorPhysicalBaseOffset, 0xFFFFFFFF)
		got := decodeControlObservation(fixture)
		if got.State != ControlObservationInvalid {
			t.Fatalf("state = %s, want invalid", got.State)
		}
		if got.Descriptor.Length != 0xFFFFFFFF || got.Descriptor.PhysicalBase != 0xFFFFFFFF {
			t.Fatalf("descriptor values were not copied: %+v", got.Descriptor)
		}
	})

	t.Run("malformed metadata", func(t *testing.T) {
		tests := []struct {
			name   string
			mutate func([]byte)
		}{
			{"magic", func(b []byte) { putControlWord(b, BootstrapMagicOffset, 0) }},
			{"version", func(b []byte) { putControlWord(b, BootstrapVersionOffset, 0) }},
			{"bootstrap readiness domain", func(b []byte) { putControlWord(b, BootstrapReadyOffset, 2) }},
			{"status version", func(b []byte) { putControlWord(b, ControlStatusOffset+StatusVersionOffset, 2) }},
			{"status readiness domain", func(b []byte) { putControlWord(b, ControlStatusOffset+StatusReadyOffset, 2) }},
			{"status phase domain", func(b []byte) { putControlWord(b, ControlStatusOffset+StatusPhaseOffset, uint32(PhaseFailed)+1) }},
			{"descriptor validity domain", func(b []byte) { putControlWord(b, ControlImageDescriptorOffset+DescriptorValidOffset, 2) }},
			{"valid descriptor base", func(b []byte) {
				putValidDescriptor(b, 8, uint32(ImageWindowStart)+4, 16, 1, 4, 3)
			}},
			{"valid descriptor capacity", func(b []byte) {
				putValidDescriptor(b, 8, uint32(ImageWindowStart), MaxImagePayloadBytes+1, 1, 4, 3)
			}},
			{"valid descriptor payload alignment", func(b []byte) {
				putValidDescriptor(b, 8, uint32(ImageWindowStart), 5, 1, 4, 3)
			}},
			{"valid descriptor identity", func(b []byte) {
				putValidDescriptor(b, 0, uint32(ImageWindowStart), 16, 1, 4, 3)
			}},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				fixture := newControlObservationFixture()
				test.mutate(fixture)
				got := decodeControlObservation(fixture)
				if got.State != ControlObservationMalformed {
					t.Fatalf("state = %s, want malformed; detail=%q", got.State, got.Detail)
				}
				if got.Detail == "" {
					t.Fatal("malformed observation has no detail")
				}
			})
		}
	})
}

func TestDecodeControlObservationLittleEndianAndReadOnly(t *testing.T) {
	fixture := newControlObservationFixture()
	putValidDescriptor(fixture, 0x1234, uint32(ImageWindowStart), 64, 9, 12, 11)
	putControlWord(fixture, ControlResultOffset+ResultKindOffset, uint32(ResultExecFailed))
	putControlWord(fixture, ControlResultOffset+ResultPIDOffset, 8)
	putControlWord(fixture, ControlResultOffset+ResultImageIDOffset, 0x1234)
	putControlWord(fixture, ControlResultOffset+ResultRequestIDOffset, 0x5678)
	putControlWord(fixture, ControlResultOffset+ResultValidTokenOffset, 21)
	putControlWord(fixture, ControlResultOffset+ResultAckTokenOffset, 20)
	putControlWord(fixture, ControlEventOffset+EventKindOffset, uint32(EventExecStarted))
	putControlWord(fixture, ControlEventOffset+EventPIDOffset, 8)
	putControlWord(fixture, ControlEventOffset+EventImageIDOffset, 0x2345)
	putControlWord(fixture, ControlEventOffset+EventRequestIDOffset, 0x6789)
	putControlWord(fixture, ControlEventOffset+EventValidTokenOffset, 31)
	putControlWord(fixture, ControlEventOffset+EventAckTokenOffset, 31)

	if got, want := fixture[BootstrapMagicOffset:BootstrapMagicOffset+4], []byte{0x31, 0x43, 0x32, 0x46}; !bytes.Equal(got, want) {
		t.Fatalf("little-endian magic bytes = %v, want %v", got, want)
	}
	before := append([]byte(nil), fixture...)
	got := decodeControlObservation(fixture)
	if got.State != ControlObservationReady {
		t.Fatalf("state = %s, want ready: %s", got.State, got.Detail)
	}
	if got.Bootstrap.Magic != ControlMagic || got.Bootstrap.ControlBasePhysical != uint32(ControlRegionStart) {
		t.Fatalf("bootstrap words decoded incorrectly: %+v", got.Bootstrap)
	}
	if got.Descriptor.ImageID != 0x1234 || got.Descriptor.PhysicalBase != ImageWindowStart || got.Descriptor.Length != 64 {
		t.Fatalf("descriptor words decoded incorrectly: %+v", got.Descriptor)
	}
	if !got.DescriptorTokenPending || !got.DescriptorTokenPresent {
		t.Fatalf("descriptor publication token status = present:%v pending:%v", got.DescriptorTokenPresent, got.DescriptorTokenPending)
	}
	if got.Descriptor.Valid != ImageDescriptorValid {
		t.Fatalf("descriptor Valid = %d, want 1", got.Descriptor.Valid)
	}
	if !got.ResultAvailable || !got.ResultPending || got.Result.PID != 8 ||
		got.Result.ImageID != 0x1234 || got.Result.RequestID != 0x5678 ||
		got.Result.ValidToken != 21 || got.Result.AckToken != 20 {
		t.Fatalf("result fields/token state not copied: %+v, available=%v pending=%v", got.Result, got.ResultAvailable, got.ResultPending)
	}
	if !got.EventAvailable || got.EventPending || got.Event.PID != 8 ||
		got.Event.ImageID != 0x2345 || got.Event.RequestID != 0x6789 ||
		got.Event.ValidToken != 31 || got.Event.AckToken != 31 {
		t.Fatalf("event fields/token state not copied: %+v, available=%v pending=%v", got.Event, got.EventAvailable, got.EventPending)
	}
	if !bytes.Equal(fixture, before) {
		t.Fatal("decoder modified guest control bytes")
	}
}

func TestControlObservationKeepsDescriptorValiditySeparateFromAck(t *testing.T) {
	for _, test := range []struct {
		name        string
		ack         uint32
		wantPending bool
	}{
		{name: "unacknowledged", ack: 6, wantPending: true},
		{name: "acknowledged", ack: 7, wantPending: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newControlObservationFixture()
			putValidDescriptor(fixture, 55, uint32(ImageWindowStart), 32, 4, 7, test.ack)
			got := decodeControlObservation(fixture)
			if got.State != ControlObservationReady {
				t.Fatalf("state = %s, want ready: %s", got.State, got.Detail)
			}
			if got.Descriptor.Valid != ImageDescriptorValid || !got.DescriptorTokenPresent {
				t.Fatalf("ack status changed descriptor validity: %+v", got)
			}
			if got.DescriptorTokenPending != test.wantPending {
				t.Fatalf("pending = %v, want %v", got.DescriptorTokenPending, test.wantPending)
			}
		})
	}
}

func TestControlObservationAcceptsAlignedDescriptorLengthBoundaries(t *testing.T) {
	for _, test := range []struct {
		name   string
		length uint32
	}{
		{name: "minimum instruction", length: ZEPAInstructionBytes},
		{name: "maximum payload", length: MaxImagePayloadBytes},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newControlObservationFixture()
			putValidDescriptor(fixture, 55, uint32(ImageWindowStart), test.length, 4, 7, 7)
			got := decodeControlObservation(fixture)
			if got.State != ControlObservationReady {
				t.Fatalf("length %d: state = %s, want ready: %s", test.length, got.State, got.Detail)
			}
		})
	}
}

func TestControlObservationCorrelatesSamePIDRecordsByRequestAndImage(t *testing.T) {
	fixture := newControlObservationFixture()
	putControlWord(fixture, ControlStatusOffset+StatusPhaseOffset, uint32(PhaseExecuting))
	putControlWord(fixture, ControlResultOffset+ResultKindOffset, uint32(ResultExecFailed))
	putControlWord(fixture, ControlResultOffset+ResultPIDOffset, 8)
	putControlWord(fixture, ControlResultOffset+ResultImageIDOffset, 0x100)
	putControlWord(fixture, ControlResultOffset+ResultRequestIDOffset, 0x200)
	putControlWord(fixture, ControlResultOffset+ResultValidTokenOffset, 10)
	putControlWord(fixture, ControlResultOffset+ResultAckTokenOffset, 10)
	putControlWord(fixture, ControlEventOffset+EventKindOffset, uint32(EventExecStarted))
	putControlWord(fixture, ControlEventOffset+EventPIDOffset, 8)
	putControlWord(fixture, ControlEventOffset+EventImageIDOffset, 0x101)
	putControlWord(fixture, ControlEventOffset+EventRequestIDOffset, 0x201)
	putControlWord(fixture, ControlEventOffset+EventValidTokenOffset, 12)
	putControlWord(fixture, ControlEventOffset+EventAckTokenOffset, 11)

	got := decodeControlObservation(fixture)
	if got.State != ControlObservationInvalid {
		t.Fatalf("state = %s, want invalid descriptor state", got.State)
	}
	if got.Result.PID != got.Event.PID {
		t.Fatalf("fixture PIDs differ: result=%d event=%d", got.Result.PID, got.Event.PID)
	}
	if got.Result.RequestID == got.Event.RequestID || got.Result.ImageID == got.Event.ImageID {
		t.Fatalf("same-PID records lost operation/image distinction: result=%+v event=%+v", got.Result, got.Event)
	}
	if got.Result.RequestID != 0x200 || got.Result.ImageID != 0x100 ||
		got.Event.RequestID != 0x201 || got.Event.ImageID != 0x101 ||
		got.Result.ValidToken != 10 || got.Result.AckToken != 10 ||
		got.Event.ValidToken != 12 || got.Event.AckToken != 11 {
		t.Fatalf("correlation/token fields not preserved: result=%+v event=%+v", got.Result, got.Event)
	}
}

func TestMachineControlObservationChecksPhysicalBackingWithoutLargeRAM(t *testing.T) {
	backing := make([]byte, 0x400)
	put := binary.LittleEndian
	put.PutUint32(backing[0:4], ControlMagic)
	before := append([]byte(nil), backing...)
	m := &Machine{memory: backing}
	got := m.ReadControlObservation()
	if got.State != ControlObservationInsufficientBacking {
		t.Fatalf("state = %s, want insufficient backing", got.State)
	}
	if !bytes.Equal(backing, before) {
		t.Fatal("machine wrapper modified memory while checking short backing")
	}
}

func newControlObservationFixture() []byte {
	fixture := make([]byte, int(controlObservationBytes))
	putControlWord(fixture, BootstrapMagicOffset, ControlMagic)
	putControlWord(fixture, BootstrapVersionOffset, ControlProtocolVersion)
	putControlWord(fixture, BootstrapCapabilitiesOffset, ControlCapabilitiesV1)
	putControlWord(fixture, BootstrapReadyOffset, 1)
	putControlWord(fixture, BootstrapControlPhysicalOffset, uint32(ControlRegionStart))
	putControlWord(fixture, BootstrapImagePhysicalOffset, uint32(ImageWindowStart))
	putControlWord(fixture, BootstrapBufferPhysicalOffset, uint32(TransferBufferStart))
	putControlWord(fixture, BootstrapExecutableLengthOffset, ZEPAInstructionBytes)
	putControlWord(fixture, ControlStatusOffset+StatusVersionOffset, ControlProtocolVersion)
	putControlWord(fixture, ControlStatusOffset+StatusPhaseOffset, uint32(PhaseReady))
	putControlWord(fixture, ControlStatusOffset+StatusReadyOffset, 1)
	return fixture
}

func putValidDescriptor(fixture []byte, imageID, physicalBase, length, generation, validToken, ackToken uint32) {
	putControlWord(fixture, ControlImageDescriptorOffset+DescriptorImageIDOffset, imageID)
	putControlWord(fixture, ControlImageDescriptorOffset+DescriptorLengthOffset, length)
	putControlWord(fixture, ControlImageDescriptorOffset+DescriptorPhysicalBaseOffset, physicalBase)
	putControlWord(fixture, ControlImageDescriptorOffset+DescriptorGenerationOffset, generation)
	putControlWord(fixture, ControlImageDescriptorOffset+DescriptorValidTokenOffset, validToken)
	putControlWord(fixture, ControlImageDescriptorOffset+DescriptorAckTokenOffset, ackToken)
	putControlWord(fixture, ControlImageDescriptorOffset+DescriptorValidOffset, ImageDescriptorValid)
}

func putControlWord(fixture []byte, offset, value uint32) {
	binary.LittleEndian.PutUint32(fixture[int(offset):int(offset+ControlWordBytes)], value)
}
