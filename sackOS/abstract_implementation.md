# sackOS Abstract Implementation

## setup
```
user_memory = LIMIT - (KERNEL_MAX_MEMORY + BUFFER)
partition_number = USER_MEMORY / PARTICION_SIZE
SP = SP_ADDRESS
ESA = EXCEPTION_SUPERVISOR_ADDRESS
EPC = LOOP ADDRESS
ESR += 16 //will enable interruptions
MRET //go to infinite loop, waiting for exceptions
```

## exception_supervisor()
```
if running_pid != 0xFFFFFFFF
  Save the 11 registers on pcb_v[running_pid]
Jump to the specific handler based on the ECR
```

## clock_int() 
```
clock_interrupt_count++
if clock_interrupt_count < TIME_SLICE:
  MRET
clock_interrupt = 0

schedule()
```

## input_int()
```
for pid in range(0, partition_number):
  if !pcb_v[pid].is_mapped:
    //initialize the PCB
    is_mapped = 1
    is_waited = 0
    scheduler_state = ready
    parent_pid = 0xFFFFFFFF
    BASE = KERNEL_MAX_MEMORY + pid * PARTICION_SIZE
    LIMIT = BASE + PARTICION_SIZE
    SP = PARTICION_SIZE
    //All other registers are 0
    Copy the buffer to the user memory, starting at BASE
    put_queue(pid)
    schedule()

MRET
```

## kill_int()
```
pid = buffer
if pid >= partition_number or !pcb_v[pid].is_mapped:
  MRET

kill(pid)
```

## syscall_int()
```
Jumps to specific handler based on W5
```

## fault_int()
```
kill(running_pid)
```

## kill(pid)
```
pcb_v[pid].is_mapped = 0

if pcb_v[pid].is_waited:
  parent_pid = pcb_v[pid].parent_pid
  pcb_v[parent_pid].scheduler_state = ready
  put_queue(parent_pid)

if pid = running_pid:
  running_pid = 0xFFFFFFFF
else if pcb_v[pid].scheduler_state = ready:
  remove_queue(pid)

schedule()
```


# Data Structures
- Each partition has the fixed size of PARTITION_SIZE bytes
- There will be at most partition_number process stored on memory
- Each PID will be defined as a 4 byte unsigned integer

### Singular values
- **partition_number**: how many partitions there will be
- **running_pid**: PID of the current running process, or 0xFFFFFFFF if no process is running
- **waiting_queue_size**: number of processes currently on the ready queue
- **clock_interrupt_count**: number of clock interruptions since last scheduler call

### Constants (set by the OS developer)
- **PARTITION_SIZE**: how large a partition is
- **TIME_SLICE**: defined as the amount of clock interrupts to trigger the scheduler
- **KERNEL_MAX_MEMORY**: how much memory the kernel occupies, code + data structures
- **BUFFER**: size of the buffer in

### PCB
- **flags**: (1 byte)
  - **is_mapped**: 1 bit [0] (informs if this position in the PCB vector correspond to a process)
  - **is_waited**: 1 bit [1]
  - **scheduler_state (running, ready, blocked)**: 2 bits [2:3]
- **parent_pid**: 4 bytes
- **registers**: 11 registers, 4 bytes each, 44 bytes total
  - **W0**
  - **W1**
  - **W2**
  - **W3**
  - **W4**
  - **W5**
  - **PC**
  - **SP**
  - **SR**
  - **BASE**
  - **LIMIT**
- **total_size**: 4 + 1 + 44 + 3 (padding) = 52 bytes

### Queue
TODO

