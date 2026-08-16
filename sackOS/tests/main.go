package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	assembler "zepa-machine/cross-assembler"
	"zepa-machine/machine"
)

const (
	bufferSize       = 64 * 1024
	paginatedMinSize = (1 << 30) + (1 << 12) // 1GB + 4KB
	pcbBase          = 0x2024
	pcbSize          = 0x300050
	flagsOffset      = 72

	maxProcessesAddr = 0x2000
	timeSliceAddr    = 0x2004
	bufferSizeAddr   = 0x2008
	runningPidAddr   = 0x2010
	readyMarker      = 0xFFFFFFFF

	maxProcesses uint32 = 256

	defaultTimeout = 30 * time.Second
	bootTimeout    = 10 * time.Second
)

var regOffsets = []uint32{20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60, 64, 68}

var regNames = map[string]int{
	"w0": 0, "w1": 1, "w2": 2, "w3": 3, "w4": 4,
	"w5": 5, "w6": 6, "w7": 7, "w8": 8, "w9": 9,
	"pc": 10, "sp": 11, "sr": 12,
}

type Program struct {
	File     string            `json:"file"`
	Expected map[string]uint32 `json:"expected,omitempty"`
}

type Config struct {
	MemorySize     int       `json:"memory_size"`
	PartitionSize  int       `json:"partition_size"`
	TimeSlice      int       `json:"time_slice"`
	TimeoutSeconds int       `json:"timeout_seconds,omitempty"`
	Programs       []Program `json:"programs"`
}

