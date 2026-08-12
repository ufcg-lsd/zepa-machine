setup:
    STRD W0, #0x200C ; memory_size low addr

    MV UPTR, #0x2024    ; pcb_v low addr
    MV W1, #80           ; page_table offset

    ADD UPTR, UPTR, W1       ; UPTR = pcb_v[0].page_table address

    MV W1, #20
    MV W9, #1
    SHL W9, W9, W1           ; W9 = sets bit 20 (valid) to 1
    STORE W9, UPTR           ; identity mapping, allocates the first frame to the first page

    MV W0, #0x3000
    MV W1, #16
    SHL W0, W0, W1

    MV W1, #0x7024
    ADD W0, W0, W1       ; W0 = 0x30007024 = kernel_page_table low first address

    MV W2, #0
    ADD KPTR, W0, W2     ; KPTR = 0x30007024 = kernel_page_table low first address

    MV W1, #1
    MV W2, #20
    SHL W1, W1, W2           ; W1 = 2^20, the size of the kernel page_table

    ADD W1, W1, W0           ; W1 = the end of the kernel page_table

    MV W2, #4                ; W2 = byte step
    MV W3, #1                ; W3 = one step
    setup_kernel_mapping_loop:
        CMP W0, W1
        BEQ setup_end_kernel_mapping

        STORE W9, W0      ; maps the high address page to the low frame
        ADD W0, W0, W2    ; W0 = next pte address
        ADD W9, W9, W3        ; W9 = next page frame id
        JUMP setup_kernel_mapping_loop


    setup_end_kernel_mapping:

    MV ESR, #32     ; enable MMU
    MV EPC, #4
    ADD EPC, EPC, PC ; EPC points to next instruction, since fetch already updates PC

    MRET

    ; now using virtual addresses

    MV W9, #3
    MV W1, #30
    SHL W9, W9, W1       ; W9 = 3GB mark
    
    MV W1, #4
    ADD W0, W9, W1       ; W0 = 3GB mark + 4 bytes
    ADD W0, W0, PC       ; W0 points to the instruction after the jump, but on high addresses
    JMPR W0

    ; now in high addresses
    MV W0 #0
    ADD W1, W9, UPTR   ; W1 = uptr virtual address
    STORE W0, W1       ; unmaps the identity, sets the pcb_v[0].page_table[0] to invalid

    MV W0, #0x3010
    MV W1, #16
    SHL W0, W0, W1
    MV W1, #0x7024
    ADD W0, W0, W1     ; W0 = physical bitmap addr
    ADD W0, W0, W9     ; W0 = virtual bitmap addr

    MV W1, #0x8000     ; W1 = 2^18 frames / 8 bits = 2^15 bytes of the bitmap to populate
    ADD W1, W1, W0     ; W1 = the end of the bitmap populate portion
    
    MV W2, #-1         ; W2 = all ones, to populate the bitmap
    MV W3, #4          ; byte step

    setup_populate_bitmap_loop:
        CMP W0, W1
        BEQ setup_end_populate_bitmap

        STORE W2, W0   ; populate the bitmap

        ADD W0, W0, W3 ; W0 goes to the next byte
        JUMP setup_populate_bitmap_loop
    
    setup_end_populate_bitmap:

    MV W0, #-1
    MV W1, #0x2010
    ADD W1, W1, W9        ; W1 = running_pid virtual addr
    STORE W0, W1          ; running_pid = -1

    MV SP, #0
    MV W0, #0x2008
    ADD W0, W0, W9                 ; W0 = buffer_size virtual addr
    LOAD W0, W0

    SUB SP, SP, W0                 ; SP = 4GB - buffer_size
    MV W0, #0x2018
    ADD W0, W0, W9                 ; W0 = kernel_stack_pointer virtual addr
    STORE SP, W0                   ; sets kernel_stack_pointer


    MV W0, #8       ; 2 instructions offset 

    MV ESR, #48     ; enable interruptions and mmu
    ADD EPC, PC, W0 ; the infinite loop below
    ADD ESA, PC, W0 ; the exception_supervisor initial address
    
    MRET ; go to infinite loop, waiting for program inputs

JUMP #0

exception_supervisor:
    ; store w0 and w1 in scratch_space

    MV K0 #3
    MV K1 #30
    SHL K1 K0 K1          ; K1 = 3 << 30 = 3GB

    MV K0 #0x201C
    ADD K1 K1 K0 ; k1 points to scratch_space_0
    STORE W0 K1 ; saves w0 in scratch_space_0

    MV W0 #4
    ADD K1 K1 W0 ; k1 points to scratch_space_1
    STORE W1 K1 ; saves w1 in scratch_space_1

    MV W0 #0
    CMP ECR W0 ; if ECR = 0 (clock interruption)
    BEQ clock_int

    MV W0 #3
    MV W1 #30
    SHL K0 W0 W1 ; K0 = 3GB

    MV W1 #0x2010 ; 
    ADD W1 K0 W1 ;
    LOAD W1 W1 ; w1 = running_pid 

    MV W0 #-1

    CMP W0 W1
    BEQ jumpToHandler


    MV W0 #3
    MV K1 #20
    SHL W0 W0 K1          ; W0 = 3 << 20 = 0x300000 (3MB)
    MV K1 #0x50
    ADD W0 W0 K1          ; W0 = 0x300050 = 3145808 (3MB + 80)
    MUL W1 W1 W0 ; W1 = RUNNING_PID * pcb_size

    MV W0 #0x2024 ; 
    ADD W0 K0 W0 ; w0 = pcb_v_addr
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

    MV W2 #0x201C
    ADD W2 W2 K0          ; w2 = 3GB + 0x201C = scratch_space_0
    LOAD W2 W2            ; w2 = w0 salvo
    STORE W2 W0           ; pcb[running_pid].w0 = w0

    MV W8 #4
    ADD W0 W0 W8 ; w0 points to pcb[RUNNING_PID].w1

    MV W2 #0x2020
    ADD W2 W2 K0          ; w2 = 3GB + 0x2020 = scratch_space_1
    LOAD W2 W2            ; w2 = w1 salvo
    STORE W2 W0           ; pcb[running_pid].w1 = w1

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

        ADD W0 W0 W1

        CMP ECR W0
        BEQ page_fault_exc


; INTERRUPTIONS


