; D1 - wait_child_zombie: filho termina (vira zombie) antes do pai dar wait
; expected: w7=7 (status do filho), w9=7 (exit do pai)
_start:
    MV W3 #-2          ; marcador de filho
    MV W0 #0           ; W0 = zero
    MV W1 #1           ; incremento
    MV W4 #30000       ; limite do loop longo do pai
    MV W7 #4
    SUB W8 SP W7       ; W8 = status_addr
    STORE W0 W8        ; touch -> mapeia a página de status_addr
    SYSCALL #0         ; fork
    CMP W9 W3
    BEQ child
parent:
    MV W2 #0
ploop:
    ADD W2 W2 W1       ; loop longo para o filho terminar primeiro
    CMP W2 W4
    BLT ploop
    SYSCALL #1         ; wait -> filho ja zombie, retorna imediatamente
    LOAD W7 W8         ; W7 = status do filho
    ADD W8 W7 W0       ; status = 7
    SYSCALL #2         ; exit
child:
    MV W4 #5           ; loop curto
    MV W2 #0
cloop:
    ADD W2 W2 W1
    CMP W2 W4
    BLT cloop
    MV W8 #7           ; status = 7
    SYSCALL #2         ; exit