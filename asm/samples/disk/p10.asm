; Teste com múltiplas operações
MV W0, #1
MV W1, #2
ADD W2, W0, W1
SUB W3, W1, W0
CMP W2, W3
BEQ #28         ; Não deve pular
MV W4, #10
MV W5, #20
HALT