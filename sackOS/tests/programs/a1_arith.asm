; A1 - arith: operacoes aritmeticas basicas
; expected: w2=13, w3=7, w4=30, w5=3, w6=3, w9=42 (exit code)
_start:
    MV W0 #10          ; W0 = 10
    MV W1 #3           ; W1 = 3
    ADD W2 W0 W1       ; W2 = 13
    SUB W3 W0 W1       ; W3 = 7
    MUL W4 W0 W1       ; W4 = 30
    UDIV W5 W0 W1      ; W5 = 3 (10/3)
    SDIV W6 W0 W1      ; W6 = 3 (10/3)
    MV W8 #42          ; status = 42
    SYSCALL #2         ; exit