; C1 - fork_basic: pai e filho rodam, cada um com exit proprio
; expected: w5=1 (marcador do pai), w7=1 (pid do filho), w9=10 (exit do pai)
_start:
    MV W3 #-2          ; marcador de filho
    MV W0 #0           ; W0 = zero
    MV W7 #4
    SUB W8 SP W7       ; W8 = status_addr
    SYSCALL #0         ; fork
    CMP W9 W3
    BEQ child          ; se -2, e o filho
parent:
    MV W5 #1           ; marcador do pai
    ADD W7 W9 W0       ; W7 = pid do filho
    MV W8 #10          ; status = 10
    SYSCALL #2         ; exit
child:
    MV W5 #2           ; marcador do filho
    MV W8 #5           ; status = 5
    SYSCALL #2         ; exit