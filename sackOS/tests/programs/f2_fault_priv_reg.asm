; F2 - fault_priv_reg: acesso a registrador privilegiado em modo usuario
; expected: w9=1 (exit code 1 = general fault)
_start:
    MV W0 #1           ; W0 = 1
    ADD W5 W0 UPTR     ; acessa UPTR (privilegiado) -> fault -> exit 1
    JUMP #0            ; inalcancavel