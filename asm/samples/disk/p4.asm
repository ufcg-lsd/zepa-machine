; Teste de saltos condicionais
MV W0, #5
MV W1, #5
CMP W0, W1      ; SR = 0 (igual)
BEQ #16         ; Deve pular para HALT
MV W2, #99      ; Não deve executar
HALT