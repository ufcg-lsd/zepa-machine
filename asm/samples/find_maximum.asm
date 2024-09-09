_start:
    MV W0, #3         ; Load the first number into register W0
    MV W1, #7         ; Load the second number into register W1
    MV W2, #2         ; Load the third number into register W2
    MV W3, #9         ; Load the fourth number into register W3
    MV W4, W0         ; Initialize W4 with the first value

    CMP W1, W4        ; Compare W1 with W4
    JG _update_W1     ; If W1 > W4, jump to _update_W1
    JUMP _check_W2

_update_W1:
    MV W4, W1         ; W4 = W1

_check_W2:
    CMP W2, W4        ; Compare W2 with W4
    JG _update_W2
    JUMP _check_W3

_update_W2:
    MV W4, W2         ; W4 = W2

_check_W3:
    CMP W3, W4        ; Compare W3 with W4
    JG _update_W3
    JUMP _end

_update_W3:
    MV W4, W3         ; W4 = W3

_end:
    ; End of program
