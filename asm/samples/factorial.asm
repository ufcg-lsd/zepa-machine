_start:
    MV W0, #1       ; Initialize result (W0) with 1
    MV W1, #5       ; Example calculating 5
    
_loop:
    CMP W1, #1      ; If W1 == 1, end
    JZ _end         ; Jump to end

    MUL W0, W0, W1  ; W0 = W0 * W1
    SUB W1, W1, #1  ; W1 = W1 - 1
    JUMP _loop      ; Loop

_end:
    ; Factorial stored in W0
