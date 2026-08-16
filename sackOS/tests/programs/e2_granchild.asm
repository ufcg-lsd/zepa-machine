; E2 - granchild: filho faz fork de um neto; cadeia de wait (3+30=33)
; expected: w7=33, w9=33 (exit do pai)
_start:
    MV W3 #-2          ; marcador de filho
    MV W0 #0           ; W0 = zero
    MV W7 #4
    SUB W8 SP W7       ; W8 = status_addr
    STORE W0 W8        ; touch -> mapeia a página de status_addr
    SYSCALL #0         ; fork pid0 -> child1
    CMP W9 W3
    BEQ child1
parent:
    SYSCALL #1         ; wait child1 -> mem[status_addr] = 33
    LOAD W4 W8         ; W4 = 33
    ADD W7 W4 W0       ; W7 = 33
    ADD W8 W7 W0       ; status = 33
    SYSCALL #2         ; exit
child1:
    SYSCALL #0         ; fork child1 -> grandchild
    CMP W9 W3
    BEQ grandchild
    SYSCALL #1         ; wait grandchild -> mem[status_addr] = 3
    LOAD W4 W8         ; W4 = 3
    MV W5 #30
    ADD W4 W4 W5       ; W4 = 33
    ADD W8 W4 W0       ; status = 33
    SYSCALL #2         ; exit
grandchild:
    MV W8 #3           ; status = 3
    SYSCALL #2         ; exit