setup:
    MV W0 #2048  
    MV W1 #4096
    MUL W1, W0, W1   ; 8MB of full kernel memory
    STRD W1 #0x1008 ; kernel_max_memory
    STRD W1, #0x1020 ; initilizes the kernel stack pointer as the kernel_max_memory
    LDD W2 #0x100C ; buffer_size
    STRD LIMIT #0x1010 ; memory_size
    ADD W3 W1 W2 ; W3 -> KERNEL_MAX_MEMORY + BUFFER

    SUB W0 LIMIT W3 ; W0 -> USER_MEMORY = limit - (KERNEL_MAX_MEMORY + BUFFER)
    LDD W2 #0x1000 ; partition_size

    UDIV W3 W0 W2 ; W3 -> NUM_PARTITIONS = user_memory / PARTITION_SIZE
    STRD W3, #0x1014 ; partition_number

    MV W0, #-1
    STRD W0, #0x1018 ; running_pid = -1

    MV W4 #0x102C ; pcb_v

    MV W8 #72
    ADD W4 W4 W8 ; gets W4 to pcb_v[0].BASE address

   ; W1 will acumulate PARTITION_SIZE * i through the loop

    MV W6 #0 ; W6 -> pid = 0
    ; for (int pid = 0, pid < NUM_PARTITIONS; pid++)

    CMP W6 W3 
    BEQ setup_registers ; loop conditions

        STORE W1 W4 ; pcb_v[pid].BASE = KERNEL_MAX_MEMORY + pid * PARTITION_SIZE
        
        MV W8 #4
        ADD W4 W4 W8 ; W4 = pcb_v[pid].LIMIT address

        ADD W7 W1 W2 ; W7 = BASE + PARTITION_SIZE

        STORE W7 W4 ; pcb_v[pid].LIMIT = BASE + PARTITION_SIZE

        MV W8 #1
        ADD W6 W6 W8 ; pid += 1 


        ; setup variables to next Iteration

        MV W8 #80
        ADD W4 W4 W8 ; W4 = pcb_v[pid+1].BASE address
        ADD W1 W1 W2 ; W1 += PARTITION_SIZE

        JUMP #-12


        setup_registers:
            LDD SP #0x1020 ; kernel_stack_pointer
            MV ESA #0x94 ; the exception_supervisor initial address
            MV EPC #0x90 ; the infinite loop below
            MV ESR #16 ; enable interruptions
            
            MRET ; go to infinite loop, waiting for program inputs

JUMP #0

exception_supervisor:
    STRD W0 #0x1024 ; scratch_space_0
    STRD W1 #0x1028 ; scratch_space_1

    MV W0 #0
    CMP ECR W0 ; if ECR = 0 (clock interruption)
    BEQ clock_int

    MV W0 #-1
    LDD W1 #0x1018 ; running_pid

    CMP W0 W1
    BEQ jumpToHandler

    MV W0 #84 ; pcb_size
    MUL W1 W1 W0 ; W1 = RUNNING_PID * pcb_size

    MV W0 #0x102C ; pcb_v
    ADD W0 W0 W1 ; W0 = pcb[RUNNING_PID] address

    MV W1 #28
    ADD W0 W0 W1 ;  W0 = pcb[RUNNING_PID].w2 address

    MV W1 #4 ; uses W1 to jump to next register address
    ; save registers

    STORE W2 W0
    ADD W0 W1 W0

    STORE W3 W0
    ADD W0 W1 W0

    STORE W4 W0
    ADD W0 W1 W0

    STORE W5 W0
    ADD W0 W1 W0

    STORE W6 W0
    ADD W0 W1 W0

    STORE W7 W0
    ADD W0 W1 W0

    STORE W8 W0
    ADD W0 W1 W0

    STORE W9 W0
    ADD W0 W1 W0

    STORE EPC W0
    ADD W0 W1 W0

    STORE SP W0
    ADD W0 W1 W0

    STORE ESR W0
  
    MV W3 #48
    SUB W0 W0 W3 ; w0 points to pcb[RUNNING_PID].w0

    LDD W2 #0x1024 ; scratch_space_0

    STORE W2 W0

    MV W8 #4
    ADD W0 W0 W8 ; w0 points to pcb[RUNNING_PID].w1

    LDD W2 #0x1028 ; scratch_space_1
    STORE W2 W0

    jumpToHandler:
        MV W0 #1
        MV W1 #1 

        CMP ECR W0
        BEQ input_int

        ADD W0 W0 W1

        CMP ECR W0
        BEQ kill_int

        ADD W0 W0 W1 
        
        CMP ECR W0
        BEQ syscall_int

        ADD W0 W0 W1 
        
        CMP ECR W0
        BEQ fault_int    


; INTERRUPTIONS


clock_int:
  LDD W0, #0x101C    ; w0 = clock_interrupt_count

  MV W1, #1
  ADD W0, W0, W1                    ; clock_interrupt_count += 1

  LDD W1, #0x1004 ; time_slice 
  CMP W0, W1
  BEQ clock_reset                   ; if clock_interrupt_count == TIME_SLICE, reset and check running

    STRD W0, #0x101C ; clock_interrupt_count
  clock_return:
    LDD W0, #0x1024 ; scratch_space_0
    LDD W1, #0x1028 ; scratch_space_1
    MRET

  clock_reset:
    MV W0, #0
    STRD W0, #0x101C ; clock_interrupt_count = 0

  LDD W0, #0x1018 ; running_pid
  MV W1, #-1
  CMP W0, W1
  BEQ clock_return                   ; if running_pid == -1, clock_return

  ; saving registers
  LDD W1, #0x1018 ; running_pid    
  MV W0, #84                      ; pcb_size
  MUL W1, W0, W1                  ; W1 = running_pid * pcb_size
  
  MV W0, #0x102C ; pcb_v
  ADD W1, W0, W1                  ; W1 = pcb_v[running_pid] initial address

  MV W0, #68
  ADD W1, W0, W1
  STORE ESR, W1                   ; saving SR

  MV W0, #4
  SUB W1, W1, W0
  STORE SP, W1                    ; saving SP

  SUB W1, W1, W0
  STORE EPC, W1                   ; saving PC

  SUB W1, W1, W0
  STORE W9, W1                    ; saving W9

  SUB W1, W1, W0
  STORE W8, W1                    ; saving W8

  SUB W1, W1, W0
  STORE W7, W1                    ; saving W7
  
  SUB W1, W1, W0
  STORE W6, W1                    ; saving W6
  
  SUB W1, W1, W0
  STORE W5, W1                    ; saving W5

  SUB W1, W1, W0
  STORE W4, W1                    ; saving W4

  SUB W1, W1, W0
  STORE W3, W1                    ; saving W3

  SUB W1, W1, W0
  STORE W2, W1                    ; saving W2

  LDD W3, #0x1028 ; scratch_space_1
  SUB W1, W1, W0
  STORE W3, W1                    ; saving W1

  LDD W3, #0x1024 ; scratch_space_0
  SUB W1, W1, W0
  STORE W3, W1                    ; saving W0

  JUMP schedule


