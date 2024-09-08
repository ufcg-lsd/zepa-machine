## 1. Adicionar dois números
```
_start:
    MV W1, #5       ; Carrega o número 5 em W1
    MV W2, #3       ; Carrega o número 3 em W2
    ADD W0, W1, W2  ; Soma W1 e W2, resultado em W0
```

## 2. Multiplicar dois números
```
_start:
    MV W1, #4       ; Carrega o número 4 em W1
    MV W2, #6       ; Carrega o número 6 em W2
    MUL W0, W1, W2  ; Multiplica W1 e W2, resultado em W0
```

## 3. Calcular fatorial
```
_start:
    MV W0, #1       ; Inicializa resultado (W0) com 1
    MV W1, #5       ; Exemplo calculando 5
    
_loop:
    CMP W1, #1      ; Se W1 == 1, fim
    JZ _end         ; Salta para o final

    MUL W0, W0, W1  ; W0 = W0 * W1
    SUB W1, W1, #1  ; W1 = W1 - 1
    JUMP _loop      ; Loop

_end:
    ; Fatorial armazenado em W0
```

## 6. Calcular o GCD (Greatest Common Divisor)
```
_start:
    MV W1, #56      ; Primeiro número
    MV W2, #98      ; Segundo número

_gcd_loop:
    CMP W2, #0      ; Se W2 == 0, fim
    JZ _end         ; Salta para o final

    MOD W0, W1, W2  ; W0 = W1 % W2
    MV W1, W2       ; W1 = W2
    MV W2, W0       ; W2 = W0
    JUMP _gcd_loop  ; Repete

_end:
    ; GCD armazenado em W1
```