clock_int:
  LDD W0, #clock_interrupt_count    ; w0 = clock_interrupt_count

  MV W1, #1
  ADD W0, W0, W1                    ; clock_interrupt_count += 1

  LDD W1, #time_slice ; time_slice 
  CMP W0, W1
  BEQ clock_reset                   ; if clock_interrupt_count == TIME_SLICE, reset and check running

    STRD W0, #clock_interrupt_count ; clock_interrupt_count
  clock_return:
    LDD W0, #scratch_space_0 ; scratch_space_0
    LDD W1, #scratch_space_1 ; scratch_space_1
    MRET

  clock_reset:
    MV W0, #0
    STRD W0, #clock_interrupt_count ; clock_interrupt_count = 0

  LDD W0, #running_pid ; running_pid
  MV W1, #-1
  CMP W0, W1
  BEQ clock_return                   ; if running_pid == -1, clock_return

  ; saving registers
  LDD W1, #running_pid ; running_pid    
  MV W0, #0x300050                      ; pcb_size
  MUL W1, W0, W1                  ; W1 = running_pid * pcb_size
  
  MV W0, #pcb_v ; pcb_v
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

  LDD W3, #scratch_space_1 ; scratch_space_1
  SUB W1, W1, W0
  STORE W3, W1                    ; saving W1

  LDD W3, #scratch_space_0 ; scratch_space_0
  SUB W1, W1, W0
  STORE W3, W1                    ; saving W0

  JUMP schedule


input_int:
    ; K0 = kernel boundary = 0xC0000000
    MV K0, #0xC000
    MV W0, #16
    SHL K0, K0, W0

    ; Get buffer_input_size from buffer
    ; BUFFER_START = memory_size - BUFFER_SIZE
    MV W3, #0x200C
    ADD W3, K0, W3
    LOAD W3, W3                      ; W3 = memory_size

    MV W2, #0x2008
    ADD W2, K0, W2
    LOAD W2, W2                      ; W2 = BUFFER_SIZE

    SUB W3, W3, W2                   ; W3 = BUFFER_START address
    LOAD W3, W3                      ; W3 = buffer_input_size (first 4 bytes of buffer)

    ; Save buffer_input_size to scratch_space_0 for later use
    MV W1, #0x201C
    ADD W1, K0, W1
    STORE W3, W1                     ; scratch_space_0 = buffer_input_size

    ; Search for a free PCB slot
    ; for pid in range(0, MAX_PROCESSES)
    MV W4, #0                        ; W4 = pid = 0
    MV W5, #0x2000
    ADD W5, K0, W5
    LOAD W5, W5                      ; W5 = MAX_PROCESSES

    input_loop:
        CMP W4, W5
        BEQ input_end                ; No free slot found

        ; W0 = pcb_size (0x00300050)
        MV W0, #0x30
        MV W1, #16
        SHL W0, W0, W1
        MV W1, #0x50
        ADD W0, W0, W1

        ; Compute &pcb_v[pid]
        MUL W1, W4, W0               ; W1 = pid * pcb_size
        MV W8, #0x2024
        ADD W0, K0, W8               ; W0 = 0xC0002024 (pcb_v virtual base address)
        ADD W0, W0, W1               ; W0 = &pcb_v[pid] (virtual address)

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
        ADD W1, W0, W1                 ; W1 = virtual address of page table
        MV UPTR, #0xC0000000
        SUB UPTR, W1, UPTR             ; UPTR = physical address of page table

        ; Map pages for the program code

        MV W3, #0x201C
        ADD W3, K0, W3
        LOAD W3, W3                    ; W3 = buffer_input_size

        MV W2, #4095
        ADD W3, W3, W2                 ; W3 = buffer_input_size + 4095
        MV W2, #4096
        UDIV W3, W3, W2                ; W3 = num_pages = ceil(buffer_input_size / 4096)

        ; Save pcb_v[pid] base (W0) and pid (W4) to kernel stack
        MV W1, #0x2018
        ADD W1, K0, W1
        LOAD SP, W1                    ; SP = kernel_stack_pointer
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
        MV W3, #0x200C
        ADD W3, K0, W3
        LOAD W3, W3                    ; W3 = memory_size

        MV W2, #0x2008
        ADD W2, K0, W2
        LOAD W2, W2                    ; W2 = BUFFER_SIZE

        SUB W3, W3, W2                 ; W3 = BUFFER_START
        MV W2, #4
        ADD W3, W3, W2                 ; W3 = buffer data start (skip size word)

        MV W6, #0x201C
        ADD W6, K0, W6
        LOAD W6, W6                    ; W6 = buffer_input_size
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
        ; W8 = buffer_input_size
        ; Pad memory with zeros until the end of the 4KB boundary.

        MV W5, #4095
        AND W5, W8, W5                 ; W5 = W8 % 4096 (bytes used in last page)
        
        MV W6, #0
        CMP W5, W6
        BEQ input_schedule             ; If aligned to 4096, no padding

        MV W6, #4096
        SUB W6, W6, W5                 ; W6 = remaining bytes to zero out
        ADD W6, W8, W6                 ; W6 = exact virtual address where padding stops
        
        MV W5, #0                      ; W5 = 0 (Data to write)
        MV W2, #4                      ; Step

    input_pad_loop:
        CMP W8, W6
        BGT input_schedule
        BEQ input_schedule

        STORE W5, W8                   ; memory[user_virtual_addr] = 0
        ADD W8, W8, W2
        JUMP input_pad_loop

    input_schedule:
        JUMP schedule

    next_input_iteration:
        MV W2, #1
        ADD W4, W4, W2                 ; pid++
        JUMP input_loop

    input_end:
        JUMP schedule


kill_int:
    ; Obtém o início físico/offset do buffer
    LDD W9, #0x200C ; #0x200C memory_size
    LDD W7, #0x2008 ; #0x2008 BUFFER_SIZE
    SUB W9, W9, W7

    ; +0 = tamanho
    ; +4 = PID
    MV  W6, #4
    ADD W9, W9, W6
    LOAD W9, [W9]

    LDD W1, #0x2000 ; #0x2000 MAX_PROCESSES (verifica pid)

    CMP W9, W1
    BGT not_valid
    BEQ not_valid

    ; PCB_SIZE = 0x00300050
    MV  W2, #0x30
    MV  W0, #16
    SHL W2, W2, W0

    MV  W0, #0x50
    ADD W2, W2, W0

    ; &pcb_v[pid]
    MUL W3, W9, W2

    MV  W4, #0x2024 ; #0x2024 pcb_vector
    ADD W3, W3, W4

    ; flags no offset 72
    MV  W5, #72
    ADD W5, W3, W5
    LDB W6, [W5]

    ; is_mapped
    MV  W8, #1
    AND W7, W6, W8

    MV  W0, #0
    CMP W7, W0
    BEQ not_valid

    ; is_zombie
    MV  W8, #2
    AND W7, W6, W8

    CMP W7, W8
    BEQ not_valid

    ; pcb_v[pid].w9 = 2
    MV  W5, #56
    ADD W5, W3, W5

    MV  W6, #2
    STORE W6, [W5]

    ; W9 continua contendo o PID
    JUMP kill