input_int:
    ; Get buffer_input_size from buffer
    ; BUFFER_START = memory_size - BUFFER_SIZE
    LDD W3, #MEMORY_SIZE_ADDR        ; W3 = memory_size
    LDD W2, #BUFFER_SIZE_ADDR        ; W2 = BUFFER_SIZE
    SUB W3, W3, W2                   ; W3 = BUFFER_START address
    LOAD W3, W3                      ; W3 = buffer_input_size (first 4 bytes of buffer)

    ; Save buffer_input_size to scratch_space_0 for later use
    STRD W3, #SCRATCH_SPACE_0_ADDR   ; scratch_space_0 = buffer_input_size

    ; Search for a free PCB slot
    ; for pid in range(0, MAX_PROCESSES)
    MV W4, #0                        ; W4 = pid = 0
    LDD W5, #MAX_PROCESSES_ADDR      ; W5 = MAX_PROCESSES

    input_loop:
        CMP W4, W5
        BEQ input_end                ; No free slot found

        ; W0 = pcb_size
        MV W0, #PCB_SIZE

        ; Compute &pcb_v[pid]
        MUL W1, W4, W0               ; W1 = pid * pcb_size
        MV W0, #PCB_V_ADDR           ; W0 = pcb_v base address
        ADD W0, W0, W1               ; W0 = &pcb_v[pid]

        ; Read flags at offset 72
        MV W1, #72
        ADD W8, W0, W1               ; W8 = &pcb_v[pid].flags
        LDB W6, W8                   ; W6 = flags byte

        ; Check is_mapped (bit 0)
        MV W1, #1
        AND W6, W6, W1               ; isolate bit 0
        MV W1, #0
        CMP W6, W1
        BGT next_input_iteration     ; is_mapped == 1, try next pid

        ; Found free slot - Initialize PCB
        ; W0 = &pcb_v[pid], W4 = pid, W8 = &flags

        ; flags = 0b00001001 = 9 (is_mapped=1, scheduler_state=ready)
        MV W6, #9
        STRB W6, W8                  ; pcb_v[pid].flags = 9

        MV W7, #-1                   ; W7 = -1 for parent, child, siblings, status_addr

        ; parent_pid = -1 (offset 0)
        STORE W7, W0

        ; child = -1 (offset 4)
        MV W1, #4
        ADD W8, W0, W1
        STORE W7, W8

        ; prev_sibling = -1 (offset 8)
        MV W1, #8
        ADD W8, W0, W1
        STORE W7, W8

        ; next_sibling = -1 (offset 12)
        MV W1, #12
        ADD W8, W0, W1
        STORE W7, W8

        ; status_addr = -1 (offset 16)
        MV W1, #16
        ADD W8, W0, W1
        STORE W7, W8

        ; pages_used = 0 (offset 76)
        MV W7, #0
        MV W1, #76
        ADD W8, W0, W1
        STORE W7, W8

        ; Initialize registers to 0
        MV W7, #0

        MV W1, #20
        ADD W8, W0, W1
        STORE W7, W8          ; pcb.W0 = 0

        MV W1, #24
        ADD W8, W0, W1
        STORE W7, W8          ; pcb.W1 = 0

        MV W1, #28
        ADD W8, W0, W1
        STORE W7, W8          ; pcb.W2 = 0

        MV W1, #32
        ADD W8, W0, W1
        STORE W7, W8          ; pcb.W3 = 0

        MV W1, #36
        ADD W8, W0, W1
        STORE W7, W8          ; pcb.W4 = 0

        MV W1, #40
        ADD W8, W0, W1
        STORE W7, W8          ; pcb.W5 = 0

        MV W1, #44
        ADD W8, W0, W1
        STORE W7, W8          ; pcb.W6 = 0

        MV W1, #48
        ADD W8, W0, W1
        STORE W7, W8          ; pcb.W7 = 0

        MV W1, #52
        ADD W8, W0, W1
        STORE W7, W8          ; pcb.W8 = 0

        MV W1, #56
        ADD W8, W0, W1
        STORE W7, W8          ; pcb.W9 = 0

        MV W1, #60
        ADD W8, W0, W1
        STORE W7, W8          ; pcb.PC = 0

        ; SP = 3GB = 0xC0000000 (offset 64)
        MV W7, #3
        MV W1, #1024
        MUL W7, W7, W1
        MUL W7, W7, W1
        MUL W7, W7, W1        ; W7 = 3GB = 0xC0000000

        MV W1, #64
        ADD W8, W0, W1
        STORE W7, W8          ; pcb.SP = 3GB

        ; SR = 56 (offset 68)
        MV W7, #56
        MV W1, #68
        ADD W8, W0, W1
        STORE W7, W8          ; pcb.SR = 56

        ; Set UPTR to page table of new process

        MV W1, #80
        ADD W1, W0, W1                 ; W1 = physical address of page table
        STRD W1, #SCRATCH_SPACE_1_ADDR ; scratch_space_1 = page table addr
        LDD UPTR, #SCRATCH_SPACE_1_ADDR; UPTR = page table addr

        ; Map pages for the program code

        LDD W3, #SCRATCH_SPACE_0_ADDR  ; W3 = buffer_input_size
        MV W2, #4095
        ADD W3, W3, W2                 ; W3 = buffer_input_size + 4095
        MV W2, #4096
        UDIV W3, W3, W2                ; W3 = num_pages = ceil(buffer_input_size / 4096)

        ; Save pcb_v[pid] base (W0) and pid (W4) to kernel stack
        LDD SP, #KERNEL_STACK_POINTER_ADDR
        MV W1, #4
        SUB SP, SP, W1
        STORE W0, SP                   ; push pcb_v[pid] base
        SUB SP, SP, W1
        STORE W4, SP                   ; push pid

        MV W2, #0                      ; W2 = page_id = 0

    input_map_loop:
        CMP W2, W3
        BEQ input_map_done             ; All pages mapped successfully

        ; Push page_id (W2) and num_pages (W3) before map_page call
        MV W1, #4
        SUB SP, SP, W1
        STORE W3, SP                   ; push num_pages
        SUB SP, SP, W1
        STORE W2, SP                   ; push page_id

        ; Setup map_page call: W9 = page_id, W8 = return address
        MV W1, #0
        ADD W9, W2, W1                 ; W9 = page_id

        MV W1, #4
        ADD W8, PC, W1                 ; W8 = return address
        JUMP map_page

    input_map_return:
        ; Pop page_id and num_pages
        MV W1, #4
        LOAD W2, SP                    ; W2 = page_id
        ADD SP, SP, W1
        LOAD W3, SP                    ; W3 = num_pages
        ADD SP, SP, W1

        ; Check map_page result (W9: 0=success, 1=no free frame)
        MV W1, #0
        CMP W9, W1
        BEQ input_map_ok

        ; ===== MAP FAILED: kill(pid) with exit code 3 =====
        MV W1, #4
        LOAD W4, SP                    ; W4 = pid
        ADD SP, SP, W1
        LOAD W0, SP                    ; W0 = pcb_v[pid] base
        ADD SP, SP, W1

        ; pcb_v[pid].W9 = 3 (page fault exit code)
        MV W1, #56
        ADD W0, W0, W1
        MV W1, #3
        STORE W1, W0                   ; pcb_v[pid].W9 = 3

        MV W1, #0
        ADD W9, W4, W1                 ; W9 = pid
        JUMP kill

    input_map_ok:
        MV W1, #1
        ADD W2, W2, W1
        JUMP input_map_loop

    input_map_done:
        MV W1, #4
        LOAD W4, SP                    ; W4 = pid
        ADD SP, SP, W1
        LOAD W0, SP                    ; W0 = pcb_v[pid] base
        ADD SP, SP, W1

        ; Copy buffer content to user virtual memory
        LDD W3, #MEMORY_SIZE_ADDR      ; W3 = memory_size
        LDD W2, #BUFFER_SIZE_ADDR      ; W2 = BUFFER_SIZE
        SUB W3, W3, W2                 ; W3 = BUFFER_START
        MV W2, #4
        ADD W3, W3, W2                 ; W3 = buffer data start (skip size word)

        LDD W6, #SCRATCH_SPACE_0_ADDR  ; W6 = buffer_input_size
        MV W8, #0                      ; W8 = index = 0 (user virtual address)
        MV W2, #4                      ; W2 = step

    input_copy_loop:
        CMP W8, W6
        BEQ input_copy_end
        BGT input_copy_end

        ADD W7, W3, W8                 ; W7 = buffer_data_start + index
        LOAD W5, W7                    ; W5 = buffer[index]

        STORE W5, W8                   ; memory[user_virtual_addr] = W5

        ADD W8, W8, W2                 ; index += 4
        JUMP input_copy_loop

    input_copy_end:
        JUMP schedule

    next_input_iteration:
        MV W2, #1
        ADD W4, W4, W2                 ; pid++
        JUMP input_loop

    input_end:
        JUMP schedule


