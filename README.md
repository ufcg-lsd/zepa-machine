# sackOS - Partitioned Kernel

A time-sharing operating system built on the ZEPA machine, with memory divided into fixed **partitions**. It provides a restricted environment for user programs, preemptively scheduled in round-robin fashion.

## Documentation

Each piece of documentation is kept on this branch:

- `specs/ISA/ISA.md` — the Instruction Set Architecture
- `specs/assembly/assembly.md` — the assembly language
- `sackOS/specs.md` — user manual (registers, memory model, syscalls)
- `sackOS/abstract_implementation.md` — kernel internals (interrupt handlers, PCB, partitions, scheduler, memory addresses)

## Running the kernel

Check out this branch and run it with the `kernel` build tag:

```
go run -tags kernel . <memory_size> <partition_size> <time_slice>
```

example of usage:
```
go run -tags kernel . 512MB 1MB 128
```

Arguments:

| Argument | Required | Description |
|---|---|---|
| `memory_size` | yes | Total memory: plain bytes or with a suffix (`KB`, `MB`, `GB`, case-insensitive, e.g. `1GB`). Must be a multiple of 4, at least `8MB + 64KB + partition_size` and at most `4GB`. |
| `partition_size` | yes | Size of each fixed memory partition: plain bytes or with a suffix (`KB`, `MB`, `GB`, case-insensitive). Must be a multiple of 4. |
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
- `pcb` — show the kernel variables and PCB vector
- `refresh` — redraw the interface
- `q` (or `quit`) — exit the debugger

## Quick notes about the kernel

The partitioned sackOS divides memory into fixed-size partitions, managed by the kernel:

- **Fixed partitions:** each process runs in a contiguous block of `partition_size` bytes, addressed relatively from `0`; the OS translates access to physical memory via `BASE`/`LIMIT`.
- **Preemptive scheduling:** processes run in round-robin; the scheduler preempts the running process every `time_slice` clock interrupts.
- **Concurrency limit:** up to `partition_number` processes (one per partition), each with its own PCB.
- **Syscalls:** invoked through `W9` — `fork`, `wait`, `exit`, `getPID`, `rele`; returns also go in `W9`.
- **Exit codes:** `1` general fault, `2` external kill, or a user-defined code on `exit()`.
- **Stack:** full descending stack; `SP` starts at `partition_size`.

To understand more of the kernel's implementation, check out the documentation.