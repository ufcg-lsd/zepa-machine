_start:
    MV W1, #2         ; Move the value 2 into W1
    MV W2, #5         ; Move the value 5 into W2

    JUMP 0x14         ; Jump to address 0x14
    MV W1, #30        ; Move the value 30 into W1 (if no jump)
    MV W2, #40        ; Move the value 40 into W2 (if no jump)

    ; Next position is 0x14
    ADD W0, W1, W2    ; Add W1 and W2, store the result in W0

_end:
    ; End of program