package machine

import (
	"fmt"
	"io"
	"os"
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
		return "running"
	case 1:
		return "ready"
	case 2:
		return "blocked"
	default:
		return "unknown"
	}
}

func pcbFlagString(flags byte) string {
	var s string

	if flags&1 != 0 {
		s += "M"
	} else {
		s += "."
	}

	if flags&2 != 0 {
		s += "Z"
	} else {
		s += "."
	}

	if flags&4 != 0 {
		s += "W"
	} else {
		s += "."
	}

	state := (flags >> 3) & 3

	switch state {
	case 0:
		s += "R"
	case 1:
		s += "Y"
	case 2:
		s += "B"
	default:
		s += "?"
	}

	return s
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

func (m *Machine) DebugRegistersString() string {
	var buf strings.Builder
	m.debugRegistersTo(&buf)
	return buf.String()
}

func (m *Machine) DebugRegisters() {
	m.debugRegistersTo(os.Stdout)
}

func (m *Machine) debugRegistersTo(out io.Writer) {
	m.mu.RLock()
	defer m.mu.RUnlock()
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

	fmt.Fprintln(out, "┌"+repeat("─", totalWidth-2)+"┐")
	fmt.Fprintln(out, titleLine(totalWidth, "CPU REGISTERS"))
	fmt.Fprintln(out, hLine("├", "┬", "┤", columnWidths))

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

	fmt.Fprintf(out,
		headerFormat,
		"REG",
		"DEC",
		"HEX",
		"REG",
		"DEC",
		"HEX",
	)

	fmt.Fprintln(out, hLine("├", "┼", "┤", columnWidths))

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
			fmt.Fprintf(out,
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

		fmt.Fprintf(out,
			rowFormat,
			pair[0].name,
			pair[0].dec,
			pair[0].hex,
			pair[1].name,
			pair[1].dec,
			pair[1].hex,
		)
	}

	fmt.Fprintln(out, hLine("└", "┴", "┘", columnWidths))

	if srLine != "" || ecrLine != "" {
		fmt.Fprintf(out, "  %s", srLine)

		if srLine != "" && ecrLine != "" {
			fmt.Fprint(out, "    ")
		}

		fmt.Fprintln(out, ecrLine)
	}

	fmt.Fprintln(out)
}

func runningPidStr(runningPid uint32) string {
	if runningPid == 0xFFFFFFFF {
		return "idle"
	}

	return fmt.Sprintf("%d", runningPid)
}

func (m *Machine) DebugSystemString() string {
	var buf strings.Builder
	m.debugSystemTo(&buf)
	return buf.String()
}

func (m *Machine) DebugSystem() {
	m.debugSystemTo(os.Stdout)
}

func (m *Machine) debugSystemTo(out io.Writer) {
	m.debugKernelVarsTo(out)
	m.debugProcessTableTo(out)
}

func (m *Machine) GetKernelVarsString() string {
	var buf strings.Builder
	m.debugKernelVarsTo(&buf)
	return buf.String()
}

func (m *Machine) debugKernelVarsTo(out io.Writer) {
	m.mu.RLock()
	defer m.mu.RUnlock()
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
		{"  total_pcb_memory", pcbSize * partitionNumber, totalPcbMemoryExtra},
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

	columnWidths := []int{nameWidth, decimalWidth, hexWidth}

	rowFormat := fmt.Sprintf("\u2502 %%-%ds │ %%%dd │ %%-%ds │\n", nameWidth, decimalWidth, hexWidth)
	headerFormat := fmt.Sprintf("\u2502 %%-%ds │ %%%ds │ %%-%ds │\n", nameWidth, decimalWidth, hexWidth)

	emitRow := func(name string, value uint32, extra string) {
		hexValue := fmt.Sprintf("0x%08X", value)
		if extra != "" {
			hexValue += " " + extra
		}
		fmt.Fprintf(out, rowFormat, name, int32(value), hexValue)
	}

	emitSection := func(label string) {
		fmt.Fprintf(out, "\u2502 %-*s │ %*s │ %-*s │\n", nameWidth, label, decimalWidth, "", hexWidth, "")
	}

	fmt.Fprintf(out, headerFormat, "Variable", "Value", "Hex")
	fmt.Fprintln(out, hLine("\u251C", "\u253C", "\u2524", columnWidths))

	emitSection("Constants")
	for i := 0; i < 4; i++ {
		emitRow(allRows[i].name, allRows[i].value, allRows[i].extra)
	}
	fmt.Fprintln(out, hLine("\u251C", "\u253C", "\u2524", columnWidths))

	emitSection("Singular Values")
	for i := 4; i < 11; i++ {
		emitRow(allRows[i].name, allRows[i].value, allRows[i].extra)
	}
	fmt.Fprintln(out, hLine("\u251C", "\u253C", "\u2524", columnWidths))

	emitSection("Data Structures")
	for i := 11; i < len(allRows); i++ {
		emitRow(allRows[i].name, allRows[i].value, allRows[i].extra)
	}

	fmt.Fprintln(out, hLine("\u2514", "\u2534", "\u2518", columnWidths))
	fmt.Fprintln(out)
}

func (m *Machine) GetProcessTableString() string {
	var buf strings.Builder
	m.debugProcessTableTo(&buf)
	return buf.String()
}

func (m *Machine) debugProcessTableTo(out io.Writer) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	memory := m.memory

	partitionNumber := readUint32(memory, 0x1014)
	runningPid := readUint32(memory, 0x1018)

	if partitionNumber == 0 {
		fmt.Fprintln(out, "Nenhum processo (partition_number = 0).")
		fmt.Fprintln(out)
		return
	}

	type pcbRow struct {
		pidCell  string
		flagsStr string
		parent   int32
		regs     [15]int32
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

		registerValues := [15]int32{
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
			int32(readUint32(memory, pcbAddress+72)),
			int32(readUint32(memory, pcbAddress+76)),
		}

		marker := " "

		if pid == runningPid {
			marker = ">"
		}

		processes = append(processes, pcbRow{
			pidCell:  fmt.Sprintf("%s%d", marker, pid),
			flagsStr: pcbFlagString(flags),
			parent:   parentPid,
			regs:     registerValues,
		})
	}

	columnHeaders := []string{
		"PID",
		"Flags",
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
		"BASE",
		"LIMIT",
	}

	columnWidths := make([]int, len(columnHeaders))

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
			process.flagsStr,
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

	fmt.Fprintln(out,
		"┌"+
			repeat("─", processTableWidth-2)+
			"┐",
	)

	fmt.Fprintln(out,
		titleLine(
			processTableWidth,
			fmt.Sprintf(
				"PROCESS TABLE  •  running_pid: %s",
				runningPidStr(runningPid),
			),
		),
	)

	fmt.Fprintln(out,
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

	fmt.Fprintf(out,
		headerFormatBuilder.String(),
		headerArguments...,
	)

	fmt.Fprintln(out,
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
			process.flagsStr,
			process.parent,
		)

		for _, registerValue := range process.regs {
			arguments = append(
				arguments,
				registerValue,
			)
		}

		fmt.Fprintf(out,
			dataFormat,
			arguments...,
		)

		if processIndex < len(processes)-1 {
			fmt.Fprintln(out,
				hLine(
					"├",
					"┼",
					"┤",
					columnWidths,
				),
			)
		}
	}

	fmt.Fprintln(out,
		hLine(
			"└",
			"┴",
			"┘",
			columnWidths,
		),
	)

	fmt.Fprintln(out,
		"  PID marker: '>' = running    "+
			"Flags: M=mapped, Z=zombie, W=waiting, "+
			"R=running, Y=ready, B=blocked",
	)

	fmt.Fprintln(out)
}

