_start:
    MV W9, #0           ; syscall ID = fork
    SYSCALL #0          ; W9 = child PID (parent) or 0 (child)

    MV W0, #0
    CMP W9, W0
    BEQ _child

    ; Parent: W9 = child PID
    MV W2, #10
    MV W3, #20
    ADD W4, W2, W3      ; W4 = 30

    MV W9, #1           ; syscall ID = wait
    SYSCALL #1

    MV W9, #2           ; syscall ID = exit
    MV W8, #0           ; status = 0
    SYSCALL #2

_child:
    ; Child: W9 = 0
    MV W0, #7
    MV W1, #6
    MUL W2, W0, W1      ; W2 = 42

    MV W9, #2           ; syscall ID = exit
    MV W8, #42          ; status = 42
    SYSCALL #2