not_valid:
    JUMP schedule

syscall_int:
    ; K0 = kernel boundary = 0xC0000000
    MV K0, #0xC000
    MV W1, #16
    SHL K0, K0, W1

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
    MV W0, #0x30
    MV W1, #16
    SHL W0, W0, W1
    
    MV W1, #0x50
    ADD W0, W0, W1              ; W0 = pcb_size (0x300050)
    
    MV W8, #0x2010
    ADD W8, K0, W8              ; W8 = 0xC0002010 (&running_pid)
    LOAD W1, W8                 ; W1 = running_pid
    MUL W1, W1, W0              ; W1 = running_pid * pcb_size

    MV W8, #0x2024
    ADD W0, K0, W8              ; W0 = 0xC0002024 (pcb_v base address)
    ADD W0, W0, W1              ; W0 = &pcb_v[running_pid]

    MV W3, #56                  ; offset of W9 in PCB
    ADD W0, W0, W3              ; W0 = &pcb_v[running_pid].w9

    MV W3, #-1                  ; value -1
    STORE W3, W0                ; pcb_v[running_pid].w9 = -1

    JUMP schedule



fault_int:
  MV W8, #3
  MV W7, #30
  SHL W8, W8, W7        ; W8 = 3GB kernel offset

  MV W0, #0x2024        ; pcb_v initial physical address
  ADD W0, W0, W8        ; pcb_v initial virtual address

  MV W9, #0x2010        ; running_pid physical address
  ADD W9, W9, W8        ; W9 = running_pid virtual address
  LOAD W9, W9           ; W9 = running_pid

  MV W1, #0x30
  MV W2, #16
  SHL W1, W1, W2
  MV W2, #0x50
  ADD W1, W1, W2        ; pcb_size

  MUL W1, W9, W1       ; W1 = running_pid * pcb_size
  ADD W0, W0, W1       ; W0 = pcb_v[running_pid] initial address

  MV W1, #56           ; pcb W9 offset
  ADD W0, W0, W1       ; W0 = pcb_v[running_pid].W9 initial address

  MV W1, #1
  STORE W1, W0         ; pcb_v[running_pid].W9 = 1

  JUMP kill            ; kill(running_pid)

; exceptions

page_fault_exc:
    ; ==========================================================
    ; 1. Verificação de Segurança (Kernel Boundary)
    ; ==========================================================

    MV W0, #3
    MV W2, #30
    SHL W0, W0, W2                 ; W0 = 3 << 30 = 0xC0000000
    
    CMP EFA, W0
    BLT page_fault_valid_address   ; Se EFA < 0xC0000000, é endereço de usuário válido
    
    ; Se EFA >= 0xC0000000, checa se ESR indica modo usuário (bit 3 == 1)
    MV W1, #8                      ; bit 3 (0b1000)
    AND W2, ESR, W1
    CMP W2, W1
    BEQ page_fault_kill            ; Se for modo usuário e acessou kernel, mata o processo

page_fault_valid_address:
    ; ==========================================================
    ; 2. Cálculo do page_id e chamada do map_page
    ; ==========================================================
    ; page_id = EFA >> 12 (Em ZEPA, SHL negativo faz shift right)
    MV W1, #-12
    SHL W9, EFA, W1
    
    MV W0 #12
    ADD W0 PC W0 ; W0 = return_from_map_page

    MV W1 #0
    ADD W8 W0 W1  ; Endereço de retorno exigido pela map_page
    JUMP map_page

return_from_map_page:
    ; W9 contém o retorno de map_page (0 = sucesso)
    MV W0, #0
    CMP W9, W0
    BEQ page_fault_clear_page      ; Se sucesso, vai limpar a página

page_fault_kill:
    ; ==========================================================
    ; Tratamento de Falha: pcb_v[running_pid].w9 = 3 e kill()
    ; ==========================================================
    MV W8, #0x2010
    ADD W8, K0, W8
    LOAD W9, W8      ; W9 = running_pid

    MV W1, #3
    MV W2, #20
    SHL W1, W1, W2                 ; W1 = 3 << 20 = 0x300000 (3MB)
    MV W2, #80
    ADD W1, W1, W2                 ; W1 = 0x300050 = 3145808 (3MB + 80)

    MUL W0, W9, W1                 ; W0 = running_pid * pcb_size
    
    MV W8, #0x2024
    ADD W1, K0, W8                  ; W1 = pcb_v base
    ADD W0, W0, W1                 ; W0 = endereço de pcb_v[running_pid]
    
    MV W1, #56                     ; Offset de W9 no PCB
    ADD W0, W0, W1                 ; W0 = endereço de pcb_v[running_pid].W9
    
    MV W1, #3
    STORE W1, W0                   ; pcb_v[running_pid].W9 = 3
    
    JUMP kill                      ; kill(running_pid) - não retorna

page_fault_clear_page:
    ; ==========================================================
    ; 3. Limpeza da Página (Zero-fill)
    ; ==========================================================
    ; aligned_efa = (EFA >> 12) << 12
    MV W1, #-12
    SHL W2, EFA, W1
    MV W1, #12
    SHL W2, W2, W1                 ; W2 = EFA alinhado (início da página)

    MV W3, #4096                   ; Tamanho da página (limit do loop)
    MV W4, #0                      ; Offset = 0
    MV W5, #0                      ; Valor zero para limpar memória
    MV W7, #4                      ; Passo do loop = 4 bytes (32 bits)

page_fault_clear_loop:
    CMP W4, W3
    BEQ page_fault_restore         ; Se offset == 4096, terminou a limpeza
    
    ADD W6, W2, W4                 ; W6 = aligned_EFA + offset
    STORE W5, W6                   ; memory[aligned_EFA + offset] = 0
    
    ADD W4, W4, W7                 ; offset += 4
    JUMP page_fault_clear_loop

