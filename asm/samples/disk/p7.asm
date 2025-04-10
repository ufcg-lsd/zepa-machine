 ; Teste de exceção de memória
MV W0, #100
STORE W0, #1000 ; Deve causar exceção se memória < 1000 bytes
HALT