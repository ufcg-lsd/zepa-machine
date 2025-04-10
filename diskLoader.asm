MV W0, 0      ; 0
MV W1, 32     ; 4
MV W3, 0      ; 8
MV W4, 0      ; 12

D2M W0, 0     ; 16
CMP W3, W0    ; 20
BEQ 48        ; 24

ADD W1, W1, W4 ; 28
JUMP 24       ; 32

JUMP 32       ;36