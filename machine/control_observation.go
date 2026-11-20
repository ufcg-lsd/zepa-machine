package machine

import "encoding/binary"

// ControlObservationState describes the result of a bounded control-map read.
// Invalid means the protocol is initialized and ready, but the descriptor's
// lifecycle Valid word is zero; it is not a decode error.
type ControlObservationState uint8

const (
	ControlObservationUninitialized ControlObservationState = iota
	ControlObservationUnready
	ControlObservationInvalid
	ControlObservationMalformed
	ControlObservationInsufficientBacking
	ControlObservationReady
)

func (s ControlObservationState) String() string {
	switch s {
	case ControlObservationUninitialized:
		return "uninitialized"
	case ControlObservationUnready:
		return "unready"
	case ControlObservationInvalid:
		return "invalid"
	case ControlObservationMalformed:
		return "malformed"
	case ControlObservationInsufficientBacking:
		return "insufficient backing"
	case ControlObservationReady:
		return "ready"
	default:
		return "unknown"
	}
}

// ControlObservation is an owned value snapshot of the v1 control records.
// It contains no slices into guest memory. Result/Event availability and token
// status are observations only; neither implies a current process lifecycle.
type ControlObservation struct {
	State  ControlObservationState
	Detail string

	Bootstrap  ControlBootstrap
	Status     ControlStatus
	Descriptor ImageDescriptor
	Result     ControlResult
	Event      ControlEvent

	DescriptorTokenPresent bool
	DescriptorTokenPending bool
	ResultAvailable        bool
	ResultPending          bool
	EventAvailable         bool
	EventPending           bool
}

const controlObservationBytes uint32 = ControlEventOffset + ControlEventSize

// ReadControlObservation copies the bounded control records at their full
// guest physical address. The mutex protects only this copy against writers
// using the same mutex; it does not synchronize writers that bypass m.mu.
func (m *Machine) ReadControlObservation() ControlObservation {
	if m == nil {
		return ControlObservation{State: ControlObservationInsufficientBacking}
	}

	physicalStart := uint64(ControlRegionStart)
	physicalEnd := physicalStart + uint64(controlObservationBytes)

	m.mu.RLock()
	if physicalEnd < physicalStart || physicalEnd > uint64(len(m.memory)) {
		m.mu.RUnlock()
		return ControlObservation{State: ControlObservationInsufficientBacking}
	}

	// The bounds check above also makes both conversions safe: the end is no
	// greater than len(m.memory), which is an int-representable value.
	controlCopy := make([]byte, int(controlObservationBytes))
	copy(controlCopy, m.memory[int(physicalStart):int(physicalEnd)])
	m.mu.RUnlock()

	// Decode and validate only the private copy, after releasing the VM lock.
	return decodeControlObservation(controlCopy)
}

