; A2 - branch_loop: desvio condicional e loop (soma 1..100 = 5050)
; expected: w0=5050, w9=5050 (exit code)
_start:
    MV W0 #0           ; W0 = soma
    MV W1 #1           ; W1 = incremento
    MV W2 #101         ; W2 = limite
    MV W4 #0           ; W4 = zero (copia)
    MV W5 #1
loop:
    ADD W0 W0 W1       ; soma += 1
    ADD W1 W1 W5
    CMP W1 W2         ; soma vs 100
    BLT loop          ; se menor, repete
    ADD W8 W0 W4       ; status = soma (5050)
    SYSCALL #2         ; exit