kill_int:

    LDD W9, #0x1010 ; memory_size
    LDD W7, #0x100C ; buffer_size
    SUB W9, W9, W7  ; initial buffer addr
    LOAD W9, W9     ; pid of kill_int

    LDD W1, #0x1014 ; partition_number

    CMP W9, W1
    BGT not_valid
    BEQ not_valid

    MV  W2, #84
    MUL W3, W9, W2
    MV W4, #0x102C ; pcb_v
    ADD W3, W3, W4

    MV  W5, #80
    ADD W5, W3, W5
    LDB W6, W5

    MV  W8, #1
    AND W7, W6, W8

    MV  W0, #0
    CMP W7, W0
    BEQ not_valid

    MV  W8, #2
    AND W7, W6, W8

    CMP W7, W8
    BEQ not_valid

    MV  W5, #56
    ADD W5, W3, W5

    MV  W6, #2
    STORE W6, W5

    JUMP kill

    not_valid:
        JUMP schedule


syscall_int:
    MV W1, #0
    CMP W9, W1
    BEQ fork

    MV W1, #1
    CMP W9, W1
    BEQ wait

    MV W1, #2
    CMP W9, W1
    BEQ exit

    MV W1, #3
    CMP W9, W1
    BEQ getPID

    MV W1, #4
    CMP W9, W1
    BEQ rele

    ; Invalid Syscall: pcb_v[running_pid].w9 = -1
    MV W0, #3145808             ; W0 = pcb_size
    LDD W1, #RUNNING_PID_ADDR   ; W1 = running_pid
    MUL W1, W1, W0              ; W1 = running_pid * pcb_size
    MV W0, #PCB_V_ADDR          ; W0 = pcb_v base address
    ADD W0, W0, W1              ; W0 = &pcb_v[running_pid]

    MV W3, #56                  ; offset of W9 in PCB
    ADD W0, W0, W3              ; W0 = &pcb_v[running_pid].w9

    MV W3, #-1                  ; value -1
    STORE W3, W0                ; pcb_v[running_pid].w9 = -1

    JUMP schedule



fault_int:
  MV W0, #0x102C       ; pcb_v initial address
  LDD W9, #0x1018      ; running_pid
  MV W1, #84           ; pcb_size

  MUL W1, W9, W1       ; W1 = running_pid * pcb_size
  ADD W0, W0, W1       ; W0 = pcb_v[running_pid] initial address

  MV W1, #56           ; pcb W9 offset
  ADD W0, W0, W1       ; W0 = pcb_v[running_pid].W9 initial address

  MV W1, #1
  STORE W1, W0         ; pcb_v[running_pid].W9 = 1

  JUMP kill            ; kill(running_pid)




; SYSCALLS

