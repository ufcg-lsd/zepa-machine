package machine

// This is a compact, test-only translation of the legacy ASM layout, not the
// F2-01 physical layout and not a boot or host interrupt integration test.
// Baseline at 23273e8: go test ./..., go test -tags kernel ./..., and
// go test ./machine ./cross-assembler passed; fork.asm/fork_test.asm assembled
// to 44/76 bytes. The latter compares child W9 to zero although specs.md says -2.
import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	assembler "zepa-machine/cross-assembler"
)

const (
	fixtureKPTR   = 0x1000000
	fixtureUPTR   = 0x1100000
	fixtureFrames = 0x1400000
	legacyPCB     = 0x2024
	legacyStride  = 0x300050
	legacyBitmap  = 0x30107024
)

type kernelFixture struct {
	t           *testing.T
	m           *Machine
	labels      map[string]uint32
	pages       map[uint32]uint32 // legacy kernel page -> compact physical frame
	next        uint32
	phase       string
	bitmapReady bool
}

func newKernelFixture(t *testing.T) *kernelFixture {
	t.Helper()
	src, err := os.ReadFile("../sackOS/kernel.asm")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := assembler.LoadAssemblyFromReader(strings.NewReader(string(src)))
	if err != nil {
		t.Fatal(err)
	}
	image, err := assembler.ConvertInstructionsToBinary(parsed)
	if err != nil {
		t.Fatal(err)
	}
	f := &kernelFixture{t: t, m: NewMachine(32<<20, false), labels: map[string]uint32{}, pages: map[uint32]uint32{}, next: fixtureFrames}
	// Duplicate the assembler's documented line handling; independently verify
	// every label operand against its resolved instruction displacement.
	type reference struct {
		ins, arg int
		name     string
	}
	var refs []reference
	s := bufio.NewScanner(strings.NewReader(string(src)))
	n := 0
	for s.Scan() {
		line := strings.TrimSpace(strings.SplitN(s.Text(), ";", 2)[0])
		if strings.HasSuffix(line, ":") {
			name := strings.TrimSuffix(line, ":")
			if _, ok := f.labels[name]; ok {
				t.Fatalf("duplicate label %s", name)
			}
			f.labels[name] = uint32(n * 4)
			continue
		}
		parts := strings.Fields(strings.ReplaceAll(line, ",", ""))
		if len(parts) == 0 {
			continue
		}
		for i, part := range parts[1:] {
			refs = append(refs, reference{n, i + 1, part})
		}
		n++
	}
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
	if n != len(parsed) || len(image) != n*4 {
		t.Fatal("assembler instruction count mismatch")
	}
	for _, r := range refs {
		if target, ok := f.labels[r.name]; ok {
			want := strconv.Itoa(int(target/4) - r.ins)
			if parsed[r.ins][r.arg] != want {
				t.Fatalf("label %s displacement: got %s want %s", r.name, parsed[r.ins][r.arg], want)
			}
		}
	}
	for off := uint32(0); off < uint32(len(image)); off += pageSize {
		f.mapKernel(off)
	}
	for i, b := range image {
		f.m.memory[f.phys(uint32(i))] = b
	}
	f.mapKernel(0x2000)                   // legacy globals and PCB 0
	f.mapKernel(legacyPCB + legacyStride) // PCB 1 header
	f.mapKernel(0x30107024)               // legacy bitmap page, deliberately relocated
	f.m.registers[kptr] = fixtureKPTR
	f.m.registers[uptr] = fixtureUPTR
	f.m.registers[sr] = 32 // supervisor, MMU enabled
	// One independently backed user page, in addition to the relocated kernel.
	userFrame := f.next
	f.next += pageSize
	binary.LittleEndian.PutUint32(f.m.memory[fixtureUPTR:], isPteMappedMask|userFrame/pageSize)
	if got, ok := f.m.lookupPhysical(0, 4); !ok || got != userFrame {
		t.Fatalf("UPTR relocation got %#x mapped=%v, want %#x", got, ok, userFrame)
	}
	f.initBitmap()
	return f
}