page_fault_restore:
    ; ==========================================================
    ; 4. Ajuste do EPC e Restauração de Contexto
    ; ==========================================================
    ; O exception_supervisor já salvou os registradores no PCB antes de chamar essa função.
    ; Precisamos subtrair 4 do PC salvo no PCB (offset 60) para re-executar a instrução.
    
    MV W8, #0x2010
    ADD W8, K0, W8
    LOAD W9, W8     ; w9 = running_pid
    
    MV W1, #3
    MV W2, #20
    SHL W1, W1, W2                 ; W1 = 3 << 20 = 0x300000 (3MB)
    MV W2, #80
    ADD W1, W1, W2                 ; W1 = 0x300050 = 3145808 (3MB + 80)

    MUL W0, W9, W1
    
    MV W8, #0x2024
    ADD W1, K0, W8                 ; w1 = pcb_v[0]
    ADD W6, W0, W1                 ; W6 = base de pcb_v[running_pid]

    MV W1, #60                     ; Offset do PC (EPC) no PCB
    ADD W2, W6, W1                 ; W2 = endereço de pcb_v[running_pid].PC
    LOAD W3, W2                    ; W3 = EPC salvo
    
    MV W4, #4
    SUB W3, W3, W4                 ; EPC = EPC - 4
    STORE W3, W2                   ; Atualiza o PC salvo no PCB

    ; O label schedule_restore_context exige:
    ; W6 = &pcb_v[running_pid]
    ; W9 = running_pid
    ; Como ambos já estão configurados no código acima (W9 com o LDD, W6 com a base),
    ; basta pular diretamente para a rotina que ela irá restaurar os registradores 
    ; deste PCB específico para a CPU e executar o MRET.
    JUMP schedule_restore_context


; SYSCALLS

fork:
    ; K0 = kernel boundary = 0xC0000000
    MV K0, #0xC000
    MV W0, #16
    SHL K0, K0, W0

    ; Search for a free PCB slot
    ; for pid in range(0, MAX_PROCESSES)
    
    MV W4, #0                        ; W4 = pid = 0

    MV W5, #0x2000
    ADD W5, K0, W5
    LOAD W5, W5                      ; W5 = MAX_PROCESSES

    fork_loop:
        CMP W4, W5
        BEQ fork_end                 ; No free slot found

        ; W0 = pcb_size (0x00300050)
        MV W0, #0x30
        MV W1, #16
        SHL W0, W0, W1
        MV W1, #0x50
        ADD W0, W0, W1

        ; Compute &pcb_v[pid]
        MUL W1, W4, W0
        MV W8, #0x2024
        ADD W0, K0, W8               ; W0 = 0xC0002024
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

        ; parent_pid = -1 (offset 0)
        ; parent_pid is only set to running_pid when fork is for sure going to work
        ; if it returns -1 and kill the child, it must not create a zombie process
        MV W7, #-1    ; W7 = -1
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

        ; Compute parent PCB address
        ; W7 = &pcb_v[running_pid]

        MV W3, #0x30
        MV W1, #16
        SHL W3, W3, W1
        MV W1, #0x50
        ADD W3, W3, W1               ; W3 = pcb_size (0x300050)
        
        MV W1, #0x2010
        ADD W1, K0, W1
        LOAD W1, W1                  ; W1 = running_pid

        MUL W1, W1, W3
        MV W8, #0x2024
        ADD W7, K0, W8               ; W7 = 0xC0002024
        ADD W7, W7, W1               ; W7 = &pcb_v[running_pid]

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
        MV W2, K0             ; W2 = 0xC0000000 (kernel mapping offset)

        ; Child PT = &pcb_v[pid] + 80
        MV W1, #80
        ADD W1, W0, W1                   ; W1 = virtual child pt addr
        SUB W1, W1, W2                   ; W1 = physical child pt addr
        
        MV W8, #0x2020
        ADD W8, K0, W8                   ; W8 = 0xC0002020 (&scratch_space_1)
        STORE W1, W8                     ; scratch_space_1 = child PT (physical)
        MV UPTR, W1                      ; UPTR = child PT (physical)

        ; Parent PT = &pcb_v[running_pid] + 80
        MV W1, #80
        ADD W1, W7, W1                   ; W1 = virtual parent pt addr
        SUB W1, W1, W2                   ; W1 = physical parent pt addr
        
        MV W8, #0x201C
        ADD W8, K0, W8                   ; W8 = 0xC000201C (&scratch_space_0)
        STORE W1, W8                     ; scratch_space_0 = parent PT (physical)

        ; Save child PCB (W0), pid (W4), parent PCB (W7) to kernel stack
        MV W8, #0x2018
        ADD W8, K0, W8                   ; W8 = 0xC0002018 (&kernel_stack_pointer)
        LOAD SP, W8                      ; SP = kernel_stack_pointer

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

        ; W6 = physical parent PT base addr
        MV W8, #0x201C
        ADD W8, K0, W8               ; W8 = 0xC000201C (&scratch_space_0)
        LOAD W6, W8                  ; W6 = physical parent PT base addr
        ADD W6, W6, W2               ; W6 = virtual parent PT base addr

        MV W2, #0                     ; W2 = pte_index = 0

    fork_pte_loop:
        CMP W2, W3
        BEQ fork_pte_done

        ; Read PTE from parent page table
        MV W1, #4
        MUL W8, W2, W1
        ADD W8, W6, W8               ; W8 = &parent_pt[pte_index]
        LOAD W5, W8                  ; W5 = PTE value

        ; Check Valid bit (bit 20 of PTE)
        MV W1, #0x100000
        CMP W5, W1
        BLT fork_pte_next            ; PTE not valid, skip

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
        MV W8, #0x2020
        ADD W8, K0, W8
        LOAD UPTR, W8                    ; UPTR = child PT (physical)

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

        ; pcb_v[parent_pid].W9 = -1
        MV W1, #56
        ADD W8, W7, W1                ; W8 = parent.W9
        MV W1, #-1
        STORE W1, W8                  ; parent.W9 = -1 failed fork

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
        MV W7, #0x201C
        ADD W7, K0, W7
        LOAD UPTR, W7                 ; UPTR = parent PT (physical)
        LOAD W5, W8                   ; W5 = parent_page[offset]

        ; Switch UPTR to child -> write
        MV W7, #0x2020
        ADD W7, K0, W7
        LOAD UPTR, W7                 ; UPTR = child PT (physical)
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

        ; Set fork return values
        ; pcb_v[running_pid].w9 = pid
        MV W1, #56
        ADD W8, W7, W1
        STORE W4, W8                  ; parent gets child pid

        ; pcb_v[pid].w9 = -2
        ADD W8, W0, W1                ; W1 still 56
        MV W1, #-2
        STORE W1, W8                  ; child gets -2

        ; parent_pid = running_pid (offset 0)
        MV W8, #0x2010
        ADD W8, K0, W8
        LOAD W5, W8                  ; W5 = running_pid
        STORE W5, W0

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
            MV W3, #0x30
            MV W0, #16
            SHL W3, W3, W0
            MV W0, #0x50
            ADD W3, W3, W0            ; W3 = pcb_size (0x300050)
            MUL W1, W1, W3
            MV W8, #0x2024
            ADD W3, K0, W8            ; W3 = 0xC0002024 (&pcb_v[0] virtual)
            ADD W1, W1, W3            ; W1 = &pcb_v[old_child] (virtual address)

            MV W3, #8
            ADD W1, W1, W3
            STORE W4, W1              ; pcb_v[old_child].prev_sibling = pid

        skip_prev_sibling:
            ; pcb_v[running_pid].child = pid
            MV W1, #4
            ADD W8, W7, W1
            STORE W4, W8

        JUMP schedule

    next_iteration:
        MV W1, #1
        ADD W4, W4, W1               ; pid++
        JUMP fork_loop

    fork_end:
        ; No free slot: pcb_v[running_pid].w9 = -1
        MV W0, #0x30
        MV W1, #16
        SHL W0, W0, W1
        MV W1, #0x50
        ADD W0, W0, W1               ; W0 = pcb_size (0x300050)
        LDD W1, #0x2010    ; W1 = running_pid
        MUL W1, W1, W0
        MV W0, #0x2024
        ADD W0, W0, W1               ; &pcb_v[running_pid]

        MV W1, #56
        ADD W0, W0, W1               ; &pcb_v[running_pid].w9
        MV W1, #-1
        STORE W1, W0                  ; pcb_v[running_pid].w9 = -1
        JUMP schedule

