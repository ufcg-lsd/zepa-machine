_getPID:
    LDD W0 #RUNNING_PID_CNST
    LDD W1 #PCBADRESS_CNST
    MUL W2 W0 #84 ; get the offset of bytes to acess pcb[RUNNING_PID]
    
    ADD W1 W1 W2 ; stores on W1 the first address of pcb[RUNNING_PID]
    ADD W1 W1 #56 ; W1 = pcb[RUNNING_PID].w9 address bytes

    STORE W0 W1 ; stores RUNNING_PID on pcb[RUNNING_PID].w9

    JUMP _schedule

_setup:
    LDD W1 #KERNEL_MAX_MEMORY_CNST
    LDD W2 #BUFFER_CNST
    ADD W3 W1 W2 ; W3 -> KERNEL_MAX_MEMORY + BUFFER

    SUB W0 LIMIT W3 ; W0 -> USER_MEMORY = limit - (KERNEL_MAX_MEMORY + BUFFER)
    LDD W2 #PARTITION_SIZE_CNST

    DIV W3 W0 W2 ; W3 -> NUM_PARTITIONS = user_memory / PARTITION_SIZE

    LDD W4 #PCBADRESS_CNST
    ADD W4 W4 #72 ; gets W4 to pcb_v[0].BASE address

   ; W1 will acumulate PARTITION_SIZE * i through the loop

    MV W6 #0 ; W6 -> pid = 0
    ; for (int pid = 0, pid < NUM_PARTITIONS; pid++)

    ADD W6 W6 1
    CMP W6 W3 
    BEQ _setup_registers ; loop conditions

    STORE W1 W4 ; pcb_v[pid].BASE = KERNEL_MAX_MEMORY + pid * PARTITION_SIZE
    
    ADD W4 W4 #4 ; W4 = pcb_v[pid].LIMIT address

    ADD W7 W1 W2 ; W7 = BASE + PARTITION_SIZE

    STORE W7 W4 ; pcb_v[pid].LIMIT = BASE + PARTITION_SIZE

    JUMP #-7 

    ; setup variables to next Iteration

    ADD W4 W4 #80 ; W4 = pcb_v[pid+1].BASE address
    ADD W1 W1 W2 ; W1 += PARTITION_SIZE


_setup_registers:
    MV SP #SP_ADDRESS_CNST
    MV ESA #EXCEPTION_SUPERVISOR_ADDRESS_CNST
    MV EPC #LOOP_ADDRESS_CNST
    MV ESR #16 ; enable interruptions
    
    MRET ; go to infinite loop, waiting for program inputs

_exception_supervisor:
    STRD W0 #SCRATCH_SPACE_0_CNST
    STRD W1 #SCRATCH_SPACE_1_CNST

    MV W0 #0
    CMP ECR W0 ; if ECR = 0 (clock interruption)
    BEQ _clock_int

    MV W0 #-1
    LDD W1 #RUNNING_PID_CNST

    CMP W0 W1
    BEQ _jumpToHandler

    LDD W0 #PCBADRESS_CNST
    LDD W2 #PARTITION_SIZE_CNST

    MUL W1 W1 W2 ; W1 = RUNNING_PID * PARTITION_SIZE
    ADD W0 W1 ; W0 = pcb[RUNNING_PID] address

    MV W8 #28
    ADD W0 W0 W8 ;  W0 = pcb[RUNNING_PID].w2 address

    ; save registers
    MV W1 #4 ; uses W1 to jump to next register address

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
    ADD W0 W1 W ; w0 points to pcb[RUNNING_PID].w1

    LDD W2 #SCRATCH_SPACE_1_CNST
    STORE W2 W0

    JUMP _jumpToHandler

_jumpToHandler:
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

    ADD W0 W0 W1 
    