// map_page reads little-endian bitmap words, least significant bit first.
// Reserve the low identity region, both page tables, and every fixture frame;
// mark bits outside actual backing unavailable too. Call after allocating frames.
func (f *kernelFixture) initBitmap() {
	f.t.Helper()
	frames := uint32(len(f.m.memory) / pageSize)
	for frame := uint32(0); frame < frames; frame += 32 {
		var word uint32
		for bit := uint32(0); bit < 32; bit++ {
			if frame+bit >= frames || (frame+bit)*pageSize < f.next {
				word |= 1 << bit
			}
		}
		f.put(legacyBitmap+frame/8, word)
	}
	f.put(0x200c, uint32(len(f.m.memory)))
	f.bitmapReady = true
}

// Fork forms physical page-table addresses by subtracting the kernel boundary,
// so those particular PCB pages must be identity backed, unlike other regions.
// Only requested tables are mapped. The fixed compact layout supports PIDs
// whose entire legacy PCB fits below KPTR; reject larger PIDs rather than alias.
func (f *kernelFixture) identityPCBPages(pids ...uint32) {
	f.t.Helper()
	for _, pid := range pids {
		start := uint64(legacyPCB) + uint64(pid)*legacyStride
		end := start + legacyStride
		if end > fixtureKPTR {
			f.t.Fatalf("PID %d PCB end %#x overlaps compact KPTR %#x", pid, end, fixtureKPTR)
		}
		for addr := uint32(start) & ^uint32(pageSize-1); uint64(addr) < end; addr += pageSize {
			page := addr / pageSize
			if old, ok := f.pages[page]; ok && old == addr {
				continue
			}
			if old, ok := f.pages[page]; ok {
				copy(f.m.memory[addr:addr+pageSize], f.m.memory[old:old+pageSize])
			}
			f.pages[page] = addr
			binary.LittleEndian.PutUint32(f.m.memory[fixtureKPTR+page*4:], isPteMappedMask|page)
			if got, ok := f.m.lookupPhysical(kernelBoundary+addr, 4); !ok || got != addr {
				f.t.Fatalf("PCB identity mapping %#x got %#x mapped=%v", addr, got, ok)
			}
		}
	}
}

func (f *kernelFixture) mapKernel(addr uint32) {
	f.t.Helper()
	page := addr / pageSize
	if _, ok := f.pages[page]; ok {
		return
	}
	for f.bitmapReady && int(f.next+pageSize) <= len(f.m.memory) {
		word := uint32(legacyBitmap) + (f.next/pageSize/32)*4
		if f.get(word)&(1<<(f.next/pageSize%32)) == 0 {
			break
		}
		f.next += pageSize
	}
	if int(f.next+pageSize) > len(f.m.memory) {
		f.t.Fatal("fixture frame exhaustion")
	}
	frame := f.next
	f.next += pageSize
	f.pages[page] = frame
	binary.LittleEndian.PutUint32(f.m.memory[fixtureKPTR+page*4:], isPteMappedMask|frame/pageSize)
	if f.bitmapReady {
		word := uint32(legacyBitmap) + (frame/pageSize/32)*4
		f.put(word, f.get(word)|(1<<(frame/pageSize%32)))
	}
	entry := binary.LittleEndian.Uint32(f.m.memory[fixtureKPTR+page*4:])
	if entry != isPteMappedMask|frame/pageSize {
		f.t.Fatal("PTE relocation mismatch")
	}
}

func (f *kernelFixture) phys(addr uint32) uint32 {
	f.t.Helper()
	frame, ok := f.pages[addr/pageSize]
	if !ok {
		f.t.Fatalf("unmapped fixture address %#x", addr)
	}
	return frame + addr%pageSize
}
func (f *kernelFixture) put(addr, value uint32) {
	binary.LittleEndian.PutUint32(f.m.memory[f.phys(addr):], value)
}
func (f *kernelFixture) get(addr uint32) uint32 {
	return binary.LittleEndian.Uint32(f.m.memory[f.phys(addr):])
}