wait:
    MV W8, #0x2024
    ADD W0, K0, W8                ; W0 points to pcb_v first byte
    MV W8 #52 
    ADD W0, W0, W8 ; w0 points to pcb_v[0].w8
    MV W8 #3145808 ; bytes size of each pcb

    MV W2, #0x2010
    ADD W2, K0, W2
    LOAD W1, W2  ; W1 = running_pid
    MUL W1 W1 W8 ; W1 = RUNNING_PID * pcb_size bytes
    ADD W0 W0 W1 ; w0 points to pcb_v[RUNNING_PID].w8
    
    LOAD W9 W0 ; w9 = status_addr

status_addr_check:

    MV W6 #0xC0000000 
    CMP W9 W6
    BGT fault_int
    BEQ fault_int ; fault_int if status_addr in a kernel address

    MV W6 #0b1000000000000 ; takes the 20 most significant bits of address
    DIV W2 W9 W6 ; w2 = page_number

    MV W6 #28 ; 
    ADD W0 W0 W6 ; w0 points to pcb_v[RUNNING_PID].page_table[0]

    MV W6 #4
    MUL W6 W2 W6 
    ADD W3 W0 W6 ; w3 points to pcb_v[RUNNING_PID].page_table[page_number]

    LOAD W4 W3 ; w4 = pcb_v[RUNNING_PID].page_table[page_number]
    
    MV W6 #0x100000 ; 20th bit, valid
    CMP W4 W6
    BLT fault_int ; page_number(status_addr) is not mapped, page_fault


    MV W8 #76
    SUB W2 W0 W8 ; w2 points to pcb_v[RUNNING_PID].child 
    LOAD W6 W2 ; w6 = pcb_v[RUNNING_PID].child 

    MV W8 #-1 
    CMP W6 W8 ; if pcb_v[RUNNING_PID].child  == -1:
    BEQ #2

    JUMP #6

        MV W8 #52
        ADD W6 W2 W8 ; w6 points to pcb_v[RUNNING_PID].w9
        MV W8 #-1
        STORE W8 W6
        JUMP schedule

    LOAD W3 W2 ; w3 = curr_child = pcb_v[running_pid].childPID
    MV W8, #0x2024
    ADD W0,K0,W8                ; W0 = pcb_v base

; w0 points to pcb_v[0] first byte
; w1  = RUNNING_PID * pcb_size bytes
; w2 points to pcb_v[RUNNING_PID].child 
; w3 = pcb_v[running_pid].childPID
; w4 = pcb_v[RUNNING_PID].page_table[page_number]

    waitLoop:
        ;calculate pcb_v[curr_child]
        ; w3 = curr_child

        MV W8 #3145808
        MUL W1 W3 W8

        ADD W1 W0 W1 ;W1 points to pcb_v[curr_child] first byte

        MV W8 #72
        ADD W1 W1 W8 ; w1 points to pcb_v[curr_child].flags
        LDB W6 W1 ; w6 = pcb_v[curr_child].flags

        MV W8 #2 ;0b00000000000000000000000000000010
        OR W8 W8 W6
        CMP W6 W8
        BEQ curr_child_is_zombie 

            ;if !pcb_v[curr_child].is_zombie:
            ; curr_child = pcb_v[curr_child].next_sibling

            MV W8 #60
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
            LDB W6 W1 ; w6 = pcb_v[curr_child].flags
            AND W8 W8 W6 
            STRB W8 W1 ; sets pcb_v[curr_child].is_mapped = 0


            
            ; w0 points to pcb_v[0] first byte
            ; w1 points to pcb_v[curr_child].flags
            ; w2 points to pcb_v[running_pid].child
            ; w3 = curr_child
            

            MV W8 #52
            ADD W4 W2 W8 ; w4 points to pcb_v[running_pid].w9
            STORE W3 W4 ; pcb_v[running_pid].w9 = curr_child

            MV W8 #16 
            SUB W1 W1 W8 ; w1 points to pcb_v[curr_child].w9
            LOAD W6 W1 ; w6 = pcb_v[curr_child].w9

            STORE W6 W9 ; running_pid_virtual_memory[status_addr] = pcb_v[curr_child].w9

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

            MV W6 #3145808 ; pcb_size
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

            MV W6 #3145808 ; pcb_size
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

        MV W8 #56
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
    MV W6, #0xC000             ; W6 = upper part of kernel boundary
    MV W7, #16                 ; W7 = shift amount
    SHL W6, W6, W7            ; W6 = 0xC0000000 (kernel boundary)

    MV W5, #0x2010             ; #0x2010 running_pid
    ADD W5, W6, W5            ; W5 = running_pid virtual address
    LOAD W9, [W5]              ; W9 = running_pid

    MV W1, #0x30               ; W1 = upper part of pcb_size
    MV W7, #16                 ; W7 = shift amount
    SHL W1, W1, W7            ; W1 = 0x00300000 (page table size)

    MV W7, #0x50               ; W7 = 80 bytes of PCB fields
    ADD W1, W1, W7             ; W1 = 0x00300050 (pcb_size)

    MUL W0, W9, W1             ; W0 = running_pid * pcb_size

    MV W1, #0x2024             ; #0x2024 pcb_vector
    ADD W1, W6, W1             ; W1 = pcb_v virtual initial address
    ADD W0, W0, W1             ; W0 = pcb_v[running_pid] initial address

    MV W1, #52                 ; W1 = W8 offset inside PCB
    ADD W0, W0, W1             ; W0 = pcb_v[running_pid].W8 address
    LOAD W2, [W0]              ; W2 = pcb_v[running_pid].W8 (status_code)

    MV W1, #4                  ; W1 = distance from W8 to W9
    ADD W0, W0, W1             ; W0 = pcb_v[running_pid].W9 address
    STORE W2, [W0]             ; pcb_v[running_pid].W9 = status_code

    JUMP kill                  ; kill(running_pid), with PID still in W9


