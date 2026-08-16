; G2 - infinite_hang: loop infinito; harness mata apos timeout (kill externo)
; expected: w9=2 (exit code 2 = external kill)
_start:
    MV W0 #0           ; W0 = zero
hang:
    ADD W0 W0 W0       ; trabalho "inutil" para ocupar a CPU
    JUMP hang          ; nunca termina