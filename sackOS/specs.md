# sackOS User Specifications Manual

## 1. Execution Environment

sackOS is a time-sharing operating system with virtual memory implemented through paging. It provides a restricted execution environment for user programs, preemptively scheduling them using a round-robin approach.

### 1.1 Registers

* `ECR`, `ESA`, `ESR`, `EPC`, `KPTR`, `UPTR` and `EFA` are kernel-only registers, using them will result in a fatal illegal access fault.
* `W9` is defined as the primary register for system call identification and return values.

### 1.2 Instructions

* `MRET` is a kernel-only instruction, using it will result in a fatal illegal access fault.

### 1.3 Memory Space
* **Virtual Memory & Paging:** Your process executes within a flat 3GB virtual address space (addresses `0x00000000` through `0xBFFFFFFF`). The OS maps these virtual pages to physical 4KB frames transparently. The initial bytecode is mapped at the start of this virtual space.
* **Demand Paging:** Physical memory is not fully pre-allocated. When your process accesses a valid, unmapped virtual address for the first time, a hardware page fault occurs, and the OS dynamically allocates a physical frame to back that page.
* **Memory Protection:** The virtual address space from `0xC0000000` up to `0xFFFFFFFF` (the upper 1GB) is strictly reserved for the kernel. Any attempt by a user process to read, write, or execute memory at or above the 3GB boundary will result in a fatal memory fault.
* **Stack Pointer:** sackOS utilizes a **Full Descending Stack**. The `SP` is initialized to the absolute top of the user virtual address space (`0xC0000000`). Because it is descending, you must decrement its value before storing data to ensure it falls within the legal user space.

---

## 2. Process Lifecycle

### 2.1 Process States
A process in sackOS can be in one of the following states:
* **Ready:** Waiting to be assigned to the CPU.
* **Running:** Currently executing instructions.
* **Blocked (Waiting):** Waiting for a child process to exit.
* **Zombie:** Execution has finished, but the parent process has not yet acknowledged the exit status.

### 2.2 Termination and Exit Codes
Processes can be terminated in three ways. The system uses specific exit codes (status codes stored in `W9`) to denote how a process ended:
* **Normal Exit:** Triggered via the `exit()` syscall. Status code is user-defined.
* **General Fault (Exit Code `1`):** Triggered if the process attempts an illegal kernel memory access, an illegal register access, privileged instruction use, or an invalid instruction.
* **External Kill (Exit Code `2`):** Triggered if the OS directly terminates the process (e.g., via a system kill signal).
* **Out of Memory / Page Fault (Exit Code `3`):** Triggered if the process attempts to access a new virtual page, but the physical memory (bitmap) has run out of free frames to allocate, or the process tried to access kernel memory.
---

## 3. System Calls (Syscalls)

System calls are invoked by placing the specific Syscall ID into the `W9` register and issuing the software interrupt/syscall instruction, any syscall with an additional argument expects this argument to be in the `W8` register. Return values are also placed in `W9`. When making any syscall, the process is preempted by the scheduler, so users must not expect to resume execution immediately after a syscall.

### `fork()` - ID 0
Creates a new process by duplicating the calling process. The child process receives an exact copy of the parent's mapped memory pages and registers at the moment of the call.
* **Returns (in `W9`):**
    * To the **parent**: The PID of the newly created child process.
    * To the **parent**: `-1` if the OS has reached its maximum process capacity, or if there are no free physical frames left to allocate the child's Page Table and Process Control Block.
    * To the **child**: `-2`.

### `wait(status_addr)` - ID 1
Pauses the execution of the calling process until one of its child processes terminates. 
* **Behavior:** * If a child has already terminated (is a zombie), `wait` returns immediately, cleaning up the child.
    * The OS writes the child's exit status code into the memory address specified by `status_addr`.
* **Returns (in `W9`):**
    * The PID of the terminated child process.
    * `-1` if the calling process has no children.
* **Exceptions:** Triggers a fatal fault if `status_addr` points outside the process's valid virtual memory, or points to an address not mapped yet, so, when setting `status_addr`, make sure to "touch" the address first.

### `exit(status_code)` - ID 2
Terminates the calling process and returns the `status_code` to the parent process (if the parent is waiting).
* **Behavior:** The process becomes a "zombie" until the parent calls `wait()`. If the parent is already dead (orphan), the process and its physical frames are destroyed and freed immediately.
* **Note:** The system will overwrite `W9` with your `status_code` internally to pass it back to the parent.

### `getPID()` - ID 3
Retrieves the Process ID (PID) of the calling process.
* **Returns (in `W9`):** An unsigned 32-bit integer representing the current process's PID.

### `rele()` - ID 4
Yields the CPU. The calling process voluntarily pauses its execution, moving from "Running" to "Ready" state, and forces the OS scheduler to pick the next available process.
* **Returns:** Nothing. Execution will resume normally the next time the scheduler selects this process.

---

## 4. Operational Limits
As a user, you must adhere to the limits set by the system administrator at boot:
* **Memory Limits:** While the virtual address space is 3GB, actual memory usage is bounded by the physical RAM available on the machine. Unchecked memory expansion will eventually consume all physical frames and trigger a fatal Out of Memory termination (Exit Code `3`).
* **Process Limits:** Total concurrent processes across the entire system cannot exceed the hardware-defined `MAX_PROCESSES` limit (typically 256).
* **Time Slicing:** CPU execution is preemptive; long-running processes will be automatically interrupted every `TIME_SLICE` clock ticks to allow other programs to run.
* **Scheduling:** When creating a process, users must not assume fairness. The process will be placed in a non-deterministic position in the ready queue.