getPID:
    MV W2, #0x2010
    ADD W2, K0, W2
    LOAD W0, W2    ; W0 = running_pid
    
    MV W8, #0x2024
    ADD W1, K0, W8                ; W1 = pcb_v base

    MV W8 #3145808 ; pcb_size 
    MUL W2 W0 W8 ; RUNNING_PID * pcb_size get the offset of bytes to acess pcb[RUNNING_PID]
    
    ADD W1 W1 W2 ; W1 = pcb[RUNNING_PID] addr

    MV W8 #56
    ADD W1 W1 W8 ; W1 = pcb[RUNNING_PID].w9 addr

    STORE W0 W1 ; stores RUNNING_PID on pcb[RUNNING_PID].w9

    JUMP schedule


rele:
    JUMP schedule


; HELPERS


kill:
  MV W0, #pcb_v           ; W0 = pcb_v initial address
  MV W1, #0x300050        ; W1 = pcb_size
  MUL W2, W1, W9
  ADD W2, W0, W2          ; W2 = pcb_v[pid].parent_pid address

  LOAD W3, W2             ; W3 = pcb_v[pid].parent_pid

  MV W4, #-1
  CMP W3, W4
  BEQ kill_no_parent      ; if parent != -1

    MUL W3, W3, W1
    ADD W3, W3, W0        ; W3 = pcb_v[parent] initial address

    MV W4, #72
    ADD W3, W3, W4        ; W3 = pcb_v[parent].flags address

    LDB W4, W3           ; W4 = pcb_v[parent].flags
    MV W5, #4
    AND W4, W4, W5        ; W4 = pcb_v[parent].is_waiting

    CMP W4, W5
    BEQ #2                ; if pcb_v[parent].is_waiting
    JUMP kill_not_waiting

      MV W4, #9           ; mapped = 1, zombie = 0, waiting = 0, state = ready
      STRB W4, W3

      MV W5, #16
      SUB W3, W3, W5      ; W3 = pcb_v[parent].W9 address
      STORE W9, W3        ; pcb_v[parent].W9 = pid

      MV W5, #40
      SUB W3, W3, W5      ; W3 = pcb_v[parent].status_addr address
      LOAD W4, W3         ; W4 = pcb_v[parent].status_addr

      MV W5, #64
      ADD UPTR, W3, W5      ; UPTR = pcb_v[parent].page_table virtual address
      MV W5, #0xC0000000    ; The 3GB kernel offset
      SUB UPTR, UPTR, W5    ; UPTR = pcb_v[parent].page_table physical address

      MV W5, #56
      ADD W2, W2, W5      ; W2 = pcb_v[pid].W9 address
      LOAD W5, W2         ; W5 = pcb_v[pid].W9

      STORE W5, W4        ; memory[pcb_v[parent].status_addr] = pcb_v[pid].w9

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

      
      MV W3, #60
      ADD W2, W2, W3      ; W2 = pcb_v[pid].flags address
      LDB W3, W2         ; W3 = pcb_v[pid].flags

      MV W4, #-2          ; everything 1 except the LSB
      AND W3, W3, W4      ; is_mapped = 0
      STRB W3, W2        ; pcb_v[pid].is_mapped = 0

      JUMP kill_orphanize ; execution after the elses

    kill_not_waiting:
      ; here W2 = pcb_v[pid].parent_pid address
      MV W3, #72
      ADD W2, W2, W3      ; W2 = pcb_v[pid].flags address
      LDB W3, W2         ; W3 = pcb_v[pid].flags

      MV W4, #2
      OR W3, W3, W4       ; is_zombie = 1
      STRB W3, W2        ; pcb_v[pid].is_zombie = 1
      JUMP kill_orphanize ; execution after the elses


  kill_no_parent:
    ; here W2 = pcb_v[pid].parent_pid address
    MV W3, #72
    ADD W2, W2, W3      ; W2 = pcb_v[pid].flags address
    LDB W3, W2         ; W3 = pcb_v[pid].flags

    MV W4, #-2          ; everything 1 except the LSB
    AND W3, W3, W4      ; is_mapped = 0
    STRB W3, W2        ; pcb_v[pid].is_mapped = 0

  
  kill_orphanize:
    ; here W2 = pcb_v[pid].flags address

    MV W3, #68
    SUB W2, W2, W3      ; W2 = pcb_v[pid].child address
    LOAD W3, W2          ; W3 = pcb_v[pid].child (curr_child)
    MV W7, #-1

    orphanize_loop:
      CMP W3, W7
      BEQ orphanize_end

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


  orphanize_end:
    MV W0, #bitmap      ; W0 = bitmap address

    MV W1, #72
    ADD W1, W2, W1      ; W1 = pcb_v[pid].pages_used address
    LOAD W2, W1         ; W2 = pcb_v[pid].pages_used

    MV W3, #4
    ADD W3, W1, W3      ; W3 = pcb_v[pid].page_table first address (pointer i)
    MV W4, #0x2FFFFC    ; 3MB - 1B
    ADD W4, W4, W3      ; W4 = pcb_v[pid].page_table last address (pointer j)

    MV W5, #0x100000    ; W5 = mask of valid (bit 20)
    MV W6, #0xFFFFF     ; W6 = mask of page frame id (bits 0-19)

    kill_free_memory_loop:
      MV W7, #0
      CMP W2, W7
      BEQ kill_end      ; while pages_used != 0

      LOAD W7, W3       ; W7 = pte_i
      AND W8, W7, W5    ; W8 = pte_i.valid
      CMP W8, W5
      BLT kill_skip_i   ; if pte_i is valid, enter block

        MV W8, #0
        STORE W8, W3    ; set pte_i invalid

        AND W8, W7, W6  ; W8 = pte_i.page_frame_id
        
        ; unmap the bitmap
        MV W9, #-5          ; W9 = -5
        SHL W9, W8, W9      ; W9 = word offset (W8 >> 5)
        
        MV W7, #2
        SHL W9, W9, W7      ; W9 = byte address offset (W9 * 4)

        ADD W9, W0, W9      ; W9 = address of the word in the bitmap
        
        MV W7, #31
        AND W8, W8, W7      ; W8 = bit index (0 to 31)
        
        MV W7, #1
        SHL W7, W7, W8      ; W7 = 1 << bit_index (clear mask)
        
        LOAD W8, W9         ; Load bitmap word into W8
        XOR W8, W8, W7      ; Clear the frame's bit
        STORE W8, W9        ; Store the updated word back to memory
        
        MV W7, #1
        SUB W2, W2, W7      ; W2 = W2 - 1 (pages_used--)

      kill_skip_i:

      LOAD W7, W4       ; W7 = pte_j
      AND W8, W7, W5    ; W8 = pte_j.valid
      CMP W8, W5
      BLT kill_skip_j   ; if pte_j is valid, enter block

        MV W8, #0
        STORE W8, W4    ; set pte_j invalid

        AND W8, W7, W6  ; W8 = pte_j.page_frame_id
        
        ; unmap the bitmap
        MV W9, #-5          ; W9 = -5
        SHL W9, W8, W9      ; W9 = word offset (W8 >> 5)
        
        MV W7, #2
        SHL W9, W9, W7  

        ADD W9, W0, W9      ; W9 = address of the word in the bitmap
        
        MV W7, #31
        AND W8, W8, W7      ; W8 = bit index (0 to 31)
        
        MV W7, #1
        SHL W7, W7, W8      ; W7 = 1 << bit_index (clear mask)
        
        LOAD W8, W9         ; Load bitmap word into W8
        XOR W8, W8, W7      ; Clear the frame's bit
        STORE W8, W9        ; Store the updated word back to memory
        
        MV W7, #1
        SUB W2, W2, W7      ; W2 = W2 - 1 (pages_used--)

      kill_skip_j:

      MV W7, #4
      ADD W3, W3, W7        ; pte_i += 1
      SUB W4, W4, W7        ; pte_j -= 1

      JUMP kill_free_memory_loop

  kill_end:
    JUMP schedule


