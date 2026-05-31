MV W1 #15               ; W1 = 15
MV W2 #20               ; W2 = 20

MV W5 #16               ; Stores the address for the JUMP _end instruction
JUMP COMPARE_FUNC       ; Jump to the function

JUMP _end               ; End of main program execution

COMPARE_FUNC:
CMP W1, W2              ; Compare W1 (15) and W2 (20). 

BEQ SET_EQUAL           ; If SR=1 (Equal), jump to SET_EQUAL
BLT SET_LESS            ; If SR=2 (Less), jump to SET_LESS
BGT SET_GREATER         ; If SR=4 (Greater), jump to SET_GREATER

SET_EQUAL:
MV W0 #1                ; W0 = 1
JMPR W5                 ; Jump to address in W5

SET_LESS:
MV W0 #2                ; W0 = 2
JMPR W5                 ; Jump to address in W5

SET_GREATER:
MV W0 #4                ; W0 = 4
JMPR W5                 ; Jump to address in W5

_end:
