MV W1, #-4      ; W1 = 0xFFFFFFFC (-4 in two's complement)
MV W2, #2       ; W2 = 2 (positive shift amount)
MV W3, #-2      ; W3 = -2 (negative shift amount)

SHL W4, W1, W2  ; W4 = W1 << 2  (0xFFFFFFF0)
SHL W5, W1, W3  ; W5 = W1 >> 2 logical (0x3FFFFFFF)

SHA W6, W1, W2  ; W6 = W1 << 2  (0xFFFFFFF0)
SHA W7, W1, W3  ; W7 = W1 >> 2 arithmetic (0xFFFFFFFF)