# sackOS Abstract Implementation

## setup
```
user_memory = LIMIT - (KERNEL_MAX_MEMORY + BUFFER)
partition_number = USER_MEMORY / PARTITION_SIZE

for pid in range(0, partition_number):
  pcb_v[pid].BASE = KERNEL_MAX_MEMORY + pid * PARTITION_SIZE
  pcb_v[pid].LIMIT = BASE + PARTITION_SIZE

SP = SP_ADDRESS
ESA = EXCEPTION_SUPERVISOR_ADDRESS
EPC = LOOP ADDRESS
ESR = 16 //will enable interruptions
MRET //go to infinite loop, waiting for exceptions
```

# Interruption handlers

## exception_supervisor()
```
Save W1 and W2 into scratch space
if ecr == clock_int:
  clock_int()

if running_pid != 0xFFFFFFFF
  Save the 11 registers on pcb_v[running_pid] //remember W1 and W2 in scratch space
Jump to the specific handler based on the ECR
```

## clock_int() 
```
clock_interrupt_count++
if clock_interrupt_count < TIME_SLICE:
  load W1 and W2 back from scratch space
  MRET

clock_interrupt = 0
if running_pid != 0xFFFFFFFF
  Save the 11 registers on pcb_v[running_pid] //remember W1 and W2 in scratch space

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
    SP = PARTITION_SIZE
    //All other registers are 0, except BASE and LIMIT
    Copy the buffer to the user memory, starting at BASE
    break

schedule()
```

## kill_int()
```
pid = buffer
if pid >= partition_number or !pcb_v[pid].is_mapped:
  schedule()

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

# Syscall handlers

## fork()
```
for pid in range(0, partition_number):
  if !pcb_v[pid].is_mapped:
    //initialize the PCB
    is_mapped = 1
    is_waited = 0
    scheduler_state = ready
    parent_pid = running_pid
    w5 of running_pid = pid
    w5 of pid = 0
    //All other registers except BASE and LIMIT are copied from pcb_v[running_pid]
    Copy the running_pid memory to the pid memory, starting at BASE
    break

schedule()
```

## wait(pid)
```
if pid >= partition_number or !pcb_v[pid].is_mapped or pcb_v[pid].parent_pid != running_pid:
  pcb_v[running_pid].w5 = 1
  schedule()

pcb_v[running_pid].w5 = 0
pcb_v[pid].is_waited = 1
pcb_v[running_pid].scheduler_state = blocked
schedule()
```

## exit()
```
kill(running_pid)
```

## getPID()
```
pcb[running_pid].w5 = running_pid
schedule()
```

## rele()
```
schedule()
```

# Helper functions

## kill(pid)
```
pcb_v[pid].is_mapped = 0

if pcb_v[pid].is_waited:
  parent_pid = pcb_v[pid].parent_pid
  pcb_v[parent_pid].scheduler_state = ready

schedule()
```

## schedule()
```
if running_pid >= partition_number:
  running_pid = 0

if pcb_v[running_pid].scheduler_state = running:
  pcb_v[running_pid].scheduler_state = ready

limit_pid = running_pid

curr_pid = running_pid+1
if curr_pid = partition_number:
  curr_pid = 0

do:
  if pcb_v[curr_pid].is_mapped and pcb_v[curr_pid].scheduler_state = ready:
    pcb_v[curr_pid].scheduler_state = running
    running_pid = curr_pid
    load every register of pcb_v[running_pid] into the cpu
    mret
  
  curr_pid++
  if curr_pid = partition_number:
    curr_pid = 0

while curr_pid != limit_pid

running_pid = 0xFFFFFFFF
EPC = LOOP ADDRESS
ESR = 16 //will enable interruptions
MRET //go to infinite loop, waiting for exceptions
```

# Data Structures
- Each partition has the fixed size of PARTITION_SIZE bytes
- There will be at most partition_number process stored on memory
- Each PID will be defined as a 4 byte unsigned integer

### Singular values (32 bits)
- **partition_number**: how many partitions there will be
- **running_pid**: PID of the current running process, or 0xFFFFFFFF if no process is running
- **clock_interrupt_count**: number of clock interruptions since last scheduler call
- **kernel_stack_pointer**: the address of the kernel stackpointer

### Constants (set by the OS developer) (32 bits)
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

Similar to the Xv6 scheduler, the PCB vector is the queue itself, with the scheduler iterating it continuously until it finds a ready process.

# Addresses
TODO