func (f *kernelFixture) run(from, until string, limit int) {
	f.t.Helper()
	start, ok := f.labels[from]
	if !ok {
		f.t.Fatalf("missing label %s", from)
	}
	end, ok := f.labels[until]
	if !ok {
		f.t.Fatalf("missing label %s", until)
	}
	f.phase = from
	f.m.registers[pc] = kernelBoundary + start
	for i := 0; i < limit; i++ {
		current := f.m.registers[pc]
		if current == kernelBoundary+end {
			return
		}
		if current < kernelBoundary {
			f.fail("PC escaped kernel")
		}
		if _, ok := f.pages[(current-kernelBoundary)/pageSize]; !ok {
			f.fail("PC is unmapped")
		}
		if !f.m.fetch() {
			f.fail("fetch fault")
		}
		inst, ok := f.m.decode()
		if !ok {
			f.fail("decode fault")
		}
		if f.m.checkIllegalRegisterAccess(inst) || f.m.checkIllegalInstruction(inst) {
			f.fail("illegal instruction or register access")
		}
		f.m.execute(inst)
		if f.m.registers[pc] < kernelBoundary {
			f.fail("exception or PC escaped kernel")
		}
	}
	f.fail(fmt.Sprintf("step budget %d exhausted", limit))
}

// Continue through the actual scheduler MRET, not merely its entry label.
func (f *kernelFixture) runToMRET(from string, limit int) {
	f.t.Helper()
	start, ok := f.labels[from]
	if !ok {
		f.t.Fatalf("missing label %s", from)
	}
	f.phase = from + "->MRET"
	f.m.registers[pc] = kernelBoundary + start
	for i := 0; i < limit; i++ {
		current := f.m.registers[pc]
		if current < kernelBoundary {
			f.fail("escaped kernel before MRET")
		}
		if _, ok := f.pages[(current-kernelBoundary)/pageSize]; !ok {
			f.fail("PC is unmapped")
		}
		if !f.m.fetch() {
			f.fail("fetch fault")
		}
		inst, ok := f.m.decode()
		if !ok || f.m.checkIllegalRegisterAccess(inst) || f.m.checkIllegalInstruction(inst) {
			f.fail("decode or access fault")
		}
		f.m.execute(inst)
		if inst.opcode == MRET {
			return
		}
	}
	f.fail(fmt.Sprintf("MRET step budget %d exhausted", limit))
}

func (f *kernelFixture) fail(reason string) {
	f.t.Helper()
	f.t.Fatal(f.diagnostic(reason))
}

