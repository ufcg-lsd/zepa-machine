//go:build kernel

package main

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"
	assembler "zepa-machine/cross-assembler"
	"zepa-machine/machine"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func runTUIWithMachine(m *machine.Machine) {
	app := tview.NewApplication()

	registersView := tview.NewTextView()
	registersView.SetDynamicColors(true)
	registersView.SetScrollable(true)
	registersView.SetTitle(" Registers ")
	registersView.SetBorder(true)

	pcbView := tview.NewTextView()
	pcbView.SetDynamicColors(true)
	pcbView.SetScrollable(true)
	pcbView.SetTitle(" Kernel Variables ")
	pcbView.SetBorder(true)

	memoryView := tview.NewTextView()
	memoryView.SetDynamicColors(true)
	memoryView.SetScrollable(true)
	memoryView.SetTitle(" Memory ")
	memoryView.SetBorder(true)

	pcbVectorView := tview.NewTextView()
	pcbVectorView.SetDynamicColors(true)
	pcbVectorView.SetScrollable(true)
	pcbVectorView.SetTitle(" PCB Vector ")
	pcbVectorView.SetBorder(true)

	outputView := tview.NewTextView()
	outputView.SetDynamicColors(true)
	outputView.SetScrollable(true)
	outputView.SetTitle(" Output ")
	outputView.SetBorder(true)

	inputField := tview.NewInputField()
	inputField.SetLabel("cmd> ")
	inputField.SetFieldWidth(60)

	refreshAll := func() {
		registersView.SetText(m.DebugRegistersString())
		pcbView.SetText(m.GetKernelVarsString())
		memoryView.SetText(m.GetMemoryViewString())
		pcbVectorView.SetText(m.GetProcessTableString())
	}

	appendOutput := func(text string) {
		current := outputView.GetText(false)
		if current != "" {
			current += "\n"
		}
		current += text
		outputView.SetText(current)
		outputView.ScrollToEnd()
	}

	refreshAll()

	inputField.SetDoneFunc(func(key tcell.Key) {
		text := strings.TrimSpace(inputField.GetText())
		inputField.SetText("")

		if text == "" {
			return
		}

		parts := strings.Fields(text)

		switch parts[0] {
		case "d", "step":
			if !m.IsDebugMode() {
				appendOutput("Machine is not in debug mode!")
				return
			}
			steps := 1
			if len(parts) > 1 {
				parsedSteps, err := strconv.Atoi(parts[1])
				if err != nil || parsedSteps <= 0 {
					appendOutput("Invalid number of steps. Running 1 step.")
				} else {
					steps = parsedSteps
				}
			}
			go func() {
				for i := 0; i < steps; i++ {
					m.StepChan <- struct{}{}
					<-m.DoneChan
				}
				app.QueueUpdateDraw(func() {
					refreshAll()
				})
			}()

		case "b", "breakpoint":
			if !m.IsDebugMode() {
				appendOutput("Machine is not in debug mode!")
				return
			}
			if len(parts) < 2 {
				appendOutput("usage: b <pc>")
				return
			}
			pcValue, err := strconv.Atoi(parts[1])
			if err != nil || pcValue < 0 {
				appendOutput("Invalid PC.")
				return
			}
			go func() {
				instructionCount := 0
				for {
					m.StepChan <- struct{}{}
					<-m.DoneChan
					instructionCount++

					pc := m.GetRegisters()[10]
					if pc == uint32(pcValue) || pc == uint32(pcValue)+kernelMappingOffset {
						break
					}
				}
				app.QueueUpdateDraw(func() {
					logicalPC := m.GetRegisters()[10]
					if logicalPC >= kernelMappingOffset {
						logicalPC -= kernelMappingOffset
					}
					appendOutput(fmt.Sprintf("Breakpoint reached at pc=%d after %d instructions", logicalPC, instructionCount))
					refreshAll()
				})
			}()

		case "c", "count":
			if !m.IsDebugMode() {
				appendOutput("Machine is not in debug mode!")
				return
			}
			appendOutput(fmt.Sprintf("Instructions executed since boot: %d", m.GetInstructionsExecuted()))

		case "kill":
			if len(parts) < 2 {
				appendOutput("usage: kill <pid>")
				return
			}
			pid, err := strconv.Atoi(parts[1])
			if err != nil {
				appendOutput("invalid PID")
				return
			}
			var buf [4]byte
			binary.LittleEndian.PutUint32(buf[:], uint32(pid))
			m.LoadBuffer(buf[:])
			m.SetKillFlag()
			appendOutput(fmt.Sprintf("[sys] kill %d sent", pid))

		case "input":
			if len(parts) < 2 {
				appendOutput("usage: input <path>")
				return
			}
			code, err := assembler.RunAssembler(parts[1])
			if err != nil {
				appendOutput(fmt.Sprintf("[err] %v", err))
				return
			}
			m.LoadBuffer(code)
			m.SetInputFlag()
			appendOutput(fmt.Sprintf("[sys] input sent (%d bytes)", len(code)))

		case "reg":
			registersView.SetText(m.DebugRegistersString())

		case "pcb":
			pcbView.SetText(m.GetKernelVarsString())

		case "refresh":
			refreshAll()

		case "q", "quit":
			appendOutput("exiting debugger")
			app.Stop()
		}
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			return nil
		}
		return event
	})

	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			app.QueueUpdateDraw(func() {
				refreshAll()
			})
		}
	}()

	topRow := tview.NewFlex()
	topRow.SetDirection(tview.FlexColumn)
	topRow.AddItem(registersView, 0, 1, false)
	topRow.AddItem(pcbView, 0, 1, false)

	middleRow := tview.NewFlex()
	middleRow.SetDirection(tview.FlexColumn)
	middleRow.AddItem(memoryView, 0, 1, false)
	middleRow.AddItem(pcbVectorView, 0, 1, false)

	outputAndCmd := tview.NewFlex()
	outputAndCmd.SetDirection(tview.FlexRow)
	outputAndCmd.AddItem(outputView, 0, 1, false)
	outputAndCmd.AddItem(inputField, 3, 0, true)

	root := tview.NewFlex()
	root.SetDirection(tview.FlexRow)
	root.AddItem(topRow, 0, 3, false)
	root.AddItem(middleRow, 0, 2, false)
	root.AddItem(outputAndCmd, 5, 0, true)

	app.SetRoot(root, true)
	app.SetFocus(inputField)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
