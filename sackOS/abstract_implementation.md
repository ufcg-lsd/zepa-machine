# sackOS Abstract Implementation

## setup
```
user_memory = LIMIT - (KERNEL_MAX_MEMORY + BUFFER)
MAX_PROCESSES = USER_MEMORY / PARTITION_SIZE

for pid in range(0, MAX_PROCESSES):
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
Save W0 and W1 into scratch space
if ecr == clock_int:
  clock_int()

if running_pid != -1
  Save all registers on pcb_v[running_pid] //remember W0 and W1 in scratch space
Jump to the specific handler based on the ECR
```

## clock_int() 
```
clock_interrupt_count++
if clock_interrupt_count < TIME_SLICE:
  load W0 and W1 back from scratch space
  MRET

clock_interrupt_count = 0

if running_pid == -1:
  load W0 and W1 back from scratch space
  MRET

Save all registers on pcb_v[running_pid] //remember W0 and W1 in scratch space

schedule()
```

## input_int()
```
Extract the size from buffer and schedule() if its bigger than the partition

for pid in range(0, MAX_PROCESSES):
  if pcb_v[pid].is_mapped = 0:
    //initialize the PCB
    pcb_v[pid].is_mapped = 1
    pcb_v[pid].is_zombie = 0
    pcb_v[pid].is_waiting = 0
    pcb_v[pid].scheduler_state = ready
    pcb_v[pid].parent_pid = -1
    pcb_v[pid].child = -1
    pcb_v[pid].next_sibling = -1
    pcb_v[pid].prev_sibling = -1
    pcb_v[pid].status_addr = -1
    pcb_v[pid].pages_used = 0

    SP = 3GB (user virtual memory limit)
    //All other registers are 0
    Copy the buffer (BUFFER bytes) to the user memory, (load PTR)
    break

schedule()
```

## kill_int()
```
pid = buffer
if pid >= MAX_PROCESSES or pcb_v[pid].is_mapped = 0 or pcb_v[pid].is_zombie = 1:
  schedule()

pcb_v[pid].w9 = 2
kill(pid)
```

## syscall_int()
```
Jumps to specific handler based on W9
```

## fault_int()
```
pcb_v[running_pid].w9 = 1
kill(running_pid)
```

## page_fault_int()
```
  if efa >= kernelBoundary and esr not in kernel mode:
    pcb_v[running_pid].w9 = 3
    kill(running_pid)

  search the bitmap for a free frame, map the PTE to the free frame, based on efa, set valid = 1
  set the pcb_v[running_process].PC to an instruction before what it is in the EPC
  update pcb_v[running_process].pages_used

  if there is no free frame:
  pcb_v[running_pid].w9 = 3
  kill(running_pid)

  schedule()
```

# Syscall handlers

## fork() - ID 0
```
for pid in range(0, MAX_PROCESSES):
  if pcb_v[pid].is_mapped = 0:
    //initialize the PCB
    pcb_v[pid].is_mapped = 1
    pcb_v[pid].is_zombie = 0
    pcb_v[pid].is_waiting = 0
    pcb_v[pid].scheduler_state = ready
    pcb_v[pid].parent_pid = running_pid
    pcb_v[pid].child = -1
    pcb_v[pid].next_sibling = -1
    pcb_v[pid].prev_sibling = -1
    pcb_v[pid].status_addr = -1
    pcb_v[pid].pages_used = 0
    
    
    pcb_v[running_pid].w9 = pid
    pcb_v[pid].w9 = -2

    pcb_v[pid].next_sibling = pcb_v[running_pid].child
    if pcb_v[running_pid].child != -1:
      pcb_v[pcb_v[running_pid].child].prev_sibling = pid
    pcb_v[running_pid].child = pid

    //All other registers are copied from pcb_v[running_pid]
    Copy the running_pid memory to the pid memory (switch between parent PTR and child PTR to copy)
    for pte in pcb_v[running_pid].page_table:
      if pte.Valid:
        copy the page content from pte.page to the child page
      
    schedule()

pcb_v[running_pid].w9 = -1
schedule()
```

## wait(status_addr) - ID 1
```
if status_addr >= 0xC0000000:
  fault_int()

if pcb_v[running_pid].child == -1:
  pcb_v[running_pid].w9 = -1
  schedule()
else:
  curr_child = pcb_v[running_pid].child
  do:
    if pcb_v[curr_child].is_zombie:
      pcb_v[curr_child].is_mapped = 0
      pcb_v[running_pid].w9 = curr_child
      memory[status_addr] = pcb_v[curr_child].w9

      if pcb_v[running_pid].child = curr_child:
        pcb_v[running_pid].child = pcb_v[curr_child].next_sibling
      else:
        pcb_v[pcb_v[curr_child].prev_sibling].next_sibling = pcb_v[curr_child].next_sibling

      if pcb_v[curr_child].next_sibling != -1:
        pcb_v[pcb_v[curr_child].next_sibling].prev_sibling = pcb_v[curr_child].prev_sibling
      
      schedule()

    else:
      curr_child = pcb_v[curr_child].next_sibling

  while curr_child != -1

  pcb_v[running_pid].status_addr = status_addr
  pcb_v[running_pid].scheduler_state = blocked
  pcb_v[running_pid].is_waiting = 1
  schedule()
```

## exit(status_code) - ID 2
```
pcb_v[running_pid].w9 = status_code
kill(running_pid)
```

## getPID() - ID 3
```
pcb[running_pid].w9 = running_pid
schedule() 
```

## rele() - ID 4
```
schedule()
```

