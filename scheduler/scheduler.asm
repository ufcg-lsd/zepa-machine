_start:
    MV W3, #4        ; Quantum = 4 instruções por processo
 
    ; Inicializa a fila de processos (endereços de início de cada processo)
    MV W0, #0x1000  ; 
    MV W1, #0x2000  ; 
    MV W2, #0x3000  ; 
    ; Armazena os endereços na memória (fila de processos)
    STORE W0, 0x0000  ; 
    STORE W1, 0x0004  ; 
    STORE W2, 0x0008  ; 

    MV W4, #0       ; Índice atual da fila