// Diagnostics must not fault again when the CPU or page tables are corrupt.
// Read effective page tables and physical memory without invoking translate,
// which changes exception registers and may itself index beyond the backing.
func (f *kernelFixture) diagnostic(reason string) string {
	pcValue := f.m.registers[pc]
	label := "<outside-kernel>"
	nearest := ""
	var best uint32
	if pcValue >= kernelBoundary {
		label = "<none>"
		for name, addr := range f.labels {
			if addr <= pcValue-kernelBoundary && (nearest == "" || addr > best || (addr == best && name < nearest)) {
				nearest, best = name, addr
			}
		}
		if nearest != "" {
			label = fmt.Sprintf("%s+%#x", nearest, pcValue-kernelBoundary-best)
		}
	}
	physicalWord := func(addr uint64) (uint32, bool) {
		if addr+4 > uint64(len(f.m.memory)) {
			return 0, false
		}
		return binary.LittleEndian.Uint32(f.m.memory[int(addr) : int(addr)+4]), true
	}
	kernelPTE := func(addr uint64) (uint32, bool) {
		if addr >= 1<<32-uint64(kernelBoundary) || !f.m.isMmuEnabled() {
			return 0, false
		}
		return physicalWord(uint64(f.m.registers[kptr]) + addr/pageSize*4)
	}
	kernelWord := func(addr uint64) (uint32, bool) {
		if addr >= 1<<32-uint64(kernelBoundary) || addr%4 != 0 || addr%pageSize > pageSize-4 {
			return 0, false
		}
		if !f.m.isMmuEnabled() {
			return physicalWord(uint64(kernelBoundary) + addr)
		}
		if !f.m.isKernelMode() {
			return 0, false
		}
		pte, ok := kernelPTE(addr)
		if !ok || pte&isPteMappedMask == 0 {
			return 0, false
		}
		return physicalWord(uint64(pte&0xfffff)*pageSize + addr%pageSize)
	}
	formatWord := func(value uint32, ok bool) string {
		if !ok {
			return "<unavailable>"
		}
		return fmt.Sprintf("%#x", value)
	}
	pid, pidOK := kernelWord(0x2010)
	pidText := "<unavailable>"
	w9Text, pagesText, pteText, bitmapText, mappingText := "<unavailable>", "<unavailable>", "<unavailable>", "<unavailable>", "<unavailable>"
	if pidOK {
		pidText = fmt.Sprint(pid)
		if pid < 256 {
			base := uint64(legacyPCB) + uint64(pid)*legacyStride
			mapping, mappingOK := kernelPTE(base + 56)
			mappingText = formatWord(mapping, mappingOK)
			value, ok := kernelWord(base + 56)
			w9Text = formatWord(value, ok)
			value, ok = kernelWord(base + 76)
			pagesText = formatWord(value, ok)
			pte, ok := kernelWord(base + 80)
			pteText = formatWord(pte, ok)
			if ok && pte&isPteMappedMask != 0 {
				frame := pte & 0xfffff
				bitmap, mapped := kernelWord(uint64(legacyBitmap) + uint64(frame/32)*4)
				if mapped {
					bitmapText = fmt.Sprintf("frame=%d word=%#x bit=%d", frame, bitmap, (bitmap>>(frame%32))&1)
				}
			}
		}
	}
	return fmt.Sprintf("%s: phase=%s PC=%#x label=%s PID=%s PCB.w9=%s PCB.pages_used=%s PCB.pte[0]=%s bitmap=%s PCB.w9_mapping_pte=%s KPTR=%#x UPTR=%#x ecr=%d efa=%#x", reason, f.phase, pcValue, label, pidText, w9Text, pagesText, pteText, bitmapText, mappingText, f.m.registers[kptr], f.m.registers[uptr], f.m.registers[ecr], f.m.registers[efa])
}

func TestKernelFixtureBudgetDiagnostic(t *testing.T) {
	const childEnv = "ZEPA_FIXTURE_BUDGET_DIAGNOSTIC_CHILD"
	if os.Getenv(childEnv) == "1" {
		f := newKernelFixture(t)
		pcb := uint32(legacyPCB + legacyStride)
		f.put(0x2010, 1)
		f.put(pcb+56, 0x12345678)
		f.put(pcb+76, 7)
		f.put(pcb+80, isPteMappedMask|1)
		f.run("loop", "schedule", 3)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestKernelFixtureBudgetDiagnostic$", "-test.count=1")
	cmd.Env = append(os.Environ(), childEnv+"=1")
	output, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("budget diagnostic child timed out: %s", output)
	}
	if err == nil {
		t.Fatalf("expected child test to fail on budget exhaustion: %s", output)
	}
	if _, ok := err.(*exec.ExitError); !ok {
		t.Fatalf("running child test: %v", err)
	}
	for _, want := range []string{"step budget 3 exhausted", "phase=loop", "label=loop+0x0", "PID=1", "PCB.w9=0x12345678", "PCB.pages_used=0x7", "PCB.pte[0]=0x100001", "bitmap=frame=1 word=0xffffffff bit=1", "PC=0x", "KPTR=", "UPTR="} {
		if !strings.Contains(string(output), want) {
			t.Errorf("budget diagnostic missing %q: %s", want, output)
		}
	}
}

