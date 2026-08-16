; E1 - orphan_child: pai sai antes do filho, filho vira orfao
; expected: w9=1 (exit do pai)
_start:
    MV W3 #-2          ; marcador de filho
    MV W0 #0           ; W0 = zero
    SYSCALL #0         ; fork
    CMP W9 W3
    BEQ child
parent:
    MV W8 #1           ; pai sai imediatamente -> filho e orfanizado
    SYSCALL #2         ; exit
child:
    MV W5 #7           ; marcador do filho (para depuracao)
    MV W8 #2           ; status = 2
    SYSCALL #2         ; exit