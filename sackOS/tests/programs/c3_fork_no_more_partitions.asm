; C3 - fork_no_more_partitions: fork falha quando nao ha particoes livres
; (exige config com partition_number == 1)
; expected: w5=7 (marcador), w9=4294967295 (exit -1)
_start:
    MV W3 #-2          ; marcador de filho
    MV W0 #0           ; W0 = zero
    MV W7 #4
    SUB W8 SP W7       ; W8 = status_addr
    SYSCALL #0         ; fork (falha: retorna -1)
    CMP W9 W3
    BEQ child          ; nao deve acontecer nesta config
    MV W5 #7           ; marcador do caso falho
    MV W8 #-1          ; status = -1
    SYSCALL #2         ; exit -> W9 = 0xFFFFFFFF
child:
    MV W8 #3           ; status = 3
    SYSCALL #2         ; exit