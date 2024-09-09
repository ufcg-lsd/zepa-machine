_start:
    MV W0, #0       ; Initialize sum as 0
    MV W1, #1       ; Example number 1
    MV W2, #2       ; Example number 2
    MV W3, #3       ; Example number 3

    ADD W0, W0, W1  ; Add W1
    ADD W0, W0, W2  ; Add W2
    ADD W0, W0, W3  ; Add W3
    JUMP _end       ; Jump to end

_end:
    ; End of program