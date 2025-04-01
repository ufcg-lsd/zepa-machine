; Teste de loop simples
MV W0, #0       ; Contador
MV W1, #5       ; Limite
MV W2, #1       ; Incremento

LOOP:
ADD W0, W0, W2  ; Incrementa contador
CMP W0, W1      ; Compara com limite
BLT #4          ; Volta para LOOP se W0 < W1 (endereço 4)
HALT