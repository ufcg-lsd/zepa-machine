//go:build kernel

package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"zepa-machine/machine"
)

type executionCommandKind uint8

const (
	executionStep executionCommandKind = iota
	executionPlay
	executionBreakpoint
	executionStop
	executionPause
)

type executionCommand struct {
	kind   executionCommandKind
	steps  int
	target uint32
	reason string
}

type executionNotice struct {
	message string
}

type instructionResult struct {
	pc     uint32
	cycles uint64
}

type instructionStepper interface {
	Step(context.Context) (instructionResult, error)
}

type executionTicker interface {
	Ticks() <-chan time.Time
	Stop()
}

type executionTickerFactory func(time.Duration) executionTicker

type realExecutionTicker struct {
	ticker *time.Ticker
}

func (t realExecutionTicker) Ticks() <-chan time.Time { return t.ticker.C }
func (t realExecutionTicker) Stop()                   { t.ticker.Stop() }

type executionCoordinator struct {
	stepper instructionStepper
	newTick executionTickerFactory
	ctx     context.Context
	cancel  context.CancelFunc

	queueMu sync.Mutex
	queue   []executionCommand
	wake    chan struct{}
	notices chan executionNotice
	done    chan struct{}
}

func newExecutionCoordinator(stepper instructionStepper) *executionCoordinator {
	return newExecutionCoordinatorWithTicker(stepper, func(interval time.Duration) executionTicker {
		return realExecutionTicker{ticker: time.NewTicker(interval)}
	})
}

func newExecutionCoordinatorWithTicker(stepper instructionStepper, newTick executionTickerFactory) *executionCoordinator {
	ctx, cancel := context.WithCancel(context.Background())
	return &executionCoordinator{
		stepper: stepper,
		newTick: newTick,
		ctx:     ctx,
		cancel:  cancel,
		wake:    make(chan struct{}, 1),
		notices: make(chan executionNotice, 32),
		done:    make(chan struct{}),
	}
}

func (c *executionCoordinator) Start() { go c.run() }

// Submit only queues intent; TUI callbacks never wait for the emulated CPU.
func (c *executionCoordinator) Submit(command executionCommand) bool {
	select {
	case <-c.ctx.Done():
		return false
	default:
	}

	c.queueMu.Lock()
	c.queue = append(c.queue, command)
	c.queueMu.Unlock()
	select {
	case c.wake <- struct{}{}:
	default:
	}
	return true
}

func (c *executionCoordinator) Notices() <-chan executionNotice { return c.notices }
func (c *executionCoordinator) Done() <-chan struct{}           { return c.done }
func (c *executionCoordinator) Close()                          { c.cancel() }

func (c *executionCoordinator) takeCommands() []executionCommand {
	c.queueMu.Lock()
	commands := c.queue
	c.queue = nil
	c.queueMu.Unlock()
	return commands
}

func (c *executionCoordinator) hasCommands() bool {
	c.queueMu.Lock()
	defer c.queueMu.Unlock()
	return len(c.queue) != 0
}

// beginStep linearizes starting an instruction against Submit. A control
// request submitted after this point treats the reserved instruction as the
// one in flight and is handled after its acknowledgement.
func (c *executionCoordinator) beginStep() bool {
	c.queueMu.Lock()
	defer c.queueMu.Unlock()
	if len(c.queue) != 0 {
		return false
	}
	return true
}

func (c *executionCoordinator) notify(message string) {
	select {
	case c.notices <- executionNotice{message: message}:
	case <-c.ctx.Done():
	}
}

