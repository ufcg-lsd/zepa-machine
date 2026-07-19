package machine

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func readUint32(memory []byte, addr uint32) uint32 {
	return uint32(memory[addr]) |
		uint32(memory[addr+1])<<8 |
		uint32(memory[addr+2])<<16 |
		uint32(memory[addr+3])<<24
}

func getRegisterName(reg Register) string {
	switch reg {
	case 0:
		return "w0"
	case 1:
		return "w1"
	case 2:
		return "w2"
	case 3:
		return "w3"
	case 4:
		return "w4"
	case 5:
		return "w5"
	case 6:
		return "w6"
	case 7:
		return "w7"
	case 8:
		return "w8"
	case 9:
		return "w9"
	case 10:
		return "pc"
	case 11:
		return "sp"
	case 12:
		return "ir"
	case 13:
		return "sr"
	case 14:
		return "mdr"
	case 15:
		return "mar"
	case 16:
		return "ecr"
	case 17:
		return "esa"
	case 18:
		return "esr"
	case 19:
		return "epc"
	case 20:
		return "base"
	case 21:
		return "limit"
	default:
		return "invalid"
	}
}

func srFlags(sr uint32) string {
	var s string

	if sr&1 != 0 {
		s += "Z"
	} else {
		s += "."
	}

	if sr&2 != 0 {
		s += "L"
	} else {
		s += "."
	}

	if sr&4 != 0 {
		s += "G"
	} else {
		s += "."
	}

	if sr&8 != 0 {
		s += "U"
	} else {
		s += "K"
	}

	if sr&16 != 0 {
		s += "I"
	} else {
		s += "."
	}

	return s
}

func exceptionName(ecr uint32) string {
	switch ecr {
	case 0:
		return "clock"
	case 1:
		return "input"
	case 2:
		return "kill"
	case 3:
		return "syscall"
	case 4:
		return "fault"
	default:
		return "unknown"
	}
}

func pcbState(flags byte) string {
	if flags&2 != 0 {
		return "zombie"
	}

	if flags&4 != 0 {
		return "blocked"
	}

	state := (flags >> 3) & 3

	switch state {
	case 0:
		return "ready"
	case 1:
		return "running"
	case 2:
		return "blocked"
	default:
		return "unknown"
	}
}

func getPcbBase(pid uint32) uint32 {
	return 0x102C + pid*84
}

func visualLen(s string) int {
	return utf8.RuneCountInString(s)
}

func truncateRunes(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	runes := []rune(s)

	if len(runes) <= maxWidth {
		return s
	}

	return string(runes[:maxWidth])
}

func maxLen(strs ...string) int {
	maximum := 0

	for _, s := range strs {
		length := visualLen(s)

		if length > maximum {
			maximum = length
		}
	}

	return maximum
}

func maxLenMin(minimum int, strs ...string) int {
	maximum := minimum

	for _, s := range strs {
		length := visualLen(s)

		if length > maximum {
			maximum = length
		}
	}

	return maximum
}

func maxLenIntMin(minimum int, vals ...int32) int {
	maximum := minimum

	for _, value := range vals {
		s := fmt.Sprintf("%d", value)

		if len(s) > maximum {
			maximum = len(s)
		}
	}

	return maximum
}

func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}

	return strings.Repeat(s, n)
}

func hLine(left, mid, right string, widths []int) string {
	var builder strings.Builder

	builder.WriteString(left)

	for i, width := range widths {
		builder.WriteString(repeat("─", width+2))

		if i < len(widths)-1 {
			builder.WriteString(mid)
		}
	}

	builder.WriteString(right)

	return builder.String()
}

func tableWidth(widths []int) int {
	if len(widths) == 0 {
		return 2
	}


	width := 2 + len(widths) - 1

	for _, columnWidth := range widths {
		width += columnWidth + 2
	}

	return width
}

