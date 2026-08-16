; B2 - getpid: syscall 3 devolve o PID do processo em W9
; expected: w5=0 (pid do 1o processo), w9=5 (exit code)
_start:
    MV W0 #0           ; W0 = zero
    SYSCALL #3         ; getPID -> W9 = pid
    ADD W5 W9 W0       ; W5 = pid
    MV W8 #5           ; status = 5
    SYSCALL #2         ; exit