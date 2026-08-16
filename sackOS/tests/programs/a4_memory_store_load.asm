; A4 - memory_store_load: STORE/LOAD (word) e STRB/LDB/LDSB, LDD/STRD
; expected: w4=23130, w5=90, w6=90, w7=23130, w9=2
_start:
    MV W0 #0           ; W0 = zero
    MV W7 #4           ; W7 = 4
    SUB W1 SP W7       ; W1 = size-4 (ultima palavra da particao)
    MV W2 #0x5A5A      ; W2 = 23130
    STORE W2 W1        ; mem[W1] = 23130
    LOAD W4 W1         ; W4 = 23130
    STRB W2 W1         ; mem[W1] byte = 0x5A
    LDB  W5 W1         ; W5 = 90
    LDSB W6 W1         ; W6 = 90
    STRD W2 #800       ; mem[800] = 23130
    LDD  W7 #800       ; W7 = 23130
    MV W8 #2           ; status = 2
    SYSCALL #2         ; exit