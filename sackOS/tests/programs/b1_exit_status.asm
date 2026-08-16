; B1 - exit_status: exit devolve o status em W9
; expected: w9=42 (exit code)
_start:
    MV W8 #42          ; status = 42
    SYSCALL #2         ; exit -> W9 = 42