type Result struct {
	File       string     `json:"file"`
	PID        uint32     `json:"pid"`
	Terminated bool       `json:"terminated"`
	Registers  [13]uint32 `json:"registers"`
	ExitCode   uint32     `json:"exit_code"`
	Passed     bool       `json:"passed"`
	Errors     []string   `json:"errors,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./sackOS/tests <config.json>")
		os.Exit(1)
	}

	cfg, err := loadConfig(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	timeout := defaultTimeout
	if cfg.TimeoutSeconds > 0 {
		timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}

	m, partitionNumber, err := bootKernel(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "boot error: %v\n", err)
		os.Exit(1)
	}

	results := runPrograms(m, cfg, partitionNumber, timeout)

	encoder := json.NewEncoder(os.Stdout)
	passed := 0
	for _, r := range results {
		if err := encoder.Encode(r); err != nil {
			fmt.Fprintf(os.Stderr, "encoding result: %v\n", err)
			os.Exit(1)
		}
		if r.Passed {
			passed++
		}
	}

	summary := struct {
		Total  int `json:"total"`
		Passed int `json:"passed"`
		Failed int `json:"failed"`
	}{len(results), passed, len(results) - passed}

	if err := encoder.Encode(summary); err != nil {
		fmt.Fprintf(os.Stderr, "encoding summary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.MemorySize == 0 || cfg.PartitionSize == 0 || cfg.TimeSlice == 0 {
		return nil, errors.New("memory_size, partition_size and time_slice are required")
	}
	if len(cfg.Programs) == 0 {
		return nil, errors.New("programs list is empty")
	}

	dir := filepath.Dir(path)
	for i := range cfg.Programs {
		if cfg.Programs[i].File == "" {
			return nil, fmt.Errorf("program %d has no file", i)
		}
		if !filepath.IsAbs(cfg.Programs[i].File) {
			cfg.Programs[i].File = filepath.Join(dir, cfg.Programs[i].File)
		}
	}

	return &cfg, nil
}

func bootKernel(cfg *Config) (*machine.Machine, uint32, error) {
	if cfg.MemorySize%4 != 0 {
		return nil, 0, fmt.Errorf("memory_size must be a multiple of 4, got %d", cfg.MemorySize)
	}
	if cfg.MemorySize < paginatedMinSize+bufferSize {
		return nil, 0, fmt.Errorf("memory_size must be at least %d + %d (paginated kernel), got %d", paginatedMinSize, bufferSize, cfg.MemorySize)
	}

	kernel, err := assembler.RunAssembler(resolveKernelPath())
	if err != nil {
		return nil, 0, fmt.Errorf("assembling kernel: %w", err)
	}

	m := machine.NewMachine(cfg.MemorySize, false)
	m.LoadProgram(kernel)

	mem := m.GetMemory()
	binary.LittleEndian.PutUint32(mem[maxProcessesAddr:maxProcessesAddr+4], maxProcesses)
	binary.LittleEndian.PutUint32(mem[timeSliceAddr:timeSliceAddr+4], uint32(cfg.TimeSlice))
	binary.LittleEndian.PutUint32(mem[bufferSizeAddr:bufferSizeAddr+4], uint32(bufferSize))

	go m.Boot()

	deadline := time.Now().Add(bootTimeout)
	for readUint32(m.GetMemory(), runningPidAddr) != readyMarker {
		if time.Now().After(deadline) {
			return nil, 0, errors.New("kernel did not finish setup in time")
		}
		time.Sleep(time.Millisecond)
	}

	return m, maxProcesses, nil
}

func resolveKernelPath() string {
	candidates := []string{
		"sackOS/kernel.asm",
		filepath.Join("..", "kernel.asm"),
		filepath.Join("sackOS", "tests", "..", "kernel.asm"),
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	return "sackOS/kernel.asm"
}

func runPrograms(m *machine.Machine, cfg *Config, partitionNumber uint32, timeout time.Duration) []Result {
	results := make([]Result, 0, len(cfg.Programs))
	tracked := make(map[uint32]bool)
	mem := m.GetMemory()

	for _, p := range cfg.Programs {
		r := Result{File: p.File}

		code, err := assembler.RunAssembler(p.File)
		if err != nil {
			r.Errors = append(r.Errors, fmt.Sprintf("assembling: %v", err))
			results = append(results, r)
			continue
		}

		if len(code) > cfg.PartitionSize {
			r.Errors = append(r.Errors, "program larger than partition_size")
			results = append(results, r)
			continue
		}

		if !m.LoadBuffer(code) {
			r.Errors = append(r.Errors, "program larger than buffer")
			results = append(results, r)
			continue
		}
		m.SetInputFlag()

		pid, ok := waitForMapping(mem, partitionNumber, tracked, timeout)
		if !ok {
			r.Errors = append(r.Errors, "no process created (input was dropped by the kernel)")
			results = append(results, r)
			continue
		}

		r.PID = pid
		tracked[pid] = true

		timedOut := false
		if !waitForDeath(mem, pid, timeout) {
			killPID(m, pid)
			if !waitForDeath(mem, pid, 2*timeout) {
				r.Terminated = false
				r.Errors = append(r.Errors, "program did not terminate even after kill")
				results = append(results, r)
				continue
			}
			timedOut = true
			r.Errors = append(r.Errors, "program did not finish in time; killed externally")
		}

		r.Terminated = true
		r.Registers = snapshot(mem, pid)
		r.ExitCode = r.Registers[9]
		delete(tracked, pid)

		validate(&r, p.Expected)
		if timedOut {
			r.Passed = false
		}

		results = append(results, r)
	}

	return results
}

func waitForMapping(mem []byte, partitionNumber uint32, tracked map[uint32]bool, timeout time.Duration) (uint32, bool) {
	deadline := time.Now().Add(timeout)
	for {
		for pid := uint32(0); pid < partitionNumber; pid++ {
			if tracked[pid] {
				continue
			}
			if isMapped(mem, pid) {
				return pid, true
			}
		}
		if time.Now().After(deadline) {
			return 0, false
		}
		time.Sleep(time.Millisecond)
	}
}

func waitForDeath(mem []byte, pid uint32, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		if !isMapped(mem, pid) {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(time.Millisecond)
	}
}

func killPID(m *machine.Machine, pid uint32) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, pid)
	m.LoadBuffer(buf)
	m.SetKillFlag()
}

func isMapped(mem []byte, pid uint32) bool {
	return mem[pcbBase+pid*pcbSize+flagsOffset]&1 != 0
}

func snapshot(mem []byte, pid uint32) [13]uint32 {
	var regs [13]uint32
	base := pcbBase + pid*pcbSize
	for i, off := range regOffsets {
		regs[i] = readUint32(mem, base+off)
	}
	return regs
}

func validate(r *Result, expected map[string]uint32) {
	if len(expected) == 0 {
		r.Passed = true
		return
	}

	r.Passed = true
	for name, want := range expected {
		idx, ok := regNames[name]
		if !ok {
			r.Errors = append(r.Errors, fmt.Sprintf("unknown register %q", name))
			r.Passed = false
			continue
		}
		got := r.Registers[idx]
		if got != want {
			r.Errors = append(r.Errors, fmt.Sprintf("%s: expected %d, got %d", name, want, got))
			r.Passed = false
		}
	}
}

func readUint32(mem []byte, addr uint32) uint32 {
	return uint32(mem[addr]) |
		uint32(mem[addr+1])<<8 |
		uint32(mem[addr+2])<<16 |
		uint32(mem[addr+3])<<24
}
