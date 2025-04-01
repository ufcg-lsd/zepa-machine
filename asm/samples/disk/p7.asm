; Teste de chamada de procedimento
MV W0, #10
JUMP #12        ; Pula para PROCEDURE
MV W1, #20      ; Retorno aqui
HALT

PROCEDURE:
ADD W0, W0, W0  ; Dobra W0
RET             ; Retorna