// decodeControlObservation decodes a compact slice whose byte zero represents
// ControlRegionStart. It deliberately does not inspect executable/image bytes.
func decodeControlObservation(data []byte) ControlObservation {
	bootstrapBytes, ok := boundedControlRecord(data, ControlBootstrapOffset, ControlBootstrapSize)
	if !ok {
		return ControlObservation{State: ControlObservationInsufficientBacking}
	}
	statusBytes, ok := boundedControlRecord(data, ControlStatusOffset, ControlStatusSize)
	if !ok {
		return ControlObservation{State: ControlObservationInsufficientBacking}
	}
	descriptorBytes, ok := boundedControlRecord(data, ControlImageDescriptorOffset, ControlImageDescriptorSize)
	if !ok {
		return ControlObservation{State: ControlObservationInsufficientBacking}
	}
	resultBytes, ok := boundedControlRecord(data, ControlResultOffset, ControlResultSize)
	if !ok {
		return ControlObservation{State: ControlObservationInsufficientBacking}
	}
	eventBytes, ok := boundedControlRecord(data, ControlEventOffset, ControlEventSize)
	if !ok {
		return ControlObservation{State: ControlObservationInsufficientBacking}
	}

	observation := ControlObservation{
		Bootstrap:  decodeControlBootstrap(bootstrapBytes),
		Status:     decodeControlStatus(statusBytes),
		Descriptor: decodeImageDescriptor(descriptorBytes),
		Result:     decodeControlResult(resultBytes),
		Event:      decodeControlEvent(eventBytes),
	}
	observation.DescriptorTokenPresent = observation.Descriptor.ValidToken != 0
	observation.DescriptorTokenPending = observation.DescriptorTokenPresent &&
		observation.Descriptor.ValidToken != observation.Descriptor.AckToken
	observation.ResultAvailable = observation.Result.ValidToken != 0
	observation.ResultPending = observation.ResultAvailable &&
		observation.Result.ValidToken != observation.Result.AckToken
	observation.EventAvailable = observation.Event.ValidToken != 0
	observation.EventPending = observation.EventAvailable &&
		observation.Event.ValidToken != observation.Event.AckToken

	if allZeroBytes(data[:int(controlObservationBytes)]) {
		observation.State = ControlObservationUninitialized
		return observation
	}
	if allZeroBytes(bootstrapBytes) {
		return malformedControlObservation(observation, "bootstrap is uninitialized while other records are populated")
	}

	if observation.Bootstrap.Magic != ControlMagic {
		return malformedControlObservation(observation, "unsupported bootstrap magic")
	}
	if observation.Bootstrap.Version != ControlProtocolVersion {
		return malformedControlObservation(observation, "unsupported bootstrap version")
	}
	if observation.Bootstrap.Capabilities != ControlCapabilitiesV1 {
		return malformedControlObservation(observation, "unsupported bootstrap capabilities")
	}
	if observation.Bootstrap.Ready > 1 {
		return malformedControlObservation(observation, "bootstrap readiness is outside its domain")
	}
	if !allZeroWords(observation.Bootstrap.Reserved[:]) {
		return malformedControlObservation(observation, "bootstrap reserved words are nonzero")
	}
	if observation.Bootstrap.Ready == 1 {
		if observation.Bootstrap.ControlBasePhysical != uint32(ControlRegionStart) ||
			observation.Bootstrap.ImageBasePhysical != uint32(ImageWindowStart) ||
			observation.Bootstrap.BufferBasePhysical != uint32(TransferBufferStart) {
			return malformedControlObservation(observation, "bootstrap physical bases do not match the shared layout")
		}
		if ValidateExecutableBootstrapLength(observation.Bootstrap.ExecutableLength) != nil {
			return malformedControlObservation(observation, "bootstrap executable length is invalid")
		}
	}

	statusUninitialized := observation.Status.Version == 0 &&
		allZeroBytes(statusBytes)
	if !statusUninitialized && observation.Status.Version != ControlProtocolVersion {
		return malformedControlObservation(observation, "unsupported status version")
	}
	if observation.Status.Ready > 1 {
		return malformedControlObservation(observation, "status readiness is outside its domain")
	}
	if observation.Status.Phase > PhaseFailed {
		return malformedControlObservation(observation, "status phase is outside its domain")
	}
	if observation.Status.Reserved != 0 || !allZeroWords(observation.Status.ReservedTail[:]) {
		return malformedControlObservation(observation, "status reserved words are nonzero")
	}
	if observation.Status.Ready == 1 && observation.Status.Phase == PhaseUninitialized {
		return malformedControlObservation(observation, "ready status has an uninitialized phase")
	}

	if observation.Descriptor.Valid > ImageDescriptorValid {
		return malformedControlObservation(observation, "descriptor validity is outside its domain")
	}
	if observation.Descriptor.Reserved != 0 {
		return malformedControlObservation(observation, "descriptor reserved word is nonzero")
	}
	if observation.Descriptor.Valid == ImageDescriptorValid {
		if observation.Descriptor.ImageID == 0 || observation.Descriptor.Generation == 0 ||
			observation.Descriptor.ValidToken == 0 {
			return malformedControlObservation(observation, "valid descriptor has unset identity, generation, or publication token")
		}
		if !validDescriptorImageMetadata(observation.Descriptor) {
			return malformedControlObservation(observation, "valid descriptor image metadata is outside the image window")
		}
	}

	if observation.Event.Reserved != 0 {
		return malformedControlObservation(observation, "event reserved word is nonzero")
	}
	if observation.ResultAvailable {
		if observation.Result.Kind < ResultLoadOK || observation.Result.Kind > ResultExecFailed ||
			observation.Result.Reason > ReasonInvalidLayout {
			return malformedControlObservation(observation, "published result kind or reason is outside its domain")
		}
	}
	if observation.EventAvailable {
		if observation.Event.Kind < EventInitCreated || observation.Event.Kind > EventExecFailed ||
			observation.Event.Reason > ReasonInvalidLayout {
			return malformedControlObservation(observation, "published event kind or reason is outside its domain")
		}
	}

	statusReady := !statusUninitialized && observation.Status.Ready == 1
	if observation.Bootstrap.Ready == 0 || !statusReady {
		observation.State = ControlObservationUnready
		return observation
	}
	if observation.Descriptor.Valid == ImageDescriptorInvalid {
		observation.State = ControlObservationInvalid
		return observation
	}
	observation.State = ControlObservationReady
	return observation
}

