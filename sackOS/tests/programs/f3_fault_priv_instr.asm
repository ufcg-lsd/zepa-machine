; F3 - fault_priv_instr: instrucao privilegiada (MRET) em modo usuario
; expected: w9=1 (exit code 1 = general fault)
_start:
    MRET               ; privilegiada -> fault -> exit 1
    JUMP #0            ; inalcancavel