schedule:

    ; W8 = kernel boundary = 0xC0000000

    MV W8, #0xC000
    MV W0, #16
    SHL W8, W8, W0

    ; clock_interrupt_count = 0

    MV W0, #0x2014 ; #0x2014 clock_interrupt_count
    ADD W0, W8, W0

    MV W5, #0
    STORE W5, W0

    ; W7 = MAX_PROCESSES

    MV W0, #0x2000 ; #0x2000 MAX_PROCESSES
    ADD W0, W8, W0
    LOAD W7, W0

    ; W6 = PCB_SIZE

    MV W6, #0x30
    MV W0, #16
    SHL W6, W6, W0     ; W6 = 0x300000

    MV W0, #0x50
    ADD W6, W6, W0     ; W6 = 0x300050

    ; if running_pid >= MAX_PROCESSES:
    ;     running_pid = 0

    MV W0, #0x2010 ; #0x2010 running_pid
    ADD W0, W8, W0
    LOAD W9, W0

    CMP W9, W7
    BEQ schedule_reset_running_pid
    BGT schedule_reset_running_pid

    JUMP schedule_check_current_process

    schedule_reset_running_pid:
    MV W9, #0


    ; if pcb_v[running_pid].scheduler_state == running:
    ;     pcb_v[running_pid].scheduler_state = ready

    schedule_check_current_process:
    MUL W1, W9, W6

    MV W0, #0x2024 ; #0x2024 pcb_vector
    ADD W0, W8, W0
    ADD W1, W0, W1

    MV W0, #72
    ADD W0, W1, W0
    LDB W2, W0

    MV W5, #0x18
    AND W5, W2, W5

    MV W0, #0
    CMP W5, W0
    BEQ schedule_set_current_ready

    JUMP schedule_calculate_limit_pid


    schedule_set_current_ready:

    MV W5, #-25
    AND W2, W2, W5

    MV W5, #0x08
    OR W2, W2, W5

    MV W0, #72
    ADD W0, W1, W0
    STRB W2, W0

    schedule_calculate_limit_pid:
    ; limit_pid = running_pid + 1
    ; limit_pid shitty name? @cartaxo

    MV W0, #1
    ADD W3, W9, W0


    ; if limit_pid == MAX_PROCESSES:
    ;     limit_pid = 0

    CMP W3, W7
    BEQ schedule_limit_pid_zero

    JUMP schedule_limit_pid_ready


    schedule_limit_pid_zero:
    MV W3, #0


    ; curr_pid = limit_pid

    schedule_limit_pid_ready:

    MV W4, #0
    ADD W4, W4, W3


    schedule_search_loop:

    ; if pcb_v[curr_pid].is_mapped == 1:

    MUL W1, W4, W6

    MV W0, #0x2024 ; #0x2024 pcb_vector
    ADD W0, W8, W0
    ADD W1, W0, W1

    MV W0, #72
    ADD W0, W1, W0
    LDB W2, W0

    MV W5, #0x01
    AND W5, W2, W5

    MV W0, #0x01
    CMP W5, W0
    BEQ schedule_check_zombie

    JUMP schedule_next_pid


    schedule_check_zombie:

    ; if pcb_v[curr_pid].is_zombie == 0:

    MV W5, #0x02
    AND W5, W2, W5

    MV W0, #0
    CMP W5, W0
    BEQ schedule_check_ready

    JUMP schedule_next_pid


    schedule_check_ready:

    ; if pcb_v[curr_pid].scheduler_state == ready:

    MV W5, #0x18
    AND W5, W2, W5

    MV W0, #0x08
    CMP W5, W0
    BEQ schedule_process_found

    JUMP schedule_next_pid


    schedule_process_found:

    ; pcb_v[curr_pid].scheduler_state = running

    MV W5, #-25
    AND W2, W2, W5

    MV W0, #72
    ADD W0, W1, W0
    STRB W2, W0


    ; running_pid = curr_pid

    MV W9, #0
    ADD W9, W9, W4

    MV W0, #0x2010 ; #0x2010 running_pid
    ADD W0, W8, W0
    STORE W9, [W0]


    ; UPTR = pcb_v[running_pid].page_table address


    MV W6, #0
    ADD W6, W6, W1

    SUB W0, W6, W8

    MV W5, #80
    ADD W0, W0, W5

    MV W5, #0
    ADD UPTR, W0, W5


    ; load every register of pcb_v[running_pid] into the CPU

    JUMP schedule_restore_context

    schedule_restore_context:


    MV W0, #24
    ADD W0, W1, W0
    LOAD W5, W0              

    MV W0, #0x201C ; #0x201C scratch_space_0
    ADD W0, W8, W0
    STORE W5, W0




    MV W0, #52
    ADD W0, W1, W0
    LOAD W5, W0              

    MV W0, #0x2020 ; #0x2020 scratch_space_1
    ADD W0, W8, W0
    STORE W5, W0


    ; restore PC -> EPC

    MV W0, #60
    ADD W0, W1, W0
    LOAD W5, W0

    MV EPC, W5


    ; restore SP

    MV W0, #64
    ADD W0, W1, W0
    LOAD W5, W0

    MV SP, W5



    MV W0, #68
    ADD W0, W1, W0
    LOAD W5, W0

    MV ESR, W5


    ; restore W2

    MV W0, #28
    ADD W0, W1, W0
    LOAD W2, W0


    ; restore W3

    MV W0, #32
    ADD W0, W1, W0
    LOAD W3, W0


    ; restore W4

    MV W0, #36
    ADD W0, W1, W0
    LOAD W4, W0


    ; restore W5

    MV W0, #40
    ADD W0, W1, W0
    LOAD W5, W0


    ; restore W6

    MV W0, #44
    ADD W0, W1, W0
    LOAD W6, W0


    ; restore W7

    MV W0, #48
    ADD W0, W1, W0
    LOAD W7, W0

    ; restore W9

    MV W0, #56
    ADD W0, W1, W0
    LOAD W9, W0


    ; restore W0

    MV W0, #20
    ADD W0, W1, W0
    LOAD W0, W0



    MV W1, #0x201C ; #0x201C scratch_space_0
    ADD W1, W8, W1
    LOAD W1, W1



    ADD W8, W8, #0x2020 ; #0x2020 scratch_space_1
    LOAD W8, W8



    MRET


    schedule_next_pid:

    ; curr_pid++

    MV W0, #1
    ADD W4, W4, W0


    ; if curr_pid == MAX_PROCESSES:
    ;     curr_pid = 0

    CMP W4, W7
    BEQ schedule_wrap_pid

    JUMP schedule_check_search_end


    schedule_wrap_pid:

    MV W4, #0


    schedule_check_search_end:

    ; while curr_pid != limit_pid

    CMP W4, W3
    BEQ schedule_no_ready_process

    JUMP schedule_search_loop


    schedule_no_ready_process:

    ; running_pid = -1

    MV W9, #-1

    MV W0, #0x2010 ; #0x2010 running_pid
    ADD W0, W8, W0
    STORE W9, W0


    MV W0, loop
    ADD W0, W8, W0

    MV W5, #0
    ADD EPC, W0, W5

    MV ESR, #48

    MRET

