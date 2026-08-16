; B3 - rele: syscall 4 (yield) e retomada da execucao
; expected: w5=7, w9=7 (exit code)
_start:
    MV W0 #0           ; W0 = zero
    MV W5 #7           ; W5 = 7
    SYSCALL #4         ; rele (cede a CPU)
    ADD W8 W5 W0       ; status = 7
    SYSCALL #2         ; exit