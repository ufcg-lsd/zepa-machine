; G1 - two_procs: fork mantem dois processos ativos (round-robin)
; expected: w9=30 (exit do pai apos loop grande; filho faz loop curto)
_start:
    MV W3 #-2          ; marcador de filho
    MV W0 #0           ; W0 = zero
    MV W1 #1           ; incremento
    MV W4 #30000       ; limite do loop do pai
    SYSCALL #0         ; fork
    CMP W9 W3
    BEQ child
parent:
    MV W2 #0
ploop:
    ADD W2 W2 W1       ; loop longo do pai
    CMP W2 W4
    BLT ploop
    MV W8 #30          ; status = 30
    SYSCALL #2         ; exit
child:
    MV W4 #15          ; loop curto do filho
    MV W2 #0
cloop:
    ADD W2 W2 W1
    CMP W2 W4
    BLT cloop
    MV W8 #15          ; status = 15
    SYSCALL #2         ; exit