# Helper functions

## kill(pid)
```
parent = pcb_v[pid].parent_pid
if parent != -1:
  if pcb_v[parent].is_waiting:
    pcb_v[parent].is_waiting = 0
    pcb_v[parent].scheduler_state = ready
    pcb_v[parent].w9 = pid
    memory[pcb_v[parent].status_addr] = pcb_v[pid].w9

    if pcb_v[parent].child = pid:
      pcb_v[parent].child = pcb_v[pid].next_sibling

    if pcb_v[pid].prev_sibling != -1:
      pcb_v[pcb_v[pid].prev_sibling].next_sibling = pcb_v[pid].next_sibling

    if pcb_v[pid].next_sibling != -1:
      pcb_v[pcb_v[pid].next_sibling].prev_sibling = pcb_v[pid].prev_sibling

    pcb_v[pid].is_mapped = 0
  
  else:
    pcb_v[pid].is_zombie = 1

else:
  pcb_v[pid].is_mapped = 0

curr_child = pcb_v[pid].child
while curr_child != -1:
  pcb_v[curr_child].parent_pid = -1
  pcb_v[curr_child].prev_sibling = -1
  next = pcb_v[curr_child].next_sibling
  pcb_v[curr_child].next_sibling = -1
  curr_child = next

free every stack page cleaning the pages from the end of memory till the stack pointer address page 
free every page from page table couting till pcb_v[pid].pages_used[i] == 0

schedule()
```

## schedule()
```
if running_pid >= MAX_PROCESSES:
  running_pid = 0

if pcb_v[running_pid].scheduler_state = running:
  pcb_v[running_pid].scheduler_state = ready

limit_pid = running_pid+1
if limit_pid = MAX_PROCESSES:
  limit_pid = 0

curr_pid = limit_pid

do:
  if pcb_v[curr_pid].is_mapped = 1 and pcb_v[curr_pid].is_zombie = 0 and pcb_v[curr_pid].scheduler_state = ready:
    pcb_v[curr_pid].scheduler_state = running
    running_pid = curr_pid
    load every register of pcb_v[running_pid] into the cpu 
    UPTR = pcb_v[running_pid].page_table address
    mret
  
  curr_pid++
  if curr_pid == MAX_PROCESSES:
    curr_pid = 0

while curr_pid != limit_pid

running_pid = -1
EPC = LOOP ADDRESS
ESR = 48 //will enable interruptions
MRET //go to infinite loop, waiting for exceptions
```

# Data Structures
- Each Page has the fixed size of 4KB
- PTE has 32bit 
- The virtual memory is divided in 3GB(3/4) to the user and 1GB(1/4) for the kernel
- There will be at most MAX_PROCESSES process stored on memory
- Each PID will be defined as a 4 byte unsigned integer

### Constants (set by the OS developer) (32 bits)

- **MAX_PROCESSES**: maximum number of processes
- **TIME_SLICE**: defined as the amount of clock interrupts to trigger the scheduler
- **BUFFER_SIZE**: size of the input buffer 

### Singular values (32 bits)
- **memory_size**: how much memory is there available
- **running_pid**: PID of the current running process, or -1 if no process is running
- **clock_interrupt_count**: number of clock interruptions since last scheduler call
- **kernel_stack_pointer**: points to the current kernel stack
- **scratch_space_0**: aux address to temporarily save W0 on interruptions
- **scratch_space_1**: aux address to temporarily save W1 on interruptions

### PCB
- **parent_pid**: 4 bytes [0]
- **child**: 4 bytes [4]
- **prev_sibling**: 4 bytes [8]
- **next_sibling**: 4 bytes [12]
- **status_addr**: 4 bytes [16]
- **registers**: 13 registers, 4 bytes each, 52 bytes total
  - **W0** [20]
  - **W1** [24]
  - **W2** [28]
  - **W3** [32]
  - **W4** [36]
  - **W5** [40]
  - **W6** [44]
  - **W7** [48]
  - **W8** [52]
  - **W9** [56]
  - **PC** [60]
  - **SP** [64]
  - **SR** [68]
- **flags**: (1 byte) [72]
  - **is_mapped**: 1 bit [0] (informs if this position in the PCB vector correspond to a process)
  - **is_zombie**: 1 bit [1]
  - **is_waiting**: 1 bit [2]
  - **scheduler_state (running, ready, blocked)**: 2 bits [3:4]
  **Pages_Used**: 4 bytes [76]
  **PAGE TABLE**: (3 MB)
  - 3*2^18 PTEs, each PTE has 4 bytes
- **total_size**: 3MB + 80 bytes = 3145808 bytes 

### Bitmap

Array that will tell whether a page frame is being used or not [2^20 bits = 128 KB]

### Queue

Similar to the Xv6 scheduler, the PCB vector is the queue itself, with the scheduler iterating it continuously until it finds a ready process.

# Addresses
The kernel code has 833 instructions as of now, resulting in 3332 bytes of memory, we rounded it to 4KB, so addresses will start at 0x1000

### Constants (set by the OS developer) (32 bits)
- **MAX_PROCESSES**: 0x1000
- **TIME_SLICE**: 0x1004
- **BUFFER_SIZE**: 0x1008

### Singular values (32 bits)
- **memory_size**: 0x100C
- **running_pid**: 0x1010
- **clock_interrupt_count**: 0x1014
- **kernel_stack_pointer**: 0x1018
- **scratch_space_0**: 0x101C
- **scratch_space_1**: 0x1020

### Data Structures
- **pcb_vector**: 0x1024