func titleLine(width int, title string) string {
	const horizontalPadding = 2

	innerWidth := width - 2

	if innerWidth <= 0 {
		return "││"
	}

	titleWidth := innerWidth - 2*horizontalPadding

	if titleWidth < 0 {
		titleWidth = 0
	}

	title = truncateRunes(title, titleWidth)

	padding := titleWidth - visualLen(title)

	if padding < 0 {
		padding = 0
	}

	leftPadding := padding / 2
	rightPadding := padding - leftPadding

	return "│" +
		repeat(" ", horizontalPadding+leftPadding) +
		title +
		repeat(" ", horizontalPadding+rightPadding) +
		"│"
}

func (m *Machine) DebugRegisters() {
	registers := m.registers

	registerOrder := []Register{
		w0, w1, w2, w3, w4, w5, w6, w7, w8, w9,
		pc, sp, ir, sr, mdr, mar, ecr, esa, esr, epc,
		base, limit,
	}

	type cell struct {
		name string
		dec  int32
		hex  string
	}

	var pairs [][2]cell

	for i := 0; i < len(registerOrder); i += 2 {
		firstRegister := registerOrder[i]
		firstValue := registers[firstRegister]
		firstName := getRegisterName(firstRegister)
		firstDisplay := fmt.Sprintf("0x%08X", firstValue)

		if firstRegister == sr {
			firstDisplay = srFlags(firstValue)
		} else if firstRegister == ecr {
			firstDisplay = exceptionName(firstValue)
		}

		firstCell := cell{
			name: firstName,
			dec:  int32(firstValue),
			hex:  firstDisplay,
		}

		if i+1 < len(registerOrder) {
			secondRegister := registerOrder[i+1]
			secondValue := registers[secondRegister]
			secondName := getRegisterName(secondRegister)
			secondDisplay := fmt.Sprintf("0x%08X", secondValue)

			if secondRegister == sr {
				secondDisplay = srFlags(secondValue)
			} else if secondRegister == ecr {
				secondDisplay = exceptionName(secondValue)
			}

			secondCell := cell{
				name: secondName,
				dec:  int32(secondValue),
				hex:  secondDisplay,
			}

			pairs = append(pairs, [2]cell{
				firstCell,
				secondCell,
			})
		} else {
			pairs = append(pairs, [2]cell{
				firstCell,
				{},
			})
		}
	}

	nameWidth := maxLenMin(2, "REG")
	decimalWidth := 1
	hexWidth := maxLenMin(3, "HEX")

	for _, pair := range pairs {
		nameWidth = maxLenMin(
			nameWidth,
			pair[0].name,
			pair[1].name,
			"REG",
		)

		decimalWidth = maxLenIntMin(
			decimalWidth,
			pair[0].dec,
			pair[1].dec,
			0,
		)

		hexWidth = maxLenMin(
			hexWidth,
			pair[0].hex,
			pair[1].hex,
			"HEX",
		)
	}

	columnWidths := []int{
		nameWidth,
		decimalWidth,
		hexWidth,
		nameWidth,
		decimalWidth,
		hexWidth,
	}

	totalWidth := tableWidth(columnWidths)

	fmt.Println("┌" + repeat("─", totalWidth-2) + "┐")
	fmt.Println(titleLine(totalWidth, "CPU REGISTERS"))
	fmt.Println(hLine("├", "┬", "┤", columnWidths))

	rowFormat := fmt.Sprintf(
		"│ %%-%ds │ %%%dd │ %%-%ds │ %%-%ds │ %%%dd │ %%-%ds │\n",
		nameWidth,
		decimalWidth,
		hexWidth,
		nameWidth,
		decimalWidth,
		hexWidth,
	)

	headerFormat := fmt.Sprintf(
		"│ %%-%ds │ %%%ds │ %%-%ds │ %%-%ds │ %%%ds │ %%-%ds │\n",
		nameWidth,
		decimalWidth,
		hexWidth,
		nameWidth,
		decimalWidth,
		hexWidth,
	)

	fmt.Printf(
		headerFormat,
		"REG",
		"DEC",
		"HEX",
		"REG",
		"DEC",
		"HEX",
	)

	fmt.Println(hLine("├", "┼", "┤", columnWidths))

	srLine := ""
	ecrLine := ""

	for _, pair := range pairs {
		for _, currentCell := range pair {
			switch currentCell.name {
			case "sr":
				srLine = fmt.Sprintf(
					"sr[Z L G U I] = %s",
					currentCell.hex,
				)

			case "ecr":
				ecrLine = fmt.Sprintf(
					"ecr = %s",
					currentCell.hex,
				)
			}
		}

		if pair[1].name == "" {
			fmt.Printf(
				rowFormat,
				pair[0].name,
				pair[0].dec,
				pair[0].hex,
				"",
				0,
				"",
			)

			continue
		}

		fmt.Printf(
			rowFormat,
			pair[0].name,
			pair[0].dec,
			pair[0].hex,
			pair[1].name,
			pair[1].dec,
			pair[1].hex,
		)
	}

	fmt.Println(hLine("└", "┴", "┘", columnWidths))

	if srLine != "" || ecrLine != "" {
		fmt.Printf("  %s", srLine)

		if srLine != "" && ecrLine != "" {
			fmt.Print("    ")
		}

		fmt.Println(ecrLine)
	}

	fmt.Println()
}

