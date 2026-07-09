_start:
    MV W1 #67
    MV W2 #0
    MV W3 #1
    JUMP _jumpAndLoop

_jumpAndLoop:
    ADD W2 W2 W3
    CMP W1 W2 
    BEQ _clean
    JUMP _jumpAndLoop

_clean:
    MV W2 #1
    JUMP _jumpAndLoop

_end:
