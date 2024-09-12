_start:
    MV W0, #48        ; Load the first number into W0
    MV W1, #18        ; Load the second number into W1

_loop:
    CMP W0, W1        ; Compare W0 with W1
    JZ _end           ; If W0 == W1, jump to _end
    JG _subtract_W0   ; If W0 > W1, jump to _subtract_W0
    SUB W1, W1, W0    ; Subtract W0 from W1
    JUMP _loop

_subtract_W0:
    SUB W0, W0, W1    ; Subtract W1 from W0
    JUMP _loop

_end:
    ; End of program