//go:build kernel

package main

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type controlledInstructionStepper struct {
	started       chan chan instructionResult
	active        atomic.Int32
	maxConcurrent atomic.Int32
}

func newControlledInstructionStepper() *controlledInstructionStepper {
	return &controlledInstructionStepper{started: make(chan chan instructionResult, 8)}
}

func (s *controlledInstructionStepper) Step(ctx context.Context) (instructionResult, error) {
	active := s.active.Add(1)
	for {
		old := s.maxConcurrent.Load()
		if active <= old || s.maxConcurrent.CompareAndSwap(old, active) {
			break
		}
	}
	defer s.active.Add(-1)
	ack := make(chan instructionResult, 1)
	select {
	case s.started <- ack:
	case <-ctx.Done():
		return instructionResult{}, ctx.Err()
	}
	// A sent step is in flight; model Boot by waiting for its completion even
	// when the coordinator is cancelled after accepting it.
	return <-ack, nil
}

type controlledExecutionTicker struct {
	ticks chan time.Time
	stops atomic.Int32
}

func (t *controlledExecutionTicker) Ticks() <-chan time.Time { return t.ticks }
func (t *controlledExecutionTicker) Stop()                   { t.stops.Add(1) }

func startTestCoordinator(t *testing.T, stepper instructionStepper, newTick executionTickerFactory) *executionCoordinator {
	t.Helper()
	coordinator := newExecutionCoordinatorWithTicker(stepper, newTick)
	coordinator.Start()
	t.Cleanup(func() {
		coordinator.Close()
		select {
		case <-coordinator.Done():
		case <-time.After(time.Second):
			t.Error("execution coordinator did not stop")
		}
	})
	return coordinator
}

func nextStep(t *testing.T, stepper *controlledInstructionStepper) chan instructionResult {
	t.Helper()
	select {
	case ack := <-stepper.started:
		return ack
	case <-time.After(time.Second):
		t.Fatal("coordinator did not emit a step")
		return nil
	}
}

func nextNotice(t *testing.T, coordinator *executionCoordinator, contains string) executionNotice {
	t.Helper()
	deadline := time.After(time.Second)
	for {
		select {
		case notice, ok := <-coordinator.Notices():
			if !ok {
				t.Fatalf("notices closed while waiting for %q", contains)
			}
			if strings.Contains(notice.message, contains) {
				return notice
			}
		case <-deadline:
			t.Fatalf("no coordinator notice contained %q", contains)
		}
	}
}

func assertNoStep(t *testing.T, stepper *controlledInstructionStepper) {
	t.Helper()
	select {
	case <-stepper.started:
		t.Fatal("coordinator emitted a step while paused or before the prior step was acknowledged")
	default:
	}
}

func TestExecutionCoordinatorCountsAcknowledgedSteps(t *testing.T) {
	stepper := newControlledInstructionStepper()
	coordinator := startTestCoordinator(t, stepper, func(time.Duration) executionTicker {
		t.Fatal("step mode must not create a play ticker")
		return nil
	})
	if !coordinator.Submit(executionCommand{kind: executionStep, steps: 2}) {
		t.Fatal("submit step request")
	}
	first := nextStep(t, stepper)
	assertNoStep(t, stepper)
	first <- instructionResult{pc: 4, cycles: 3}
	second := nextStep(t, stepper)
	assertNoStep(t, stepper)
	second <- instructionResult{pc: 8, cycles: 4}
	notice := nextNotice(t, coordinator, "Cycles used: 7")
	if !strings.Contains(notice.message, "7") {
		t.Fatalf("unexpected step completion notice: %q", notice.message)
	}
	assertNoStep(t, stepper)
	if got := stepper.maxConcurrent.Load(); got != 1 {
		t.Fatalf("concurrent instruction steps = %d, want 1", got)
	}
}

func TestExecutionCoordinatorPauseDrainsInflightStepAndResumes(t *testing.T) {
	stepper := newControlledInstructionStepper()
	coordinator := startTestCoordinator(t, stepper, func(time.Duration) executionTicker {
		t.Fatal("breakpoint/step mode must not create a play ticker")
		return nil
	})
	coordinator.Submit(executionCommand{kind: executionBreakpoint, target: 20})
	inflight := nextStep(t, stepper)
	coordinator.Submit(executionCommand{kind: executionPause, reason: "init-created"})
	inflight <- instructionResult{pc: 8, cycles: 2}
	nextNotice(t, coordinator, "init-created")
	assertNoStep(t, stepper)

	coordinator.Submit(executionCommand{kind: executionBreakpoint, target: 20})
	nextNotice(t, coordinator, "pending while paused")
	assertNoStep(t, stepper)

	coordinator.Submit(executionCommand{kind: executionStep, steps: 1})
	nextNotice(t, coordinator, "Execution resumed")
	resumedBreakpoint := nextStep(t, stepper)
	resumedBreakpoint <- instructionResult{pc: 20, cycles: 5}
	nextNotice(t, coordinator, "Breakpoint reached at pc=20 after 1 instructions")
	resumedStep := nextStep(t, stepper)
	resumedStep <- instructionResult{pc: 24, cycles: 1}
	nextNotice(t, coordinator, "Cycles used: 1")
	assertNoStep(t, stepper)
	if got := stepper.maxConcurrent.Load(); got != 1 {
		t.Fatalf("concurrent instruction steps = %d, want 1", got)
	}
}

func TestExecutionCoordinatorPlayStopAcknowledgesInflightStep(t *testing.T) {
	stepper := newControlledInstructionStepper()
	tickers := make(chan *controlledExecutionTicker, 1)
	coordinator := startTestCoordinator(t, stepper, func(time.Duration) executionTicker {
		ticker := &controlledExecutionTicker{ticks: make(chan time.Time, 1)}
		tickers <- ticker
		return ticker
	})
	coordinator.Submit(executionCommand{kind: executionPlay})
	nextNotice(t, coordinator, "play started")
	ticker := <-tickers
	ticker.ticks <- time.Time{}
	inflight := nextStep(t, stepper)
	coordinator.Submit(executionCommand{kind: executionStop})
	inflight <- instructionResult{pc: 4, cycles: 1}
	nextNotice(t, coordinator, "Execution stopped")
	assertNoStep(t, stepper)
	if got := ticker.stops.Load(); got != 1 {
		t.Fatalf("play ticker stop count = %d, want 1", got)
	}
	if got := stepper.maxConcurrent.Load(); got != 1 {
		t.Fatalf("concurrent instruction steps = %d, want 1", got)
	}
}

func TestExecutionCoordinatorCloseWaitsForInflightAcknowledgement(t *testing.T) {
	stepper := newControlledInstructionStepper()
	coordinator := startTestCoordinator(t, stepper, func(time.Duration) executionTicker {
		t.Fatal("breakpoint mode must not create a play ticker")
		return nil
	})
	coordinator.Submit(executionCommand{kind: executionBreakpoint, target: 20})
	inflight := nextStep(t, stepper)

	coordinator.Close()
	assertNoStep(t, stepper)
	inflight <- instructionResult{pc: 8, cycles: 2}

	select {
	case <-coordinator.Done():
	case <-time.After(time.Second):
		t.Fatal("coordinator did not exit after the in-flight step was acknowledged")
	}
	assertNoStep(t, stepper)
	if got := stepper.maxConcurrent.Load(); got != 1 {
		t.Fatalf("concurrent instruction steps = %d, want 1", got)
	}
}