fork:

    ; Search for a free PCB slot
    ; for pid in range(0, MAX_PROCESSES)
    MV W4, #0                        ; W4 = pid = 0
    LDD W5, #MAX_PROCESSES_ADDR      ; W5 = MAX_PROCESSES

    fork_loop:
        CMP W4, W5
        BEQ fork_end                 ; No free slot found

        ; W0 = pcb_size
        MV W0, #PCB_SIZE

        ; Compute &pcb_v[pid]
        MUL W1, W4, W0
        MV W0, #PCB_V_ADDR
        ADD W0, W0, W1               ; W0 = &pcb_v[pid]

        ; Read flags (offset 72)
        MV W1, #72
        ADD W8, W0, W1
        LDB W6, W8

        ; Check is_mapped (bit 0)
        MV W1, #1
        AND W6, W6, W1
        MV W1, #0
        CMP W6, W1
        BGT next_iteration           ; is_mapped == 1, skip

        ; Initialize child PCB
        ; W0 = &pcb_v[pid], W4 = pid

        ; flags = 9 (is_mapped=1, scheduler_state=ready)
        MV W6, #9
        STRB W6, W8

        ; parent_pid = running_pid (offset 0)
        LDD W7, #RUNNING_PID_ADDR    ; W7 = running_pid
        STORE W7, W0

        MV W7, #-1

        ; child = -1 (offset 4)
        MV W1, #4
        ADD W8, W0, W1
        STORE W7, W8

        ; prev_sibling = -1 (offset 8)
        MV W1, #8
        ADD W8, W0, W1
        STORE W7, W8

        ; next_sibling = -1 (offset 12)
        MV W1, #12
        ADD W8, W0, W1
        STORE W7, W8

        ; status_addr = -1 (offset 16)
        MV W1, #16
        ADD W8, W0, W1
        STORE W7, W8

        ; pages_used = 0 (offset 76)
        MV W7, #0
        MV W1, #76
        ADD W8, W0, W1
        STORE W7, W8

        ; Compute parent PCB address
        ; W7 = &pcb_v[running_pid]

        LDD W1, #RUNNING_PID_ADDR    ; running_pid
        MV W3, #PCB_SIZE             ; pcb_size
        MUL W1, W1, W3
        MV W7, #PCB_V_ADDR
        ADD W7, W7, W1               ; W7 = &pcb_v[running_pid]

        ; Set fork return values
        ; pcb_v[running_pid].w9 = pid
        MV W1, #56
        ADD W8, W7, W1
        STORE W4, W8                  ; parent gets child pid

        ; pcb_v[pid].w9 = -2
        ADD W8, W0, W1                ; W1 still 56
        MV W1, #-2
        STORE W1, W8                  ; child gets -2

        ; Link sibling list
        ; pcb_v[pid].next_sibling = pcb_v[running_pid].child
        MV W1, #4
        ADD W8, W7, W1
        LOAD W1, W8                   ; W1 = old_child = pcb_v[running_pid].child

        MV W3, #12
        ADD W8, W0, W3
        STORE W1, W8                  ; pcb_v[pid].next_sibling = old_child

        ; if old_child != -1: pcb_v[old_child].prev_sibling = pid
        MV W3, #-1
        CMP W1, W3
        BEQ skip_prev_sibling

            ; Compute &pcb_v[old_child]
            MV W3, #PCB_SIZE          ; pcb_size
            MUL W1, W1, W3
            MV W3, #PCB_V_ADDR
            ADD W1, W1, W3            ; W1 = &pcb_v[old_child]

            MV W3, #8
            ADD W1, W1, W3
            STORE W4, W1              ; pcb_v[old_child].prev_sibling = pid

    skip_prev_sibling:
        ; pcb_v[running_pid].child = pid
        MV W1, #4
        ADD W8, W7, W1
        STORE W4, W8

        ; Copy registers from parent to child
        ; W0 = child PCB, W7 = parent PCB
        ; Loop offsets 20..68, skip offset 56 (W9 already set)
        ; Copies: W0-W8, PC, SP, SR (12 registers)

        MV W2, #20                    ; W2 = current offset
        MV W3, #72                    ; W3 = end (exclusive)
        MV W6, #4                     ; W6 = step

    fork_copy_regs_loop:
        CMP W2, W3
        BEQ fork_copy_regs_done

        ; Skip W9 at offset 56
        MV W1, #56
        CMP W2, W1
        BEQ fork_copy_regs_skip

        ; Copy: parent[offset] -> child[offset]
        ADD W8, W7, W2
        LOAD W1, W8
        ADD W8, W0, W2
        STORE W1, W8

    fork_copy_regs_skip:
        ADD W2, W2, W6                ; offset += 4
        JUMP fork_copy_regs_loop

    fork_copy_regs_done:

        ; Setup UPTR and store page table addresses

        ; Child PT = &pcb_v[pid] + 80
        MV W1, #80
        ADD W1, W0, W1
        STRD W1, #SCRATCH_SPACE_1_ADDR; scratch_space_1 = child PT
        LDD UPTR, #SCRATCH_SPACE_1_ADDR; UPTR = child PT

        ; Parent PT = &pcb_v[running_pid] + 80
        MV W1, #80
        ADD W1, W7, W1
        STRD W1, #SCRATCH_SPACE_0_ADDR; scratch_space_0 = parent PT

        ; Iterate parent page table
        ; For each valid PTE: map_page(pageId) + copy 4KB

        ; Save child PCB (W0), pid (W4), parent PCB (W7) to kernel stack
        LDD SP, #KERNEL_STACK_POINTER_ADDR
        MV W1, #4
        SUB SP, SP, W1
        STORE W0, SP                  ; push child PCB addr
        SUB SP, SP, W1
        STORE W4, SP                  ; push pid
        SUB SP, SP, W1
        STORE W7, SP                  ; push parent PCB addr

        ; Compute total PTEs = 786432
        MV W3, #768
        MV W1, #1024
        MUL W3, W3, W1               ; W3 = 786432

        ; W6 = parent PT base addr
        LDD W6, #SCRATCH_SPACE_0_ADDR

        MV W2, #0                     ; W2 = pte_index = 0

    fork_pte_loop:
        CMP W2, W3
        BEQ fork_pte_done

        ; Read PTE from parent page table
        MV W1, #4
        MUL W8, W2, W1
        ADD W8, W6, W8               ; W8 = &parent_pt[pte_index]
        LOAD W5, W8                  ; W5 = PTE value

        ; Check Valid bit (bit 0 of PTE)
        MV W1, #1
        AND W1, W5, W1
        MV W0, #0
        CMP W1, W0
        BEQ fork_pte_next            ; PTE not valid, skip

        ; PTE is valid: map page + copy content 

        ; Push pte_index (W2), total_ptes (W3), parent_pt_base (W6)
        MV W1, #4
        SUB SP, SP, W1
        STORE W6, SP
        SUB SP, SP, W1
        STORE W3, SP
        SUB SP, SP, W1
        STORE W2, SP

        ; Set UPTR to child PT for map_page to map into child's table
        LDD UPTR, #SCRATCH_SPACE_1_ADDR

        ; Call map_page: W9 = page_id = pte_index, W8 = return addr
        MV W1, #0
        ADD W9, W2, W1               ; W9 = page_id

        MV W1, #4
        ADD W8, PC, W1               ; W8 = return addr
        JUMP map_page

    fork_map_return:
        ; Pop pte_index, total_ptes, parent_pt_base
        MV W1, #4
        LOAD W2, SP
        ADD SP, SP, W1
        LOAD W3, SP
        ADD SP, SP, W1
        LOAD W6, SP
        ADD SP, SP, W1

        ; Check map_page result (W9: 0=ok, 1=fail)
        MV W1, #0
        CMP W9, W1
        BEQ fork_map_ok

        ; MAP FAILED: kill(pid) with exit code 3
        MV W1, #4
        LOAD W7, SP
        ADD SP, SP, W1
        LOAD W4, SP                   ; W4 = pid
        ADD SP, SP, W1
        LOAD W0, SP                   ; W0 = child PCB addr
        ADD SP, SP, W1

        ; pcb_v[pid].W9 = 3
        MV W1, #56
        ADD W0, W0, W1
        MV W1, #3
        STORE W1, W0

        ; W9 = pid for kill()
        MV W1, #0
        ADD W9, W4, W1
        JUMP kill

    fork_map_ok:
        ; Copy page content (4KB) from parent to child
        ; virtual_base = pte_index * 4096
        MV W1, #4096
        MUL W9, W2, W1               ; W9 = page virtual base addr

        ; Push pte_index (W2), total_ptes (W3), parent_pt_base (W6)
        MV W1, #4
        SUB SP, SP, W1
        STORE W6, SP
        SUB SP, SP, W1
        STORE W3, SP
        SUB SP, SP, W1
        STORE W2, SP

        MV W2, #0                     ; W2 = word_offset = 0
        MV W3, #4096                  ; W3 = page_size (4KB)
        MV W6, #4                     ; W6 = step

    fork_copy_page_loop:
        CMP W2, W3
        BEQ fork_copy_page_done

        ; W8 = virtual address to copy
        ADD W8, W9, W2

        ; Switch UPTR to parent -> read
        LDD UPTR, #SCRATCH_SPACE_0_ADDR
        LOAD W5, W8                   ; W5 = parent_page[offset]

        ; Switch UPTR to child -> write
        LDD UPTR, #SCRATCH_SPACE_1_ADDR
        STORE W5, W8                  ; child_page[offset] = W5

        ADD W2, W2, W6                ; offset += 4
        JUMP fork_copy_page_loop

    fork_copy_page_done:
        ; Pop pte_index, total_ptes, parent_pt_base
        MV W1, #4
        LOAD W2, SP
        ADD SP, SP, W1
        LOAD W3, SP
        ADD SP, SP, W1
        LOAD W6, SP
        ADD SP, SP, W1

    fork_pte_next:
        MV W1, #1
        ADD W2, W2, W1
        JUMP fork_pte_loop

    fork_pte_done:
        ; All parent PTEs processed. Pop parent PCB, pid, child PCB.
        MV W1, #4
        LOAD W7, SP
        ADD SP, SP, W1
        LOAD W4, SP
        ADD SP, SP, W1
        LOAD W0, SP
        ADD SP, SP, W1

        JUMP schedule

    next_iteration:
        MV W1, #1
        ADD W4, W4, W1               ; pid++
        JUMP fork_loop

    fork_end:
        ; No free slot: pcb_v[running_pid].w9 = -1
        MV W0, #PCB_SIZE             ; pcb_size
        LDD W1, #RUNNING_PID_ADDR    ; running_pid
        MUL W1, W1, W0
        MV W0, #PCB_V_ADDR
        ADD W0, W0, W1               ; &pcb_v[running_pid]

        MV W1, #56
        ADD W0, W0, W1               ; &pcb_v[running_pid].w9
        MV W1, #-1
        STORE W1, W0                  ; pcb_v[running_pid].w9 = -1
        JUMP schedule