var opNames = map[byte]string{
	0: "MV", 1: "AND", 2: "OR", 3: "XOR",
	4: "ADD", 5: "SUB", 6: "MUL", 7: "UDIV",
	8: "SDIV", 9: "CMP", 10: "JUMP", 11: "JMPR",
	12: "BEQ", 13: "BLT", 14: "BGT",
	15: "LOAD", 16: "STORE", 17: "LDD", 18: "STRD",
	19: "LDB", 20: "LDSB", 21: "STRB", 22: "MRET",
	23: "SYSCALL",
}

var regNames = [22]string{
	"w0", "w1", "w2", "w3", "w4", "w5", "w6", "w7", "w8", "w9",
	"pc", "sp", "ir", "sr", "mdr", "mar",
	"ecr", "esa", "esr", "epc", "base", "limit",
}

func decodeInstruction(inst uint32) string {
	opcode := byte((inst >> 26) & 0x3F)
	name, ok := opNames[opcode]
	if !ok {
		return ".word 0x" + fmt.Sprintf("%08X", inst)
	}

	rType := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 15, 16, 19, 20, 21}
	isRType := false
	for _, op := range rType {
		if opcode == op {
			isRType = true
			break
		}
	}

	if isRType {
		rd := byte((inst >> 21) & 0x1F)
		rs1 := byte((inst >> 16) & 0x1F)
		rs2 := byte((inst >> 11) & 0x1F)
		if opcode == 9 {
			return fmt.Sprintf("%s %s, %s", name, regNames[rs1], regNames[rs2])
		}
		if opcode == 11 || opcode == 15 || opcode == 16 || opcode == 19 || opcode == 20 || opcode == 21 {
			return fmt.Sprintf("%s %s, [%s]", name, regNames[rd], regNames[rs1])
		}
		return fmt.Sprintf("%s %s, %s, %s", name, regNames[rd], regNames[rs1], regNames[rs2])
	}

	rs1rd := byte((inst >> 21) & 0x1F)
	imm := int16((inst >> 5) & 0xFFFF)
	if opcode == 10 || opcode == 12 || opcode == 13 || opcode == 14 {
		return fmt.Sprintf("%s %+d", name, imm)
	}
	if opcode == 22 {
		return "MRET"
	}
	if opcode == 23 {
		return fmt.Sprintf("SYSCALL #%d", imm)
	}
	return fmt.Sprintf("%s %s, #%d", name, regNames[rs1rd], imm)
}

