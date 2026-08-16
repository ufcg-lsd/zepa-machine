; A3 - bitwise_shift: AND/OR/XOR e SHL/SHA com shift negativo
; expected: w2=8, w3=4294967295, w4=4294967287, w6=268435455, w7=4294967295, w9=1
_start:
    MV W0 #-8          ; W0 = 0xFFFFFFF8 (-8)
    MV W1 #15          ; W1 = 0x0F
    AND W2 W0 W1       ; W2 = 0xFFFFFFF8 & 0x0F = 8
    OR  W3 W0 W1       ; W3 = 0xFFFFFFFF
    XOR W4 W0 W1       ; W4 = 0xFFFFFFF7
    MV W5 #-4          ; shift = -4 (desloca para a direita)
    SHL W6 W0 W5       ; logico: 0xFFFFFFF8 >> 4 = 0x0FFFFFFF
    SHA W7 W0 W5       ; aritmetico: 0xFFFFFFF8 >> 4 = 0xFFFFFFFF
    MV W8 #1           ; status = 1
    SYSCALL #2         ; exit