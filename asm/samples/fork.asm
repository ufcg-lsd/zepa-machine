MV W0, #0         ; Inicializa o acumulador do pai (W0) com 0
MV W1, #0         ; Inicializa o acumulador do filho (W1) com 0
MV W2, #1         ; Define W2 com 1 para ser usado como valor de incremento
MV W3, #-2        ; Define W3 com -2 para usarmos na instrução de comparação (CMP)

SYSCALL #0        ; Aciona a interrupção de software. O retorno do SO ficará em W9

CMP W9, W3        ; Compara o valor de retorno em W9 com 0 (armazenado em W3)
BEQ child_loop    ; Se W9 for igual a 0 (flag Z setada), pula para a rotina do filho

parent_loop:
ADD W0, W0, W2    ; Incrementa W0 em 1 (W0 = W0 + W2)
JUMP parent_loop  ; Loop infinito: retorna para o início da rotina do pai

child_loop:
ADD W1, W1, W2    ; Incrementa W1 em 1 (W1 = W1 + W2)
JUMP child_loop   ; Loop infinito: retorna para o início da rotina do filho