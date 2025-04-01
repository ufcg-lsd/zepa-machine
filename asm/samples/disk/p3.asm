; Teste de comparações
MV W0, #10
MV W1, #5
CMP W0, W1      ; Deve setar SR para 2 (W0 > W1)
HALT