wait:
    MV W0 #0x102C ; w0 points to pcb_v[0] first byte
    MV W8 #52 
    ADD W0 W8 ; w0 points to pcb_v[0].w8
    MV W8 #84 ; bytes size of each pcb

    LDD W1 #0x1018 ; running_pid
    MUL W1 W1 W8 ; W1 = RUNNING_PID * 84 bytes
    ADD W0 W0 W1 ; w0 points to pcb_v[RUNNING_PID].w8
    
    LOAD W9 W0 ; w9 = status_addr

    MV W8 #20
    ADD W0 W8 ; w0 points to pcb_v[running_pid].BASE

    LOAD W2 W0 ; w2 = pcb_v[RUNNING_PID].BASE
    
    MV W8 #4
    ADD W0 W0 W8 ; w0 points to pcb_v[RUNNING_PID].LIMIT
    
    LOAD W3 W0 ; w3 = pcb_v[RUNNING_PID].LIMIT

    ADD W7 W2 W9 ; w7 = pcb_v[running_pid].BASE + status_addr
    ADD W2 W2 W9 ; w2 = pcb_v[running_pid].BASE + status_addr
    
    MV W8 #4
    ADD W2 W2 W8 ; w2 = pcb_v[running_pid].BASE + status_addr + 4


    ; if  pcb_v[running_pid].BASE + status_addr + 4 > pcb_v[RUNNING_PID].LIMIT
    ; fault_int()


    CMP W2 W3
    BGT fault_int 


    MV W8 #72
    SUB W2 W0 W8 ; w2 points to pcb_v[RUNNING_PID].child 
    LOAD W6 W2 ; w6 = pcb_v[RUNNING_PID].child 

    MV W8 #-1 
    CMP W6 W8 ; if pcb_v[RUNNING_PID].child  == -1:
    BEQ #2

    JUMP #6

        MV W8 #20
        SUB W3 W0 W8 ; w3 points to pcb_v[RUNNING_PID].w9
        MV W8 #-1
        STORE W8 W3
        JUMP schedule

    LOAD W3 W2 ; w3 = curr_child = pcb_v[running_pid].childPID
    MV W0 #0x102C ;  w0 points to pcb_v[0] first byte

    waitLoop:
        ;calculate pcb_v[curr_child]
        ; w3 = curr_child

        MV W8 #84
        MUL W1 W3 W8

        ADD W1 W0 W1 ;W1 points to pcb_v[curr_child] first byte

        MV W8 #80
        ADD W1 W1 W8 ; w1 points to pcb_v[curr_child].flags
        LDB W6 W1 ; w6 = pcb_v[curr_child].flags

        MV W8 #2 ;0b00000000000000000000000000000010
        OR W8 W8 W6
        CMP W6 W8
        BEQ curr_child_is_zombie 

            ;if !pcb_v[curr_child].is_zombie:
            ; curr_child = pcb_v[curr_child].next_sibling

            MV W8 #68
            SUB W3 W1 W8 ; w3 points to pcb_v[curr_child].next_siblingPID
            LOAD W6 W3 ; w6 = pcb_v[curr_child].next_siblingPID

            MV W8 #-1
            CMP W6 W8 ; if pcb_v[curr_child].next_siblingPID == -1
            BEQ setWaitingStatus

            LOAD W3 W3 ; w3 = curr_child = pcb_v[curr_child].next_siblingPID
            JUMP waitLoop ; goes to next Iteration

        ; else (pcb_v[curr_child].is_zombie)

        curr_child_is_zombie:
            MV W8  #-2 ;0b11111111111111111111111111111110
            LDB W6 W1
            AND W8 W8 W6 
            STRB W8 W1 ; sets pcb_v[curr_child].is_mapped = 0


            
            ; w0 points to pcb_v[0] first byte
            ; w1 points to pcb_v[curr_child].flags
            ; w2 points to pcb_v[running_pid].child
            ; w3 = curr_child
            ; w7 = pcb_v[running_pid].BASE + status_addr

            MV W8 #52
            ADD W4 W2 W8 ; w4 points to pcb_v[running_pid].w9
            STORE W3 W4 ; pcb_v[running_pid].w9 = curr_child

            MV W8 #24 
            SUB W1 W1 W8 ; w1 points to pcb_v[curr_child].w9
            LOAD W6 W1 ; w6 = pcb_v[curr_child].w9

            STORE W6 W7 ; memory[pcb_v[running_pid].BASE+status_addr] = pcb_v[curr_child].w9

            MV W8 #44
            SUB W1 W1 W8 ; w1 points to pcb_v[curr_child].next_sibling

            LOAD W6 W2 ; w6 = pcb_v[running_pid].child
            CMP W6 W3 ; if curr_child == pcb_v[running_pid].child
            BEQ #2

            JUMP curr_child_is_not_first_children

            LOAD W6 W1 ; w6 = pcb_v[curr_child].next_sibling
            STORE W6 W2 ; pcb_v[running_pid].child = pcb_v[curr_child].next_sibling

            JUMP link_next_sibling
            
        

        curr_child_is_not_first_children:
            MV W8 #4
            SUB W4 W1 W8 ; w4 points to pcb_v[curr_child].prev_sibling
            LOAD W7 W4   ; w7 = pcb_v[curr_child].prev_sibling

            MV W6 #84 ; 84 bytes each pcb
            MUL W5 W7 W6 ; 
            ADD W5 W0 W5 ; w5 points to pcb_v[pcb_v[running_pid].prev_sibling] first byte

            MV W8 #12
            ADD W5 W5 W8 ; w5 points to pcb_v[pcb_v[running_pid].prev_sibling].next_sibling

            LOAD W7 W1 ; w7 = pcb_v[curr_child].next_sibling
            STORE W7 W5 ; pcb_v[pcb_v[running_pid].prev_sibling].next_sibling = pcb_v[curr_child].next_sibling

        
            ; else
        link_next_sibling:


            ; w1 points to pcb_v[curr_child].next_sibling
            LOAD W7 W1 ; w7 = pcb_v[curr_child].next_sibling
            MV W8 #-1
            CMP W7 W8 ; if pcb_v[curr_child].next_sibling == -1 
            BEQ schedule

            MV W6 #84 ; 84 bytes each pcb
            MUL W5 W7 W6 
            ADD W5 W0 W5 ; w5 points to pcb_v[pcb_v[running_pid].next_sibling] first byte

            MV W8 #8
            ADD W5 W5 W8 ; w5 points to pcb_v[pcb_v[running_pid].next_sibling].prev_sibling

            MV W8 #4
            SUB W1 W1 W8 ; w1 = pcb_v[curr_child].prev_sibling
            LOAD W8 W1
            STORE W8 W5 ; pcb_v[pcb_v[running_pid].next_sibling].prev_sibling = pcb_v[curr_child].prev_sibling

            JUMP schedule

    setWaitingStatus:
        ; w2 = pcb_v[running_pid].child

        MV W8 #12
        ADD W2 W2 W8 ; w2 = pcb_v[running_pid].status_add
        STORE W9 W2 ; pcb_v[running_pid].status_addr = status_addr

        MV W8 #64
        ADD W2 W2 W8 ; w2 points to pcb_v[running_pid].flags

        LDB W4 W2 ; w4 = pcb_v[running_pid].flags

        ; pcb_v[running_pid].scheduler_state = blocked
        ; pcb_v[running_pid].is_waiting = 1
        ; same to set flags byte to  0bXXX101XX

        
        ; set 1's 
        MV W8 #20 ; 0b00010100
        OR W4 W4 W8

        ; set 0
        MV W8 #-9 ; 0b11110111
        AND W4 W4 W8

        STRB W4 W2 ;

        JUMP schedule


