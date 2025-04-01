; Teste de LOAD/STORE
MV W0, #42
STORE W0, #0    ; Armazena 42 no endereço 0
LOAD W1, #0     ; Carrega do endereço 0 para W1
HALT