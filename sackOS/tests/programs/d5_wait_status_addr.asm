; D5 - wait_status_addr: wait grava o status do filho em um endereco fixo (400)
; expected: w7=9 (status do filho lido do status_addr), w9=9 (exit do pai)
_start:
    MV W3 #-2          ; marcador de filho
    MV W0 #0           ; W0 = zero
    MV W8 #400         ; status_addr fixo (fora da area de codigo)
    SYSCALL #0         ; fork
    CMP W9 W3
    BEQ child
parent:
    SYSCALL #1         ; wait -> mem[400] = status do filho
    LOAD W7 W8         ; W7 = 9
    ADD W8 W7 W0       ; status = 9
    SYSCALL #2         ; exit
child:
    MV W8 #9           ; status = 9
    SYSCALL #2         ; exit