MV W1, #4       ; Load the number 4 into W1
MV W2, #6       ; Load the number 6 into W2
MV W0, #0       ; Initialize W0 with 0

SUB W0, W0, W1  ; W0 = W0 + W1 (1st time)
CMP W0, W1
JUMP #65
ADD W0, W0, W1  ; W0 = W0 + W1 (2nd time)
ADD W0, W0, W1  ; W0 = W0 + W1 (3rd time)
ADD W0, W0, W1  ; W0 = W0 + W1 (4th time)
ADD W0, W0, W1  ; W0 = W0 + W1 (5th time)
ADD W0, W0, W1  ; W0 = W0 + W1 (6th time)