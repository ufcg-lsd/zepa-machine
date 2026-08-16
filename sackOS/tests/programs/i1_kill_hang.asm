; I1 - kill_hang: loop infinito puro; so para via kill externo do harness
; expected: w9=2 (exit code 2 = external kill)
_start:
loop:
    JUMP loop          ; nunca termina