func TestKernelFixtureDiagnosticUnavailable(t *testing.T) {
	f := newKernelFixture(t)
	f.phase = "user-return"
	f.m.registers[pc] = 8
	f.put(0x2010, ^uint32(0))
	text := f.diagnostic("probe")
	for _, want := range []string{"PC=0x8", "label=<outside-kernel>", "PID=4294967295", "PCB.w9=<unavailable>", "PCB.pages_used=<unavailable>", "PCB.pte[0]=<unavailable>", "bitmap=<unavailable>"} {
		if !strings.Contains(text, want) {
			t.Errorf("diagnostic missing %q: %s", want, text)
		}
	}
	f.put(0x2010, 200)
	if text := f.diagnostic("unmapped PCB"); !strings.Contains(text, "PID=200 PCB.w9=<unavailable>") {
		t.Fatalf("unmapped PCB diagnostic: %s", text)
	}
	pcb := uint32(legacyPCB + legacyStride)
	f.put(0x2010, 1)
	f.put(pcb+56, 0x12345678)
	f.put(pcb+76, 7)
	f.put(pcb+80, isPteMappedMask|1)
	pteAddr := fixtureKPTR + (pcb/pageSize)*4
	if _, ok := f.pages[pcb/pageSize]; !ok {
		t.Fatal("expected PCB backing")
	}
	binary.LittleEndian.PutUint32(f.m.memory[pteAddr:], 0)
	text = f.diagnostic("invalid PCB PTE")
	for _, want := range []string{"PID=1", "PCB.w9=<unavailable>", "PCB.pages_used=<unavailable>", "PCB.pte[0]=<unavailable>", "bitmap=<unavailable>", "PCB.w9_mapping_pte=0x0"} {
		if !strings.Contains(text, want) {
			t.Errorf("invalid PCB PTE diagnostic missing %q: %s", want, text)
		}
	}
	binary.LittleEndian.PutUint32(f.m.memory[fixtureKPTR+(0x2000/pageSize)*4:], 0)
	if text := f.diagnostic("unmapped"); !strings.Contains(text, "PID=<unavailable>") {
		t.Fatalf("unmapped global diagnostic: %s", text)
	}
}

func TestKernelFixtureHandlerAndIsolation(t *testing.T) {
	a, b := newKernelFixture(t), newKernelFixture(t)
	a.put(0x2010, 0)
	a.put(legacyPCB+56, 0x12345678)
	b.put(legacyPCB+56, 0x87654321)
	a.m.registers[w9] = 3 // real syscall dispatcher -> getPID -> schedule
	a.run("syscall_int", "schedule", 200)
	if got := a.get(legacyPCB + 56); got != 0 {
		t.Fatalf("getPID result %#x", got)
	}
	if got := b.get(legacyPCB + 56); got != 0x87654321 {
		t.Fatalf("fixture leaked PCB state: %#x", got)
	}
	for page, frame := range a.pages {
		va := kernelBoundary + page*pageSize
		got, ok := a.m.lookupPhysical(va, 4)
		if !ok || got != frame {
			t.Fatalf("relocation %#x -> %#x, want %#x", va, got, frame)
		}
	}
}

func TestKernelFixtureForkNoSlots(t *testing.T) {
	f := newKernelFixture(t)
	f.put(0x2000, 1)
	f.put(0x2010, 0)
	f.put(legacyPCB+72, 1)
	f.run("fork", "schedule", 300)
	if got := f.get(legacyPCB + 56); got != ^uint32(0) {
		t.Fatalf("fork no-slot result = %#x, want -1", got)
	}
}

func TestKernelFixturePID2Handler(t *testing.T) {
	f := newKernelFixture(t)
	f.identityPCBPages(0, 2)
	other := uint32(legacyPCB + 2*legacyStride)
	f.put(0x2010, 2)
	f.put(legacyPCB+56, 0xabcdef01)
	f.put(other+56, 0x12345678)
	f.m.registers[w9] = 3
	f.run("syscall_int", "schedule", 200)
	if f.get(other+56) != 2 || f.get(legacyPCB+56) != 0xabcdef01 {
		t.Fatalf("PID 2 getPID changed wrong PCB: pid2=%#x pid0=%#x", f.get(other+56), f.get(legacyPCB+56))
	}
	if f.phys(other)&^(pageSize-1) == f.phys(legacyPCB)&^(pageSize-1) {
		t.Fatal("PID 0 and PID 2 PCB pages share a frame")
	}
}

