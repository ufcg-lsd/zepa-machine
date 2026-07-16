kill_int:

    LDD W9, [buffer_addr]

    LDD W1, [partition_number]

    CMP W9, W1
    BGT not_valid
    BEQ not_valid

    MV  W2, #84
    MUL W3, W9, W2
    LDD W4, [pcb_v]
    ADD W3, W3, W4

    MV  W5, #80
    ADD W5, W3, W5
    LDB W6, [W5]

    MV  W8, #1
    AND W7, W6, W8

    MV  W0, #0
    CMP W7, W0
    BEQ not_valid

    MV  W8, #2
    AND W7, W6, W8

    CMP W7, W8
    BEQ not_valid

    MV  W5, #56
    ADD W5, W3, W5

    MV  W6, #2
    STORE W6, [W5]

    JUMP kill

not_valid:
    JUMP schedule