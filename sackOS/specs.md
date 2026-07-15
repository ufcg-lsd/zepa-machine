# sackOS User Specifications Manual

## 1. Execution Environment

sackOS is a fixed-partition, time-sharing operating system. It provides a restricted execution environment for user programs, preemptively scheduling them using a round-robin approach.

### 1.1 Registers

* `ECR`, `ESA`, `ESR`, `EPC`, `BASE` and `LIMIT` are kernel-only registers, using them will result in a fatal illegal access fault.
* `W9` is defined as the primary register for system call identification and return values.

### 1.2 Instructions

* `MRET` is a kernel-only instruction, using it will result in a fatal illegal access fault.

### 1.3 Memory Space
* **Fixed Partitions:** Your process is loaded into a contiguous memory block of `PARTITION_SIZE` bytes, the bytecode is stored at the start of this block. 
* **Addressing:** User space interacts with memory using relative addresses starting from `0` up to `PARTITION_SIZE - 1`. The OS handles the translation to physical memory (`BASE` and `LIMIT`) transparently.
* **Memory Protection:** Any attempt to read or write memory outside your designated partition (any address greater than `PARTITION_SIZE - 1`) will result in a fatal memory fault.
* **Stack Pointer:** sackOS has a **Full Descending Stack**, so the `SP` is initializes to `PARTITION_SIZE`, so make sure to decrement its value before storing data.

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
* **General Fault (Exit Code `1`):** Triggered if the process attempts an out-of-bounds memory access, an illegal register access or a privileged instruction use.
* **External Kill (Exit Code `2`):** Triggered if the OS directly terminates the process (e.g., via kill signal).

---

## 3. System Calls (Syscalls)

System calls are invoked by placing the specific Syscall ID into the `W9` register and issuing the software interrupt/syscall instruction, any syscall with an additional argument expects this argument to be in the `W8` register. Return values are also placed in `W9`. When making any syscall, the process is preempted by the scheduler, so users must not expect to resume execution immediately after a syscall.

### `fork()` - ID 0
Creates a new process by duplicating the calling process. The child process receives an exact copy of the parent's memory and registers at the moment of the call.
* **Returns (in `W9`):**
    * To the **parent**: The PID of the newly created child process.
    * To the **parent**: `-1` if the OS has reached its maximum partition capacity and cannot spawn a new process.
    * To the **child**: `0`.

### `wait(status_addr)` - ID 1
Pauses the execution of the calling process until one of its child processes terminates. 
* **Behavior:** * If a child has already terminated (is a zombie), `wait` returns immediately, cleaning up the child.
    * The OS writes the child's exit status code into the memory address specified by `status_addr`.
* **Returns (in `W9`):**
    * The PID of the terminated child process.
    * `-1` if the calling process has no children.
* **Exceptions:** Triggers a fatal fault if `status_addr` points outside the process's valid memory partition.

### `exit(status_code)` - ID 2
Terminates the calling process and returns the `status_code` to the parent process (if the parent is waiting).
* **Behavior:** The process becomes a "zombie" until the parent calls `wait()`. If the parent is already dead (orphan), the process is destroyed immediately.
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
* Memory is strictly confined to `PARTITION_SIZE`, which can be seem at runtime as the initial value of `SP`. You cannot allocate more memory dynamically.
* Total concurrent processes across the entire system cannot exceed `partition_number`, calculated as the maximum amount of partitions that the memory can store (excluding kernel memory).
* CPU execution is preemptive; long-running processes will be automatically interrupted every `TIME_SLICE` clock ticks to allow other programs to run.
* When creating a process, users must not assume fairness, the process will be put in a non deterministic position of the queue.