func runningPidStr(runningPid uint32) string {
	if runningPid == 0xFFFFFFFF {
		return "idle"
	}

	return fmt.Sprintf("%d", runningPid)
}

func (m *Machine) DebugSystem() {
	memory := m.memory

	type keyValue struct {
		name  string
		value uint32
		extra string
	}

	partitionSize := readUint32(memory, 0x1000)
	timeSlice := readUint32(memory, 0x1004)
	kernelMaxMemory := readUint32(memory, 0x1008)
	bufferSize := readUint32(memory, 0x100C)
	memorySize := readUint32(memory, 0x1010)
	partitionNumber := readUint32(memory, 0x1014)
	runningPid := readUint32(memory, 0x1018)
	clockInterruptCount := readUint32(memory, 0x101C)
	kernelStackPointer := readUint32(memory, 0x1020)
	scratchSpace0 := readUint32(memory, 0x1024)
	scratchSpace1 := readUint32(memory, 0x1028)
	pcbVector := readUint32(memory, 0x102C)

	const pcbSize uint32 = 84

	runningPidExtra := ""

	if runningPid == 0xFFFFFFFF {
		runningPidExtra = "(idle)"
	}

	totalPcbMemoryExtra := fmt.Sprintf(
		"(%d x %d)",
		pcbSize,
		partitionNumber,
	)

	allRows := []keyValue{
		{"PARTITION_SIZE", partitionSize, ""},
		{"TIME_SLICE", timeSlice, ""},
		{"KERNEL_MAX_MEMORY", kernelMaxMemory, ""},
		{"BUFFER_SIZE", bufferSize, ""},

		{"memory_size", memorySize, ""},
		{"partition_number", partitionNumber, ""},
		{"running_pid", runningPid, runningPidExtra},
		{"clock_interrupt_count", clockInterruptCount, ""},
		{"kernel_stack_pointer", kernelStackPointer, ""},
		{"scratch_space_0", scratchSpace0, ""},
		{"scratch_space_1", scratchSpace1, ""},

		{"pcb_vector (base)", pcbVector, ""},
		{"  pcb_size", pcbSize, ""},
		{
			"  total_pcb_memory",
			pcbSize * partitionNumber,
			totalPcbMemoryExtra,
		},
	}

	nameWidth := visualLen("Variable")
	decimalWidth := visualLen("Value")
	hexWidth := visualLen("Hex")

	for _, row := range allRows {
		nameWidth = maxLenMin(nameWidth, row.name)

		decimalValue := fmt.Sprintf("%d", int32(row.value))
		decimalWidth = maxLenMin(decimalWidth, decimalValue)

		hexValue := fmt.Sprintf("0x%08X", row.value)

		if row.extra != "" {
			hexValue += " " + row.extra
		}

		hexWidth = maxLenMin(hexWidth, hexValue)
	}

	columnWidths := []int{
		nameWidth,
		decimalWidth,
		hexWidth,
	}

	totalWidth := tableWidth(columnWidths)

	fmt.Println("┌" + repeat("─", totalWidth-2) + "┐")
	fmt.Println(titleLine(totalWidth, "KERNEL VARIABLES"))
	fmt.Println(hLine("├", "┬", "┤", columnWidths))

	rowFormat := fmt.Sprintf(
		"│ %%-%ds │ %%%dd │ %%-%ds │\n",
		nameWidth,
		decimalWidth,
		hexWidth,
	)

	headerFormat := fmt.Sprintf(
		"│ %%-%ds │ %%%ds │ %%-%ds │\n",
		nameWidth,
		decimalWidth,
		hexWidth,
	)

	emitRow := func(name string, value uint32, extra string) {
		hexValue := fmt.Sprintf("0x%08X", value)

		if extra != "" {
			hexValue += " " + extra
		}

		fmt.Printf(
			rowFormat,
			name,
			int32(value),
			hexValue,
		)
	}

	emitSection := func(label string) {
		fmt.Printf(
			"│ %-*s │ %*s │ %-*s │\n",
			nameWidth,
			label,
			decimalWidth,
			"",
			hexWidth,
			"",
		)
	}

	fmt.Printf(
		headerFormat,
		"Variable",
		"Value",
		"Hex",
	)

	fmt.Println(hLine("├", "┼", "┤", columnWidths))

	emitSection("[Constants]")

	for i := 0; i < 4; i++ {
		row := allRows[i]

		emitRow(
			row.name,
			row.value,
			row.extra,
		)
	}

	fmt.Println(hLine("├", "┼", "┤", columnWidths))

	emitSection("[Singular Values]")

	for i := 4; i < 11; i++ {
		row := allRows[i]

		emitRow(
			row.name,
			row.value,
			row.extra,
		)
	}

	fmt.Println(hLine("├", "┼", "┤", columnWidths))

	emitSection("[Data Structures]")

	for i := 11; i < len(allRows); i++ {
		row := allRows[i]

		emitRow(
			row.name,
			row.value,
			row.extra,
		)
	}

	fmt.Println(hLine("└", "┴", "┘", columnWidths))
	fmt.Println()


	if partitionNumber == 0 {
		fmt.Println("Nenhum processo (partition_number = 0).")
		fmt.Println()
		return
	}

	type pcbRow struct {
		pidCell string
		state   string
		parent  int32
		regs    [13]int32
	}

	var processes []pcbRow

	for pid := uint32(0); pid < partitionNumber; pid++ {
		pcbAddress := getPcbBase(pid)
		flags := memory[pcbAddress+80]


		if flags&1 == 0 {
			continue
		}

		parentPid := int32(
			readUint32(memory, pcbAddress+0),
		)

		registerValues := [13]int32{
			int32(readUint32(memory, pcbAddress+20)),
			int32(readUint32(memory, pcbAddress+24)),
			int32(readUint32(memory, pcbAddress+28)),
			int32(readUint32(memory, pcbAddress+32)),
			int32(readUint32(memory, pcbAddress+36)),
			int32(readUint32(memory, pcbAddress+40)),
			int32(readUint32(memory, pcbAddress+44)),
			int32(readUint32(memory, pcbAddress+48)),
			int32(readUint32(memory, pcbAddress+52)),
			int32(readUint32(memory, pcbAddress+56)),
			int32(readUint32(memory, pcbAddress+60)),
			int32(readUint32(memory, pcbAddress+64)),
			int32(readUint32(memory, pcbAddress+68)),
		}

		marker := " "

		if pid == runningPid {
			marker = ">"
		}

		processes = append(processes, pcbRow{
			pidCell: fmt.Sprintf("%s%d", marker, pid),
			state:   pcbState(flags),
			parent:  parentPid,
			regs:    registerValues,
		})
	}

	columnHeaders := []string{
		"PID",
		"State",
		"Parent",
		"W0",
		"W1",
		"W2",
		"W3",
		"W4",
		"W5",
		"W6",
		"W7",
		"W8",
		"W9",
		"PC",
		"SP",
		"SR",
	}

	columnWidths = make([]int, len(columnHeaders))

	for i, header := range columnHeaders {
		columnWidths[i] = visualLen(header)
	}

	for _, process := range processes {
		columnWidths[0] = maxLenMin(
			columnWidths[0],
			process.pidCell,
		)

		columnWidths[1] = maxLenMin(
			columnWidths[1],
			process.state,
		)

		columnWidths[2] = maxLenMin(
			columnWidths[2],
			fmt.Sprintf("%d", process.parent),
		)

		for registerIndex, registerValue := range process.regs {
			columnIndex := registerIndex + 3

			columnWidths[columnIndex] = maxLenMin(
				columnWidths[columnIndex],
				fmt.Sprintf("%d", registerValue),
			)
		}
	}

	processTableWidth := tableWidth(columnWidths)

	fmt.Println(
		"┌" +
			repeat("─", processTableWidth-2) +
			"┐",
	)

	fmt.Println(
		titleLine(
			processTableWidth,
			fmt.Sprintf(
				"PROCESS TABLE  •  running_pid: %s",
				runningPidStr(runningPid),
			),
		),
	)

	fmt.Println(
		hLine(
			"├",
			"┬",
			"┤",
			columnWidths,
		),
	)

	var headerFormatBuilder strings.Builder

	headerFormatBuilder.WriteString("│")

	for _, width := range columnWidths {
		headerFormatBuilder.WriteString(
			fmt.Sprintf(" %%-%ds │", width),
		)
	}

	headerFormatBuilder.WriteString("\n")

	headerArguments := make(
		[]interface{},
		len(columnHeaders),
	)

	for i, header := range columnHeaders {
		headerArguments[i] = header
	}

	fmt.Printf(
		headerFormatBuilder.String(),
		headerArguments...,
	)

	fmt.Println(
		hLine(
			"├",
			"┼",
			"┤",
			columnWidths,
		),
	)

	var dataFormatBuilder strings.Builder

	dataFormatBuilder.WriteString("│")

	for columnIndex, width := range columnWidths {
		switch columnIndex {
		case 0, 1:
			dataFormatBuilder.WriteString(
				fmt.Sprintf(" %%-%ds │", width),
			)

		default:
			dataFormatBuilder.WriteString(
				fmt.Sprintf(" %%%dd │", width),
			)
		}
	}

	dataFormatBuilder.WriteString("\n")

	dataFormat := dataFormatBuilder.String()

	for processIndex, process := range processes {
		arguments := make(
			[]interface{},
			0,
			len(columnHeaders),
		)

		arguments = append(
			arguments,
			process.pidCell,
			process.state,
			process.parent,
		)

		for _, registerValue := range process.regs {
			arguments = append(
				arguments,
				registerValue,
			)
		}

		fmt.Printf(
			dataFormat,
			arguments...,
		)

		if processIndex < len(processes)-1 {
			fmt.Println(
				hLine(
					"├",
					"┼",
					"┤",
					columnWidths,
				),
			)
		}
	}

	fmt.Println(
		hLine(
			"└",
			"┴",
			"┘",
			columnWidths,
		),
	)

	fmt.Println(
		"  PID marker: '>' = running    " +
			"Flags: bit0=mapped, bit1=zombie, bit2=waiting, " +
			"bit3-4=state(00=ready,01=running,10=blocked)",
	)

	fmt.Println()
}