exit:
	LDD W9, #0x1018        ; W9 = running_pid
	MV W1, #84               ; W1 = pcb_size
	MUL W0, W9, W1           ; W0 = running_pid * pcb_size

	MV W1, #0x102C           ; W1 = pcb_v initial address
	ADD W0, W0, W1           ; W0 = pcb_v[running_pid] initial address

	MV W1, #52
	ADD W0, W0, W1           ; W0 = pcb_v[running_pid].W8 address
	LOAD W2, W0              ; W2 = pcb_v[running_pid].W8 (status_code)

	MV W1, #4
	ADD W0, W0, W1           ; W0 = pcb_v[running_pid].W9 address
	STORE W2, W0             ; pcb_v[running_pid].W9 = pcb_v[running_pid].W8

	JUMP kill


getPID:
    LDD W0 #0x1018 ; running_pid
    MV W1 #0x102C ; pcb_v

    MV W8 #84
    MUL W2 W0 W8 ; RUNNING_PID * 84 get the offset of bytes to acess pcb[RUNNING_PID]
    
    ADD W1 W1 W2 ; stores on W1 the first address of pcb[RUNNING_PID]

    MV W8 #56
    ADD W1 W1 W8 ; W1 = pcb[RUNNING_PID].w9 address bytes

    STORE W0 W1 ; stores RUNNING_PID on pcb[RUNNING_PID].w9

    JUMP schedule


