; C2 - fork_child_exit: pai espera filho e devolve o status do filho
; expected: w7=6 (status do filho), w9=6 (exit do pai)
_start:
    MV W3 #-2          ; marcador de filho
    MV W0 #0           ; W0 = zero
    MV W7 #4
    SUB W8 SP W7       ; W8 = status_addr
    STORE W0 W8        ; touch -> mapeia a página de status_addr
    SYSCALL #0         ; fork
    CMP W9 W3
    BEQ child          ; se -2, e o filho
parent:
    SYSCALL #1         ; wait -> W9 = pid; mem[status_addr] = status
    LOAD W7 W8         ; W7 = status do filho
    ADD W8 W7 W0       ; status = 6
    SYSCALL #2         ; exit
child:
    MV W8 #6           ; status = 6
    SYSCALL #2         ; exit