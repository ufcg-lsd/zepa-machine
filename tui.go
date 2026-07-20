

package main

import (
	"encoding/binary"
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
	pcbView.SetTitle(" Processes ")
	pcbView.SetBorder(true)

	memoryView := tview.NewTextView()
	memoryView.SetDynamicColors(true)
	memoryView.SetScrollable(true)
	memoryView.SetTitle(" Memory ")
	memoryView.SetBorder(true)

	inputField := tview.NewInputField()
	inputField.SetLabel("cmd> ")
	inputField.SetFieldWidth(60)

	refreshAll := func() {
		registersView.SetText(m.DebugRegistersString())
		pcbView.SetText(m.DebugSystemString())
		memoryView.SetText(m.GetMemoryViewString())
	}

	step := func() {
		m.StepChan <- struct{}{}
		<-m.DoneChan
		refreshAll()
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
				return
			}
			go func() {
				app.QueueUpdateDraw(func() {
					step()
				})
			}()

		case "kill":
			if len(parts) < 2 {
				return
			}
			pid, err := strconv.Atoi(parts[1])
			if err != nil {
				return
			}
			var buf [4]byte
			binary.LittleEndian.PutUint32(buf[:], uint32(pid))
			m.LoadBuffer(buf[:])
			m.SetKillFlag()

		case "input":
			if len(parts) < 2 {
				return
			}
			code, err := assembler.RunAssembler(parts[1])
			if err != nil {
				return
			}
			m.LoadBuffer(code)
			m.SetInputFlag()

		case "reg":
			registersView.SetText(m.DebugRegistersString())

		case "pcb":
			pcbView.SetText(m.DebugSystemString())

		case "refresh":
			refreshAll()
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

	// Top row: registers + processes side by side
	topRow := tview.NewFlex()
	topRow.SetDirection(tview.FlexColumn)
	topRow.AddItem(registersView, 0, 1, false)
	topRow.AddItem(pcbView, 0, 1, false)

	// Full layout: top row, memory, cmd
	root := tview.NewFlex()
	root.SetDirection(tview.FlexRow)
	root.AddItem(topRow, 0, 3, false)
	root.AddItem(memoryView, 0, 2, false)
	root.AddItem(inputField, 3, 0, true)

	app.SetRoot(root, true)
	app.SetFocus(inputField)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
