_getPID:
    LOAD W0 #RUNNING_PID
    LOAD W1 #PCBADRESS
    MUL W2 W0 #84 ; get the offset of bytes to acess pcb[RUNNING_PID]
    
    ADD W1 W1 W2 ; stores on W1 the first address of pcb[RUNNING_PID]
    ADD W1 W1 #56 ; W1 = pcb[RUNNING_PID].w9 address bytes

    STORE W0 W1 ; stores RUNNING_PID on pcb[RUNNING_PID].w9

    JUMP _schedule
