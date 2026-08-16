; D4 - wait_two_children: pai espera dois filhos e soma os status (1+2=3)
; expected: w6=3 (soma dos status), w9=3 (exit do pai)
_start:
    MV W3 #-2          ; marcador de filho
    MV W0 #0           ; W0 = zero (W6 ja comeca em 0 pela kernel)
    MV W7 #4
    SUB W8 SP W7       ; W8 = status_addr
    STORE W0 W8        ; touch -> mapeia a página de status_addr
    SYSCALL #0         ; fork #1
    CMP W9 W3
    BEQ child_a
    SYSCALL #0         ; fork #2 (somente no pai)
    CMP W9 W3
    BEQ child_b
parent:
    SYSCALL #1         ; wait #1
    LOAD W4 W8         ; W4 = status do 1o filho reaped
    ADD W6 W6 W4       ; W6 += status
    SYSCALL #1         ; wait #2
    LOAD W5 W8         ; W5 = status do 2o filho reaped
    ADD W6 W6 W5       ; W6 = 1 + 2 = 3
    ADD W8 W6 W0       ; status = 3
    SYSCALL #2         ; exit
child_a:
    MV W8 #1           ; status = 1
    SYSCALL #2         ; exit
child_b:
    MV W8 #2           ; status = 2
    SYSCALL #2         ; exit