# sackOS - Paginated Kernel

A time-sharing operating system built on the ZEPA machine, with virtual memory implemented through **paging**. It provides a restricted environment for user programs, preemptively scheduled in round-robin fashion.

## What is the paginated sackOS

The paginated sackOS manages memory with fixed-size pages of **4 KB**, translated through page tables enabled by the MMU:

- **Virtual memory:** `3 GB` for user space (low addresses) and `1 GB` for the kernel (high addresses, mapped `3 GB` above the user space).
- **Preemptive scheduling:** processes run in round-robin; the scheduler preempts the running process every `time_slice` clock interrupts.
- **Lazy mapping:** pages are mapped on demand on a page fault; if there is no free frame, the process dies with exit code `3`.
- **Up to 256 processes**, each with its own page table and PCB.
- **Exit codes:** `1` general fault, `2` external kill, `3` out of frames (page fault), or a user-defined code on `exit()`.

### Documentation

Each piece of documentation is kept on this branch:

- `specs/ISA/ISA.md` — the Instruction Set Architecture
- `specs/assembly/assembly.md` — the assembly language
- `sackOS/specs.md` — user manual (registers, memory model, syscalls)
- `sackOS/abstract_implementation.md` — kernel internals (interrupt handlers, PCB, page tables, scheduler, memory addresses)

## Running the kernel

Check out the `sackOS-paginated` branch and run it with the `kernel` build tag:

```
go run -tags kernel . <memory_size> <time_slice>
```

example of usage:
```
go run -tags kernel . 2GB 128
```

Arguments:

| Argument | Required | Description |
|---|---|---|
| `memory_size` | yes | Total memory: plain bytes or with a suffix (`KB`, `MB`, `GB`, case-insensitive, e.g. `1GB`). Must be a multiple of 4, at least `1GB + 4KB` and at most `4GB`. |
| `time_slice` | yes | Scheduler quantum, in **clock interrupts**. Each clock interrupt fires every `128` executed instructions. |

There are no flags: when the kernel starts, the TUI launches automatically and debug mode is always enabled.

### Using the TUI

The TUI opens a debugger with an output panel and a command line (`cmd> `). Available commands:

- `play` / `stop` — start / pause continuous execution
- `d <steps>` (or `step`) — step N instructions (default 1)
- `b <pc>` (or `breakpoint`) — run until the PC reaches the value
- `c` (or `count`) — show instructions/cycles executed since boot
- `input <path>` — load and run an assembly program from a `.asm` file
- `kill <pid>` — kill a process
- `reg` — show the registers
- `pcb` — show the process table
- `refresh` — redraw the interface