rele:
    JUMP schedule


; HELPERS


kill:
  MV W0, #0x102C          ; W0 = pcb_v initial address
  MV W1, #84              ; W1 = pcb_size
  MUL W2, W1, W9
  ADD W2, W0, W2          ; W2 = pcb_v[pid].parent_pid address

  LOAD W3, W2             ; W3 = pcb_v[pid].parent_pid

  MV W4, #-1
  CMP W3, W4
  BEQ kill_no_parent      ; if parent != -1

    MUL W3, W3, W1
    ADD W3, W3, W0        ; W3 = pcb_v[parent] initial address

    MV W4, #80
    ADD W3, W3, W4        ; W3 = pcb_v[parent].flags address

    LDB W4, W3           ; W4 = pcb_v[parent].flags
    MV W5, #4
    AND W4, W4, W5        ; W4 = pcb_v[parent].is_waiting

    CMP W4, W5
    BEQ #2                ; if pcb_v[parent].is_waiting
    JUMP kill_not_waiting

      MV W4, #9           ; mapped = 1, zombie = 0, waiting = 0, state = ready
      STRB W4, W3

      MV W4, #8
      SUB W3, W3, W4      ; W3 = pcb_v[parent].BASE address
      LOAD W4, W3         ; W4 = pcb_v[parent].BASE

      MV W5, #56
      SUB W3, W3, W5      ; W3 = pcb_v[parent].status_addr address
      LOAD W5, W3         ; W5 = pcb_v[parent].status_addr

      ADD W4, W4, W5      ; W4 = base+status_addr address

      MV W5, #56
      ADD W2, W2, W5      ; W2 = pcb_v[pid].W9 address
      LOAD W5, W2         ; W5 = pcb_v[pid].W9

      STORE W5, W4        ; memory[pcb_v[parent].BASE+pcb_v[parent].status_addr] = pcb_v[pid].w9

      MV W5, #12
      SUB W3, W3, W5      ; W3 = pcb_v[parent].child address
      LOAD W4, W3         ; W4 = pcb_v[parent].child

      CMP W4, W9
      BEQ #2              ; if pcb_v[parent].child = pid
      JUMP #6

        MV W5, #44
        SUB W2, W2, W5    ; W2 = pcb_v[pid].next_sibling address
        LOAD W4, W2       ; W4 = pcb_v[pid].next_sibling
        STORE W4, W3      ; pcb_v[parent].child = pcb_v[pid].next_sibling
        ADD W2, W2, W5    ; W2 = pcb_v[pid].W9 address


      MV W5, #48
      SUB W2, W2, W5      ; W2 = pcb_v[pid].prev_sibling address
      LOAD W3, W2         ; W3 = pcb_v[pid].prev_sibling

      MV W5, #4
      ADD W2, W2, W5      ; W2 = pcb_v[pid].next_sibling address
      LOAD W4, W2         ; W4 = pcb_v[pid].next_sibling


      MV W5, #-1
      CMP W3, W5
      BEQ #6              ; if pcb_v[pid].prev_sibling != -1

        MUL W5, W3, W1
        ADD W5, W5, W0    ; W5 = pcb_v[pcb_v[pid].prev_sibling] initial address
        MV W6, #12
        ADD W5, W5, W6    ; W5 = pcb_v[pcb_v[pid].prev_sibling].next_sibling address

        STORE W4, W5      ; pcb_v[pcb_v[pid].prev_sibling].next_sibling = pcb_v[pid].next_sibling


      MV W5, #-1
      CMP W4, W5
      BEQ #6              ; if pcb_v[pid].next_sibling != -1

        MUL W5, W4, W1
        ADD W5, W5, W0    ; W5 = pcb_v[pcb_v[pid].next_sibling] initial address
        MV W6, #8
        ADD W5, W5, W6    ; W5 = pcb_v[pcb_v[pid].next_sibling].prev_sibling address

        STORE W3, W5      ; pcb_v[pcb_v[pid].next_sibling].prev_sibling = pcb_v[pid].prev_sibling

      
      MV W3, #68
      ADD W2, W2, W3      ; W2 = pcb_v[pid].flags address
      LDB W3, W2         ; W3 = pcb_v[pid].flags

      MV W4, #-2          ; everything 1 except the LSB
      AND W3, W3, W4      ; is_mapped = 0
      STRB W3, W2        ; pcb_v[pid].is_mapped = 0

      JUMP kill_orphanize ; execution after the elses

    kill_not_waiting:
      ; here W2 = pcb_v[pid].parent_pid address
      MV W3, #80
      ADD W2, W2, W3      ; W2 = pcb_v[pid].flags address
      LDB W3, W2         ; W3 = pcb_v[pid].flags

      MV W4, #2
      OR W3, W3, W4       ; is_zombie = 1
      STRB W3, W2        ; pcb_v[pid].is_zombie = 1
      JUMP kill_orphanize ; execution after the elses


  kill_no_parent:
    ; here W2 = pcb_v[pid].parent_pid address
    MV W3, #80
    ADD W2, W2, W3      ; W2 = pcb_v[pid].flags address
    LDB W3, W2         ; W3 = pcb_v[pid].flags

    MV W4, #-2          ; everything 1 except the LSB
    AND W3, W3, W4      ; is_mapped = 0
    STRB W3, W2        ; pcb_v[pid].is_mapped = 0

  
  kill_orphanize:
    ; here W2 = pcb_v[pid].flags address

    MV W3, #76
    SUB W2, W2, W3      ; W2 = pcb_v[pid].child address
    LOAD W3, W2          ; W3 = pcb_v[pid].child (curr_child)
    MV W7, #-1

    orphanize_loop:
      CMP W3, W7
      BEQ kill_end

      MUL W4, W3, W1
      ADD W4, W4, W0    ; W4 = pcb_v[curr_child].parent_pid address
      STORE W7, W4      ; pcb_v[curr_child].parent_pid = -1

      MV W6, #8
      ADD W4, W4, W6    ; W4 = pcb_v[curr_child].prev_sibling address
      STORE W7, W4      ; pcb_v[curr_child].prev_sibling = -1

      MV W6, #4
      ADD W4, W4, W6    ; W4 = pcb_v[curr_child].next_sibling address
      LOAD W5, W4       ; W5 = pcb_v[curr_child].next_sibling = next
      STORE W7, W4      ; pcb_v[curr_child].next_sibling = -1

      MV W8, #0
      ADD W3, W5, W8    ; W3 = curr_child = next
      JUMP orphanize_loop

  kill_end:
    JUMP schedule



