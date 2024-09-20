_start:
    MV W1, #4       ; Load the number 4 into W1
    MV W2, #6       ; Load the number 6 into W2
    MV W0, #0       ; Inicializa W0 com 0, onde o resultado será armazenado
    
    ADD W0, W0, W1  ; W0 = W0 + W1 (soma 1ª vez)
    ADD W0, W0, W1  ; W0 = W0 + W1 (soma 2ª vez)
    ADD W0, W0, W1  ; W0 = W0 + W1 (soma 3ª vez)
    ADD W0, W0, W1  ; W0 = W0 + W1 (soma 4ª vez)
    ADD W0, W0, W1  ; W0 = W0 + W1 (soma 5ª vez)
    ADD W0, W0, W1  ; W0 = W0 + W1 (soma 6ª vez)

_end:
    ; End of program
