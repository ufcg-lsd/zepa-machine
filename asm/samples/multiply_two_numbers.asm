_start:
    MV W1, #4       ; Load the number 4 into W1
    MV W2, #6       ; Load the number 6 into W2
    MV W0, #0       ; Initialize W0 with 0, where the result will be stored
    
    ADD W0, W0, W1  ; W0 = W0 + W1 (sum 1st time)
    ADD W0, W0, W1  ; W0 = W0 + W1 (sum 2nd time)
    ADD W0, W0, W1  ; W0 = W0 + W1 (sum 3rd time)
    ADD W0, W0, W1  ; W0 = W0 + W1 (sum 4th time)
    ADD W0, W0, W1  ; W0 = W0 + W1 (sum 5th time)
    ADD W0, W0, W1  ; W0 = W0 + W1 (sum 6th time)

_end:
    ; End of program