loop:
    JUMP loop


map_page:
  MV W0, #bitmap       ; W0 = bitmap start address
  
  LDD W1, #memory_size ; W1 = memory_size
  MV W2, #-12
  SHL W1, W1, W2       ; W1 = frame_number (memory/4KB)    
  
  MV W7, #0            ; W7 = current frame_id

  map_find_word_loop:
    CMP W7, W1
    BEQ map_frame_not_found
    BGT map_frame_not_found     ; if current frame_id >= frame_number

    LOAD W2, W0                 ; W2 = current word of the bitmap
    MV W3, #-1                  ; mask of a fully occupied bitmap word
    CMP W2, W3
    BLT map_found_free_word     ; if there is a bit = 0 in the word, jump

      MV W2, #4                 
      ADD W0, W0, W2
      MV W2, #32
      ADD W7, W7, W2            ;update current bitmap word address and current frame id
      JUMP map_find_word_loop   ; go to next word if there is no bit = 0

    
    map_found_free_word:

      MV W3, #1            ; current bit to be checked

      map_find_bit_loop:

        MV W5, #1
        AND W4, W2, W3     ; W4 = current bit of the bitmap word
        CMP W4, W3
        BLT map_found_free_bit    ; if the bitmap is free at the set bit in W3, jump

          SHL W3, W3, W5
          ADD W7, W7, W5          ; update current bit and page frame id

          CMP W7, W1
          BEQ map_frame_not_found
          BGT map_frame_not_found     ; if current frame_id >= frame_number

          JUMP map_find_bit_loop  ; go to next bit

      map_found_free_bit:

        OR W2, W2, W3
        STORE W2, W0       ; set the found free bit to 1 and save it to the bitmap

        MV W1, #0xC0000    ; id of the first kernel page
        CMP W9, W1
        BGT map_kernel_pte
        BEQ map_kernel_pte       ; jump if it is a kernel page

          ; if it is an user page:
          MV W1, #2
          SHL W3, W9, W1     ; W3 = page table offset of the page to be mapped
          
          ADD W4, UPTR, W3   ; W4 = physical pte address to be set 
          MV W3, #0xC0000000 ; The 3GB kernel offset
          ADD W4, W4, W3     ; W4 = virtual pte address to be set 

          MV W5, #4
          SUB W5, UPTR, W5   ; W5 = process.pages_used physical address
          ADD W5, W5, W3     ; W5 = process.pages_used virtual address

          LOAD W6, W5        ; W6 = process.pages_used
          MV W1, #1
          ADD W6, W6, W1
          STORE W6, W5       ; process.pages_used += 1
          JUMP map_pte
        
        map_kernel_pte:
          SUB W3, W9, W1     ; page_id -= 0xC0000 (normalize to the start of KPTR)
          MV W1, #2
          SHL W3, W3, W1     ; W3 = page table offset of the page to be mapped
          
          ADD W4, KPTR, W3   ; W4 = physical pte address to be set 
          MV W3, #0xC0000000 ; The 3GB kernel offset
          ADD W4, W4, W3     ; W4 = virtual pte address to be set

        map_pte:

          MV W1, #0x100000   ; to set the valid bit = 1 of the pte
          ADD W7, W7, W1     ; W7 = pte (valid = 1 and page frame id)
          STORE W7, W4       ; mapping the page to the frame

          MV W9, #0
          JUMPR W8           ; return

  map_frame_not_found:

    MV W1, #0xC0000    ; id of the first kernel page
    CMP W9, W1
    BLT map_frame_not_found_user ; jump if it is an user page

      ; kernel page and not found, panic
      JUMP #0
    
    map_frame_not_found_user:
      MV W9, #1
      JUMPR W8   ; return