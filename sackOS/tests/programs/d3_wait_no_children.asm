; D3 - wait_no_children: pai sem filhos recebe -1 no wait
; expected: w5=4294967295 (wait retorna -1), w9=4 (exit code)
_start:
    MV W0 #0           ; W0 = zero
    MV W7 #4
    SUB W8 SP W7       ; W8 = status_addr (valido, mas nao ha filhos)
    STORE W0 W8        ; touch -> mapeia a página de status_addr
    SYSCALL #1         ; wait -> W9 = -1
    ADD W5 W9 W0       ; W5 = -1 (0xFFFFFFFF)
    MV W8 #4           ; status = 4
    SYSCALL #2         ; exit