func (c *executionCoordinator) run() {
	defer close(c.done)
	defer close(c.notices)

	var active *executionCommand
	var remaining, instructionCount int
	var cycleCount uint64
	var paused bool
	var pending []executionCommand
	var ticker executionTicker
	var tick <-chan time.Time

	stopTicker := func() {
		if ticker != nil {
			ticker.Stop()
			ticker = nil
			tick = nil
		}
	}
	setActive := func(command executionCommand) {
		stopTicker()
		active = &command
		instructionCount = 0
		cycleCount = 0
		remaining = command.steps
		if command.kind == executionStep && remaining < 1 {
			remaining = 1
		}
		if command.kind == executionPlay {
			ticker = c.newTick(100 * time.Millisecond)
			tick = ticker.Ticks()
			c.notify("play started (100ms/step)")
		}
	}
	clearActive := func() {
		stopTicker()
		active = nil
	}
	startPending := func() {
		if len(pending) == 0 {
			return
		}
		command := pending[0]
		pending = pending[1:]
		setActive(command)
	}
	process := func(commands []executionCommand) {
		for _, command := range commands {
			switch command.kind {
			case executionPause:
				clearActive()
				paused = true
				reason := command.reason
				if reason == "" {
					reason = "external request"
				}
				c.notify(fmt.Sprintf("Execution paused (%s); in-flight instruction acknowledged.", reason))
			case executionStop:
				if active == nil {
					c.notify("Not playing")
					continue
				}
				clearActive()
				c.notify("Execution stopped after in-flight instruction was acknowledged.")
			case executionStep, executionPlay, executionBreakpoint:
				if paused {
					pending = append(pending, command)
					c.notify("Execution request pending while paused")
					if command.kind == executionStep || command.kind == executionPlay {
						paused = false
						c.notify("Execution resumed")
						startPending()
					}
					continue
				}
				if active != nil && active.kind == executionPlay && command.kind == executionPlay {
					c.notify("Already playing (use 'stop')")
					continue
				}
				setActive(command)
			}
		}
	}

	for {
		select {
		case <-c.ctx.Done():
			stopTicker()
			return
		default:
		}

		process(c.takeCommands())
		if c.ctx.Err() != nil {
			stopTicker()
			return
		}
		if !paused && active == nil {
			startPending()
		}

		if paused || active == nil {
			select {
			case <-c.ctx.Done():
				stopTicker()
				return
			case <-c.wake:
			}
			continue
		}

		if active.kind == executionPlay {
			select {
			case <-c.ctx.Done():
				stopTicker()
				return
			case <-c.wake:
				continue
			case <-tick:
			}
			// A control request that arrived with the tick wins before a new step.
			if c.hasCommands() {
				continue
			}
		} else if c.hasCommands() {
			continue
		}
		if !c.beginStep() {
			continue
		}

		result, err := c.stepper.Step(c.ctx)
		if err != nil {
			clearActive()
			if c.ctx.Err() == nil {
				c.notify(fmt.Sprintf("[err] execution step: %v", err))
			}
			continue
		}
		instructionCount++
		cycleCount += result.cycles

		switch active.kind {
		case executionStep:
			remaining--
			if remaining == 0 {
				clearActive()
				c.notify(fmt.Sprintf("Cycles used: %d", cycleCount))
			}
		case executionBreakpoint:
			if result.pc == active.target || result.pc == active.target+kernelMappingOffset {
				logicalPC := result.pc
				if logicalPC >= kernelMappingOffset {
					logicalPC -= kernelMappingOffset
				}
				clearActive()
				c.notify(fmt.Sprintf("Breakpoint reached at pc=%d after %d instructions", logicalPC, instructionCount))
				c.notify(fmt.Sprintf("Cycles used: %d", cycleCount))
			}
		}
	}
}

type machineInstructionStepper struct {
	machine *machine.Machine
}

func (s machineInstructionStepper) Step(ctx context.Context) (instructionResult, error) {
	cyclesBefore := s.machine.GetCyclesExecuted()
	select {
	case <-ctx.Done():
		return instructionResult{}, ctx.Err()
	case s.machine.StepChan <- struct{}{}:
		// Once accepted by Boot, this instruction is in flight and must be
		// acknowledged even if cancellation is requested meanwhile.
	}
	<-s.machine.DoneChan
	registers := s.machine.GetRegistersSnapshot()
	return instructionResult{
		pc:     registers[machine.Register(10)],
		cycles: s.machine.GetCyclesExecuted() - cyclesBefore,
	}, nil
}
