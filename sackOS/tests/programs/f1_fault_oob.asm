; F1 - fault_oob: acesso de memoria fora da particao gera fault
; expected: w9=1 (exit code 1 = general fault)
_start:
    MV W0 #0           ; W0 = zero
    ADD W1 SP W0       ; W1 = SP = partition_size (fora da particao)
    STORE W0 W1        ; store fora do limite -> fault -> exit 1
    JUMP #0            ; inalcancavel