; Teste de saltos condicionais
MV W0, #5
MV W1, #5
CMP W0, W1      ; SR = 0 (igual)
BEQ 24          ; Deve pular para HALT (24 é o endereço do HALT)
MV W2, 99       ; Não deve executar
HALT