func (m *Machine) GetMemoryViewString() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var buf strings.Builder
	w := &buf

	pc := m.registers[pc]
	if pc < 4 {
		pc = 4
	}
	currentAddr := pc

	startAddr := currentAddr
	if startAddr >= 8 {
		startAddr -= 8
	} else {
		startAddr = 0
	}

	type instrEntry struct {
		addr      uint32
		raw       uint32
		decode    string
		isCurrent bool
	}

	var entries []instrEntry
	for addr := startAddr; addr <= currentAddr+8; addr += 4 {
		physAddr, _ := m.translate(addr, 3)
		raw := m.ReadWord(physAddr)
		if raw == 0 {
			continue
		}
		decoded := decodeInstruction(raw)
		entries = append(entries, instrEntry{
			addr:      addr,
			raw:       raw,
			decode:    decoded,
			isCurrent: addr == currentAddr,
		})
	}

	centerIdx := 0
	for i, e := range entries {
		if e.isCurrent {
			centerIdx = i
			break
		}
	}

	from := centerIdx - 2
	if from < 0 {
		from = 0
	}
	to := from + 4
	if to >= len(entries) {
		to = len(entries) - 1
		from = to - 4
		if from < 0 {
			from = 0
		}
	}

	displayed := entries[from : to+1]

	addrW := 6
	hexW := 10
	instrW := 0
	for _, e := range displayed {
		if len(e.decode) > instrW {
			instrW = len(e.decode)
		}
	}
	if instrW < 7 {
		instrW = 7
	}

	fmt.Fprintf(w, " %-*s \u2502 %-*s \u2502 %-*s\n",
		addrW+1, "ADDR",
		hexW, "HEX",
		instrW, "INSTRUCTION")

	fmt.Fprintf(w, strings.Repeat("\u2500", addrW+3)+"\u253C"+
		strings.Repeat("\u2500", hexW+2)+"\u253C"+
		strings.Repeat("\u2500", instrW+2)+"\n")

	for _, e := range displayed {
		marker := " "
		if e.isCurrent {
			marker = ">"
		}
		fmt.Fprintf(w, "%s%-*s \u2502 0x%08X \u2502 %-*s\n",
			marker,
			addrW+1, fmt.Sprintf("0x%04X", e.addr),
			e.raw,
			instrW, e.decode)
	}

	return buf.String()
}

