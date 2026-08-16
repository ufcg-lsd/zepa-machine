; D2 - wait_child_running: pai espera filho que ainda esta em execucao (bloqueia)
; expected: w7=7 (status do filho), w9=7 (exit do pai)
_start:
    MV W3 #-2          ; marcador de filho
    MV W0 #0           ; W0 = zero
    MV W1 #1           ; incremento
    MV W4 #100         ; limite do loop do filho
    MV W7 #4
    SUB W8 SP W7       ; W8 = status_addr
    STORE W0 W8        ; touch -> mapeia a página de status_addr
    SYSCALL #0         ; fork
    CMP W9 W3
    BEQ child
parent:
    SYSCALL #1         ; wait imediato -> pai bloqueia ate filho terminar
    LOAD W7 W8         ; W7 = status do filho
    ADD W8 W7 W0       ; status = 7
    SYSCALL #2         ; exit
child:
    MV W2 #0
cloop:
    ADD W2 W2 W1       ; loop do filho
    CMP W2 W4
    BLT cloop
    MV W8 #7           ; status = 7
    SYSCALL #2         ; exit