func TestKernelFixtureForkSuccessAndSampleDiscrepancy(t *testing.T) {
	f := newKernelFixture(t)
	f.identityPCBPages(0, 1, 2)
	// Stack is independently relocated; the CPU only accesses it virtually.
	const stack = uint32(0x1000000)
	f.mapKernel(stack)
	parent := uint32(legacyPCB)
	child := uint32(legacyPCB + legacyStride)
	parentPT, childPT := parent+80, child+80
	f.put(0x2000, 2)
	f.put(0x2010, 0)
	f.put(0x2018, kernelBoundary+stack+pageSize)
	f.put(parent+4, ^uint32(0))
	f.put(parent+72, 1)
	for off := uint32(20); off <= 68; off += 4 {
		f.put(parent+off, 0x1000+off)
	}
	f.put(parent+60, 8)
	f.put(parent+64, 0x3ff0)
	f.put(parent+68, 32|8)
	// Place the real sample in the sole parent page; fork must copy it.
	sample, err := assembler.RunAssembler("../asm/samples/fork_test.asm")
	if err != nil {
		t.Fatal(err)
	}
	if len(sample) != 76 {
		t.Fatalf("fork_test sample length %d, baseline 76", len(sample))
	}
	parentFrame := f.next
	f.next += pageSize
	otherFrame := f.next
	f.next += pageSize
	otherPCB := uint32(legacyPCB + 2*legacyStride)
	f.put(otherPCB+20, 0x12345678)
	f.put(otherPCB+56, 0x87654321)
	f.put(otherPCB+72, 9)
	f.put(otherPCB+76, 1)
	f.put(otherPCB+80, isPteMappedMask|otherFrame/pageSize)
	for i := uint32(0); i < pageSize; i++ {
		f.m.memory[otherFrame+i] = byte(i*37 + 11)
	}
	otherSnapshot := append([]byte(nil), f.m.memory[otherFrame:otherFrame+pageSize]...)
	otherHeader := append([]byte(nil), f.m.memory[f.phys(otherPCB):f.phys(otherPCB)+84]...)
	if parentFrame == otherFrame || f.phys(parent) == f.phys(otherPCB) {
		t.Fatal("processes share backing")
	}
	copy(f.m.memory[parentFrame:parentFrame+pageSize], sample)
	f.put(parentPT, isPteMappedMask|parentFrame/pageSize)
	f.put(parent+76, 1)
	f.initBitmap()
	allocated := f.next
	bitmapWord := legacyBitmap + (allocated/pageSize/32)*4
	bit := uint32(1) << (allocated / pageSize % 32)
	if f.get(bitmapWord)&bit != 0 {
		t.Fatal("next frame should be free")
	}
	if f.get(legacyBitmap) != ^uint32(0) {
		t.Fatal("low reserved frames not marked")
	}
	if f.get(legacyBitmap+(parentFrame/pageSize/32)*4)&(1<<(parentFrame/pageSize%32)) == 0 {
		t.Fatal("parent frame not reserved")
	}
	// Full 786432-entry scan and one real map_page/copy: no loop shortcut.
	f.run("fork", "schedule", 20000000)
	if f.get(parent+56) != 1 || f.get(child+56) != ^uint32(1) {
		t.Fatalf("fork returns parent=%#x child=%#x", f.get(parent+56), f.get(child+56))
	}
	for off := uint32(20); off <= 68; off += 4 {
		if off != 56 && f.get(child+off) != f.get(parent+off) {
			t.Fatalf("context offset %d parent=%#x child=%#x", off, f.get(parent+off), f.get(child+off))
		}
	}
	if f.get(parent+4) != 1 || f.get(child) != 0 || f.get(child+72)&1 == 0 || f.get(child+76) != 1 {
		t.Fatalf("fork PCB links/flags/pages parent.child=%d child.parent=%d flags=%#x pages=%d", f.get(parent+4), f.get(child), f.get(child+72), f.get(child+76))
	}
	pte := f.get(childPT)
	if pte != isPteMappedMask|allocated/pageSize || f.get(bitmapWord)&bit == 0 {
		t.Fatalf("child PTE=%#x bitmap=%#x", pte, f.get(bitmapWord))
	}
	if string(f.m.memory[parentFrame:parentFrame+pageSize]) != string(f.m.memory[allocated:allocated+pageSize]) {
		t.Fatal("fork did not copy parent page")
	}
	f.m.memory[allocated+pageSize-1] = 0xa5
	if f.m.memory[parentFrame+pageSize-1] != 0 {
		t.Fatal("parent and child frames alias")
	}
	// The scheduler selects child PID 1, restores the PCB, then executes MRET.
	f.runToMRET("schedule", 500)
	if f.get(0x2010) != 1 || f.m.registers[pc] != 8 || f.m.registers[sr] != 40 || f.m.registers[sp] != 0x3ff0 || f.m.registers[uptr] != childPT {
		t.Fatalf("scheduler return pid=%d pc=%#x sr=%#x sp=%#x uptr=%#x", f.get(0x2010), f.m.registers[pc], f.m.registers[sr], f.m.registers[sp], f.m.registers[uptr])
	}
	for reg := w0; reg <= w9; reg++ {
		want := f.get(child + 20 + uint32(reg)*4)
		if got := f.m.registers[reg]; got != want {
			t.Fatalf("scheduler register W%d=%#x want %#x", reg, got, want)
		}
	}
	if string(otherHeader) != string(f.m.memory[f.phys(otherPCB):f.phys(otherPCB)+84]) || string(otherSnapshot) != string(f.m.memory[otherFrame:otherFrame+pageSize]) {
		t.Fatal("fork/scheduler changed unrelated PID 2 PCB or frame")
	}
	// Branch probes resume the real MRET context, but do not execute the
	// samples' wait/exit syscalls end-to-end. fork_test expects zero and falls
	// into the parent path with the actual -2 child result.
	for i := 0; i < 3; i++ {
		if !f.m.fetch() {
			t.Fatal("sample fetch fault")
		}
		inst, ok := f.m.decode()
		if !ok || f.m.checkIllegalInstruction(inst) || f.m.checkIllegalRegisterAccess(inst) {
			t.Fatal("sample decode/access fault")
		}
		f.m.execute(inst)
	}
	if got := f.m.registers[pc]; got != 20 {
		t.Fatalf("sample child branch PC=%#x, expected parent fallthrough 0x14", got)
	}
	t.Log("existing fork_test.asm compares child W9 with zero; actual fork child W9=-2 takes parent fallthrough")
	other, err := assembler.RunAssembler("../asm/samples/fork.asm")
	if err != nil {
		t.Fatal(err)
	}
	if len(other) != 44 {
		t.Fatalf("fork sample length %d, baseline 44", len(other))
	}
	copy(f.m.memory[allocated:], other)
	f.m.registers[pc] = 0
	for i := 0; i < 4; i++ {
		if !f.m.fetch() {
			t.Fatal("fork sample fetch fault")
		}
		inst, ok := f.m.decode()
		if !ok {
			t.Fatal("fork sample decode fault")
		}
		f.m.execute(inst)
	}
	f.m.registers[pc] = 20 // resume after syscall, with the real child result
	f.m.registers[w9] = f.get(child + 56)
	for i := 0; i < 2; i++ {
		if !f.m.fetch() {
			t.Fatal("fork sample fetch fault")
		}
		inst, ok := f.m.decode()
		if !ok {
			t.Fatal("fork sample decode fault")
		}
		f.m.execute(inst)
	}
	if got := f.m.registers[pc]; got != 36 {
		t.Fatalf("fork.asm child branch PC=%#x, want 0x24", got)
	}
}
