getPID:
    LDD W0 #RUNNING_PID_CNST
    LDD W1 #PCBADRESS_CNST

    MV W8 #84
    MUL W2 W0 W8 ; RUNNING_PID * 84 get the offset of bytes to acess pcb[RUNNING_PID]
    
    ADD W1 W1 W2 ; stores on W1 the first address of pcb[RUNNING_PID]

    MV W8 #56
    ADD W1 W1 W8 ; W1 = pcb[RUNNING_PID].w9 address bytes

    STORE W0 W1 ; stores RUNNING_PID on pcb[RUNNING_PID].w9

    JUMP schedule

setup:
    LDD W1 #KERNEL_MAX_MEMORY_CNST
    LDD W2 #BUFFER_CNST
    ADD W3 W1 W2 ; W3 -> KERNEL_MAX_MEMORY + BUFFER

    SUB W0 LIMIT W3 ; W0 -> USER_MEMORY = limit - (KERNEL_MAX_MEMORY + BUFFER)
    LDD W2 #PARTITION_SIZE_CNST

    DIV W3 W0 W2 ; W3 -> NUM_PARTITIONS = user_memory / PARTITION_SIZE

    LDD W4 #PCBADRESS_CNST

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
    MV SP #SP_ADDRESS_CNST
    MV ESA #EXCEPTION_SUPERVISOR_ADDRESS_CNST
    MV EPC #LOOP_ADDRESS_CNST
    MV ESR #16 ; enable interruptions
    
    MRET ; go to infinite loop, waiting for program inputs

exception_supervisor:
    STRD W0 #SCRATCH_SPACE_0_CNST
    STRD W1 #SCRATCH_SPACE_1_CNST

    MV W0 #0
    CMP ECR W0 ; if ECR = 0 (clock interruption)
    BEQ _clock_int

    MV W0 #-1
    LDD W1 #RUNNING_PID_CNST

    CMP W0 W1
    BEQ jumpToHandler

    LDD W0 #84 ; pcb_size
    MUL W1 W1 W0 ; W1 = RUNNING_PID * pcb_size

    LDD W0 #PCBADRESS_CNST
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

    LDD W2 #SCRATCH_SPACE_0_CNST

    STORE W2 W0

    MV W8 #4
    ADD W0 W1 W8 ; w0 points to pcb[RUNNING_PID].w1

    LDD W2 #SCRATCH_SPACE_1_CNST
    STORE W2 W0

jumpToHandler:
    MV W0 #1
    MV W1 #1 

    CMP ECR W0
    BEQ _input_int

    ADD W0 W0 W1

    CMP ECR W0
    BEQ _kill_int

    ADD W0 W0 W1 
    

    CMP ECR W0
    BEQ _syscall_int

    ADD W0 W0 W1 
    

    CMP ECR W0
    BEQ _falt_int    

wait:


    LDD W0 #PCBADRESS_CNST ; w0 points to pcb_v[0] first byte 
    MV W8 #52 
    ADD W0 W8 ; w0 points to pcb_v[0].w8
    MV W8 #84 ; bytes size of each pcb

    LDD W1 #RUNNING_PID_CNST
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
    BGT _fault_int 


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
    LDD W0 #PCBADRESS_CNST ;  w0 points to pcb_v[0] first byte
    ; 
    ; starts loop
    JUMP waitLoop

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
