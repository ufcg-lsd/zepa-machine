# ZEPA Machine

ZEPA is a simulated machine designed to help you learn aspects of computer organization and operating systems.

## Branches

This repository has three branches. Each one serves a different purpose; pick the one that fits what you want to do.

### `old-zepa`

The ZEPA machine itself, with no kernel. Use this branch if you only want to run programs on the ZEPA machine, not the operating system kernel.

```
git checkout old-zepa
go run ./main.go <asm/file/path>
```

example of usage:
```
go run ./main.go asm/samples/add_two_number.asm
```

### `sackOS-partitioned`

The sackOS kernel built on top of the ZEPA machine, with memory divided into fixed partitions. Check out this branch if you want to run the partitioned kernel.

```
git checkout sackOS-partitioned
go run -tags kernel . <memory_size_in_MB> <partition_size> <time_slice> [--tui] [--no-debug]
```

example of usage:
```
go run -tags kernel . 1024 128 128 [--tui]
```

### `sackOS-paginated`

The sackOS kernel built on top of the ZEPA machine, with virtual memory implemented through paging. Check out this branch if you want to run the paginated kernel.

```
git checkout sackOS-paginated
go run -tags kernel . <memory_size_in_MB> <time_slice> [--tui] [--no-debug]
```

example of usage:
```
go run -tags kernel . 2048 128 [--tui]
```

Each kernel branch carries its own documentation: `README.md`, `specs/ISA/ISA.md`, `specs/assembly/assembly.md`, `sackOS/specs.md` and `sackOS/abstract_implementation.md`.