func (m *Machine) GetPCBVectorString() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var buf strings.Builder
	w := &buf

	memory := m.memory
	partitionNumber := readUint32(memory, 0x1014)
	runningPid := readUint32(memory, 0x1018)

	if partitionNumber == 0 {
		fmt.Fprintln(w, "No processes.")
		return buf.String()
	}

	type pcbEntry struct {
		pid         int
		parent      int32
		child       int32
		nextSib     int32
		prevSib     int32
		statusAddr  int32
		w0, w1, w2  int32
		w3, w4, w5  int32
		w6, w7, w8  int32
		w9          int32
		pc, sp, sr  int32
		base, limit int32
		flags       byte
		state       string
	}

	var entries []pcbEntry

	for pid := uint32(0); pid < partitionNumber; pid++ {
		pcbAddr := 0x102C + pid*84
		flags := memory[pcbAddr+80]
		if flags&1 == 0 {
			continue
		}

		entries = append(entries, pcbEntry{
			pid:        int(pid),
			parent:     int32(readUint32(memory, pcbAddr+0)),
			child:      int32(readUint32(memory, pcbAddr+4)),
			nextSib:    int32(readUint32(memory, pcbAddr+12)),
			prevSib:    int32(readUint32(memory, pcbAddr+8)),
			statusAddr: int32(readUint32(memory, pcbAddr+16)),
			w0:         int32(readUint32(memory, pcbAddr+20)),
			w1:         int32(readUint32(memory, pcbAddr+24)),
			w2:         int32(readUint32(memory, pcbAddr+28)),
			w3:         int32(readUint32(memory, pcbAddr+32)),
			w4:         int32(readUint32(memory, pcbAddr+36)),
			w5:         int32(readUint32(memory, pcbAddr+40)),
			w6:         int32(readUint32(memory, pcbAddr+44)),
			w7:         int32(readUint32(memory, pcbAddr+48)),
			w8:         int32(readUint32(memory, pcbAddr+52)),
			w9:         int32(readUint32(memory, pcbAddr+56)),
			pc:         int32(readUint32(memory, pcbAddr+60)),
			sp:         int32(readUint32(memory, pcbAddr+64)),
			sr:         int32(readUint32(memory, pcbAddr+68)),
			base:       int32(readUint32(memory, pcbAddr+72)),
			limit:      int32(readUint32(memory, pcbAddr+76)),
			flags:      flags,
			state:      pcbState(flags),
		})
	}

	if len(entries) == 0 {
		fmt.Fprintln(w, "No mapped processes.")
		return buf.String()
	}

	cols := []string{"PID", "State", "Parent", "Child", "W0", "W1", "W2", "W3", "W4", "W5", "W6", "W7", "W8", "W9", "PC", "SP", "SR"}
	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = len(c)
	}

	for _, e := range entries {
		pidStr := fmt.Sprintf("%d", e.pid)
		if e.pid == int(runningPid) {
			pidStr = ">" + pidStr
		}
		widths[0] = max(widths[0], len(pidStr))
		widths[1] = max(widths[1], len(e.state))
		vals := []int32{e.parent, e.child, e.w0, e.w1, e.w2, e.w3, e.w4, e.w5, e.w6, e.w7, e.w8, e.w9, e.pc, e.sp, e.sr}
		for i, v := range vals {
			widths[i+2] = max(widths[i+2], len(fmt.Sprintf("%d", v)))
		}
	}

	fmt.Fprintf(w, "\u250C")
	for i, width := range widths {
		fmt.Fprintf(w, "%s\u2500\u2500", strings.Repeat("\u2500", width+2))
		if i < len(widths)-1 {
			fmt.Fprint(w, "\u252C")
		}
	}
	fmt.Fprintln(w, "\u2510")

	fmt.Fprintf(w, "\u2502")
	for i, c := range cols {
		fmt.Fprintf(w, " %-*s \u2502", widths[i], c)
	}
	fmt.Fprintln(w)

	fmt.Fprintf(w, "\u251C")
	for i, width := range widths {
		fmt.Fprintf(w, "%s\u2500\u2500", strings.Repeat("\u2500", width+2))
		if i < len(widths)-1 {
			fmt.Fprint(w, "\u253C")
		}
	}
	fmt.Fprintln(w, "\u2524")

	for idx, e := range entries {
		pidStr := fmt.Sprintf("%d", e.pid)
		if e.pid == int(runningPid) {
			pidStr = ">" + pidStr
		}
		vals := []interface{}{
			pidStr, e.state, e.parent, e.child,
			e.w0, e.w1, e.w2, e.w3, e.w4, e.w5, e.w6, e.w7, e.w8, e.w9,
			e.pc, e.sp, e.sr,
		}
		fmt.Fprintf(w, "\u2502")
		for i, v := range vals {
			if i <= 1 {
				fmt.Fprintf(w, " %-*s \u2502", widths[i], fmt.Sprintf("%v", v))
			} else {
				fmt.Fprintf(w, " %*d \u2502", widths[i], v)
			}
		}
		fmt.Fprintln(w)

		if idx < len(entries)-1 {
			fmt.Fprintf(w, "\u251C")
			for i, width := range widths {
				fmt.Fprintf(w, "%s\u2500\u2500", strings.Repeat("\u2500", width+2))
				if i < len(widths)-1 {
					fmt.Fprint(w, "\u253C")
				}
			}
			fmt.Fprintln(w, "\u2524")
		}
	}

	fmt.Fprintf(w, "\u2514")
	for i, width := range widths {
		fmt.Fprintf(w, "%s\u2500\u2500", strings.Repeat("\u2500", width+2))
		if i < len(widths)-1 {
			fmt.Fprint(w, "\u2534")
		}
	}
	fmt.Fprintln(w, "\u2518")

	return buf.String()
}
