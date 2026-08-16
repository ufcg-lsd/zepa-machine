; D6 - wait_status_addr_oob: status_addr fora da particao gera fault
; expected: w9=1 (exit code 1 = general fault)
_start:
    MV W0 #0           ; W0 = zero
    ADD W8 SP W0       ; W8 = SP = partition_size (fora da particao)
    SYSCALL #1         ; wait -> check de limite falha -> fault -> exit 1