schedule:

    ; valida running_pid

    LDD W9, #0x1018 ; running_pid

    MV W0, #-1
    CMP W9, W0
    BEQ schedule_reset_running_pid

    ; valida running_pid

    LDD W0, #0x1014 ; partition_number

    CMP W9, W0
    BEQ schedule_reset_running_pid
    BGT schedule_reset_running_pid

    JUMP schedule_check_current_process


    schedule_reset_running_pid:

        MV W9, #0


    schedule_check_current_process:

        ; verifica se processo atual esta running
        ; W1 = pcb_v[running_pid]

        MV W0, #84
        MUL W1, W9, W0

        MV W0, #0x102C ; pcb_v
        ADD W1, W0, W1

        ; W2 = pcb_v[running_pid].flags

        MV W0, #80
        ADD W0, W1, W0

        LDB W2, W0

        MV W5, #0x18
        AND W5, W2, W5

        MV W0, #0
        CMP W5, W0
        BEQ schedule_set_current_ready

        JUMP schedule_calculate_limit_pid


    schedule_set_current_ready:

        ; marca como ready

        MV W5, #-25
        AND W2, W2, W5

        MV W5, #0x08
        OR W2, W2, W5

        MV W0, #80
        ADD W0, W1, W0

        STRB W2, W0


    schedule_calculate_limit_pid:

        ; limit_pid = running_pid + 1

        MV W0, #1
        ADD W3, W9, W0

        ; wrap limit_pid se necessario

        LDD W0, #0x1014 ; partition_number

        CMP W3, W0
        BEQ schedule_limit_pid_zero

        JUMP schedule_limit_pid_ready


    schedule_limit_pid_zero:

        MV W3, #0


    schedule_limit_pid_ready:

        ; curr_pid = limit_pid

        MV W4, #0
        ADD W4, W4, W3

    schedule_search_loop:

        ; busca processo ready
        ; W1 = &pcb_v[curr_pid]

        MV W0, #84
        MUL W1, W4, W0

        MV W0, #0x102C ; pcb_v
        ADD W1, W0, W1

        ; W2 = pcb_v[curr_pid].flags

        MV W0, #80
        ADD W0, W1, W0

        LDB W2, W0

        ; esta mapeado?

        MV W5, #0x01
        AND W5, W2, W5

        MV W0, #0x01
        CMP W5, W0
        BEQ schedule_check_zombie

        JUMP schedule_next_pid


    schedule_check_zombie:

        ; eh zombie?

        MV W5, #0x02
        AND W5, W2, W5

        MV W0, #0
        CMP W5, W0
        BEQ schedule_check_ready

        JUMP schedule_next_pid


    schedule_check_ready:

        ; esta ready?

        MV W5, #0x18
        AND W5, W2, W5

        MV W0, #0x08
        CMP W5, W0
        BEQ schedule_process_found

        JUMP schedule_next_pid


    schedule_process_found:

        ; processo encontrado, seta running e restaura

        MV W5, #-25
        AND W2, W2, W5

        MV W0, #80
        ADD W0, W1, W0

        STRB W2, W0

        ; W9 = curr_pid

        MV W9, #0
        ADD W9, W9, W4
        STRD W9, #0x1018 ; running_pid = curr_pid

        ; W6 = &pcb_v[running_pid]

        MV W6, #0
        ADD W6, W6, W1

        JUMP schedule_restore_context


    schedule_next_pid:

        ; curr_pid++

        MV W0, #1
        ADD W4, W4, W0

        ; wrap curr_pid se necessario

        LDD W0, #0x1014 ; partition_number

        CMP W4, W0
        BEQ schedule_wrap_pid

        JUMP schedule_check_search_end


    schedule_wrap_pid:

        MV W4, #0


    schedule_check_search_end:

        ; while curr_pid != limit_pid

        CMP W4, W3
        BEQ schedule_no_ready_process

        JUMP schedule_search_loop


    ; precisa de futuro review a partir daqui
    schedule_no_ready_process:

        ; nenhum pronto, idle loop

        MV W9, #-1
        STRD W9, #0x1018 ; running_pid = -1

        MV EPC, #0x90 ; infinite_loop

        MV ESR, #16

        MRET


    ; W6 = endereco inicial do PCB selecionado
    ; W9 = running_pid (nao restaurado do PCB)
    schedule_restore_context:

        ; carrega registros do PCB
        ; W0 (20) -> scratch

        MV W0, #20
        ADD W8, W6, W0

        LOAD W0, W8

        STRD W0, #0x1024 ; scratch_space_0

        ; PC (60)

        MV W0, #60
        ADD W8, W6, W0

        LOAD EPC, W8

        ; SP (64)

        MV W0, #64
        ADD W8, W6, W0

        LOAD SP, W8

        ; SR (68)

        MV W0, #68
        ADD W8, W6, W0

        LOAD ESR, W8

        ; BASE (72)

        MV W0, #72
        ADD W8, W6, W0

        LOAD BASE, W8

        ; LIMIT (76)

        MV W0, #76
        ADD W8, W6, W0

        LOAD LIMIT, W8

        ; W1 (24)

        MV W0, #24
        ADD W8, W6, W0

        LOAD W1, W8

        ; W2 (28)

        MV W0, #28
        ADD W8, W6, W0

        LOAD W2, W8

        ; W3 (32)

        MV W0, #32
        ADD W8, W6, W0

        LOAD W3, W8

        ; W4 (36)

        MV W0, #36
        ADD W8, W6, W0

        LOAD W4, W8

        ; W5 (40)

        MV W0, #40
        ADD W8, W6, W0

        LOAD W5, W8

        ; W7 (48) -- restaura antes de W6 (W6 ainda tem endereco do PCB)

        MV W0, #48
        ADD W8, W6, W0

        LOAD W7, W8

        ; W9 (56) -- restaura antes de W8 (W8 ainda tem endereco do registrador)

        MV W0, #56
        ADD W8, W6, W0

        LOAD W9, W8

        ; W8 (52) -- usa W0 como temporario para nao perder o endereco

        MV W0, #52
        ADD W0, W6, W0

        LOAD W8, W0

        ; W6 (44) -- restaurado por ultimo entre W1-W9

        MV W0, #44
        ADD W0, W6, W0

        LOAD W6, W0

        ; W0 <- scratch (ultimo registrador antes do MRET)

        LDD W0, 0x1024 ; scratch_space_0

        MRET