func boundedControlRecord(data []byte, offset, size uint32) ([]byte, bool) {
	start := uint64(offset)
	end := start + uint64(size)
	if end < start || end > uint64(len(data)) {
		return nil, false
	}
	return data[int(start):int(end)], true
}

func decodeControlBootstrap(record []byte) ControlBootstrap {
	var value ControlBootstrap
	value.Magic = controlWord(record, 0x00)
	value.Version = controlWord(record, 0x04)
	value.Capabilities = controlWord(record, 0x08)
	value.Ready = controlWord(record, 0x0C)
	value.ControlBasePhysical = controlWord(record, 0x10)
	value.ImageBasePhysical = controlWord(record, 0x14)
	value.BufferBasePhysical = controlWord(record, 0x18)
	value.ExecutableLength = controlWord(record, 0x1C)
	for i := range value.Reserved {
		value.Reserved[i] = controlWord(record, uint32(0x20+i*4))
	}
	return value
}

func decodeControlStatus(record []byte) ControlStatus {
	var value ControlStatus
	value.Version = controlWord(record, StatusVersionOffset)
	value.Phase = ControlPhase(controlWord(record, StatusPhaseOffset))
	value.Ready = controlWord(record, StatusReadyOffset)
	value.Reserved = controlWord(record, StatusReservedOffset)
	for i := range value.ReservedTail {
		value.ReservedTail[i] = controlWord(record, 0x10+uint32(i*4))
	}
	return value
}

func decodeImageDescriptor(record []byte) ImageDescriptor {
	return ImageDescriptor{
		ImageID:      ImageID(controlWord(record, DescriptorImageIDOffset)),
		Length:       controlWord(record, DescriptorLengthOffset),
		PhysicalBase: PhysicalAddress(controlWord(record, DescriptorPhysicalBaseOffset)),
		Generation:   controlWord(record, DescriptorGenerationOffset),
		ValidToken:   PublicationToken(controlWord(record, DescriptorValidTokenOffset)),
		AckToken:     PublicationToken(controlWord(record, DescriptorAckTokenOffset)),
		Valid:        controlWord(record, DescriptorValidOffset),
		Reserved:     controlWord(record, DescriptorReservedOffset),
	}
}

func decodeControlResult(record []byte) ControlResult {
	return ControlResult{
		Kind:            ResultKind(controlWord(record, ResultKindOffset)),
		Reason:          DiagnosticReason(controlWord(record, ResultReasonOffset)),
		PID:             controlWord(record, ResultPIDOffset),
		ImageID:         ImageID(controlWord(record, ResultImageIDOffset)),
		RequestID:       RequestID(controlWord(record, ResultRequestIDOffset)),
		ValidToken:      PublicationToken(controlWord(record, ResultValidTokenOffset)),
		AckToken:        PublicationToken(controlWord(record, ResultAckTokenOffset)),
		TerminationCode: controlWord(record, ResultTerminationCodeOffset),
	}
}

func decodeControlEvent(record []byte) ControlEvent {
	return ControlEvent{
		Kind:       EventKind(controlWord(record, EventKindOffset)),
		Reason:     DiagnosticReason(controlWord(record, EventReasonOffset)),
		PID:        controlWord(record, EventPIDOffset),
		ImageID:    ImageID(controlWord(record, EventImageIDOffset)),
		RequestID:  RequestID(controlWord(record, EventRequestIDOffset)),
		ValidToken: PublicationToken(controlWord(record, EventValidTokenOffset)),
		AckToken:   PublicationToken(controlWord(record, EventAckTokenOffset)),
		Reserved:   controlWord(record, EventReservedOffset),
	}
}

func controlWord(record []byte, offset uint32) uint32 {
	return binary.LittleEndian.Uint32(record[int(offset):int(offset+ControlWordBytes)])
}

func allZeroBytes(data []byte) bool {
	for _, value := range data {
		if value != 0 {
			return false
		}
	}
	return true
}

func allZeroWords(words []uint32) bool {
	for _, value := range words {
		if value != 0 {
			return false
		}
	}
	return true
}

func validDescriptorImageMetadata(descriptor ImageDescriptor) bool {
	base := uint64(descriptor.PhysicalBase)
	windowStart := uint64(ImageWindowStart)
	windowEnd := uint64(ImageWindowEnd)
	length := uint64(descriptor.Length)
	return base == windowStart && base%uint64(ControlWordBytes) == 0 &&
		length != 0 && length%uint64(ZEPAInstructionBytes) == 0 &&
		length <= uint64(MaxImagePayloadBytes) &&
		base+uint64(ImageLengthPrefixBytes)+length <= windowEnd
}

func malformedControlObservation(observation ControlObservation, detail string) ControlObservation {
	observation.State = ControlObservationMalformed
	observation.Detail = detail
	return observation
}
