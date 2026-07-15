_schedule:

    ; valida running_pid

    MV W0, #-1
    CMP W9, W0
    BEQ _schedule_reset_running_pid

    ; valida running_pid

    LDD W0, [PARTITION_NUMBER_ADDRESS]

    CMP W9, W0
    BEQ _schedule_reset_running_pid
    BGT _schedule_reset_running_pid

    JUMP _schedule_check_current_process


schedule_reset_running_pid:

    MV W9, #0


schedule_check_current_process:

    ; verifica se processo atual esta running
    ; W1 = pcb_v[running_pid]

    MV W0, #84
    MUL W1, W9, W0

    MV W0, #PCB_VECTOR_ADDRESS
    ADD W1, W0, W1

    ; W2 = pcb_v[running_pid].flags

    MV W0, #80
    ADD W0, W1, W0

    LDB W2, [W0]

    MV W5, #0x18
    AND W5, W2, W5

    MV W0, #0
    CMP W5, W0
    BEQ schedule_set_current_ready

    JUMP schedule_calculate_limit_pid


schedule_set_current_ready:

    ; marca como ready

    MV W5, #-25
    AND W2, W2, W5

    MV W5, #0x08
    OR W2, W2, W5

    MV W0, #80
    ADD W0, W1, W0

    STRB W2, [W0]


_schedule_calculate_limit_pid:

    ; limit_pid = running_pid + 1

    MV W0, #1
    ADD W3, W9, W0

    ; wrap limit_pid se necessario

    LDD W0, [PARTITION_NUMBER_ADDRESS]

    CMP W3, W0
    BEQ _schedule_limit_pid_zero

    JUMP _schedule_limit_pid_ready


_schedule_limit_pid_zero:

    MV W3, #0


_schedule_limit_pid_ready:

    ; curr_pid = limit_pid

    MV W4, W3

_schedule_search_loop:

    ; busca processo ready
    ; W1 = &pcb_v[curr_pid]

    MV W0, #84
    MUL W1, W4, W0

    MV W0, #PCB_VECTOR_ADDRESS
    ADD W1, W0, W1

    ; W2 = pcb_v[curr_pid].flags

    MV W0, #80
    ADD W0, W1, W0

    LDB W2, [W0]

    ; esta mapeado?

    MV W5, #0x01
    AND W5, W2, W5

    MV W0, #0x01
    CMP W5, W0
    BEQ _schedule_check_zombie

    JUMP _schedule_next_pid


_schedule_check_zombie:

    ; eh zombie?

    MV W5, #0x02
    AND W5, W2, W5

    MV W0, #0
    CMP W5, W0
    BEQ _schedule_check_ready

    JUMP _schedule_next_pid


_schedule_check_ready:

    ; esta ready?

    MV W5, #0x18
    AND W5, W2, W5

    MV W0, #0x08
    CMP W5, W0
    BEQ _schedule_process_found

    JUMP _schedule_next_pid


_schedule_process_found:

    ; processo encontrado, seta running e restaura

    MV W5, #-25
    AND W2, W2, W5

    MV W0, #80
    ADD W0, W1, W0

    STRB W2, [W0]

    ; W9 = curr_pid

    MV W9, W4

    ; W6 = &pcb_v[running_pid]

    MV W6, W1

    JUMP _schedule_restore_context


_schedule_next_pid:

    ; curr_pid++

    MV W0, #1
    ADD W4, W4, W0

    ; wrap curr_pid se necessario

    LDD W0, [PARTITION_NUMBER_ADDRESS]

    CMP W4, W0
    BEQ _schedule_wrap_pid

    JUMP _schedule_check_search_end


_schedule_wrap_pid:

    MV W4, #0


_schedule_check_search_end:

    ; while curr_pid != limit_pid

    CMP W4, W3
    BEQ _schedule_no_ready_process

    JUMP _schedule_search_loop


; precisa de futuro review a partir daqui
_schedule_no_ready_process:

    ; nenhum pronto, idle loop

    MV W9, #-1

    MV W0, #LOOP_ADDRESS
    MV EPC, W0

    MV W0, #16
    MV ESR, W0

    MRET


; W6 = endereco inicial do PCB selecionado
; W9 = running_pid (nao restaurado do PCB)
_schedule_restore_context:

    ; carrega registros do PCB
    ; W0 (20) -> scratch

    MV W0, #20
    ADD W8, W6, W0

    LOAD W0, [W8]

    STRD W0, [RESTORE_W0_SCRATCH_ADDRESS]

    ; PC (60)

    MV W0, #60
    ADD W8, W6, W0

    LOAD W0, [W8]
    MV EPC, W0

    ; SP (64)

    MV W0, #64
    ADD W8, W6, W0

    LOAD W0, [W8]
    MV SP, W0

    ; SR (68)

    MV W0, #68
    ADD W8, W6, W0

    LOAD W0, [W8]
    MV ESR, W0

    ; BASE (72)

    MV W0, #72
    ADD W8, W6, W0

    LOAD W0, [W8]
    MV BASE, W0

    ; LIMIT (76)

    MV W0, #76
    ADD W8, W6, W0

    LOAD W0, [W8]
    MV LIMIT, W0

    ; W1 (24)

    MV W0, #24
    ADD W8, W6, W0

    LOAD W1, [W8]

    ; W2 (28)

    MV W0, #28
    ADD W8, W6, W0

    LOAD W2, [W8]

    ; W3 (32)

    MV W0, #32
    ADD W8, W6, W0

    LOAD W3, [W8]

    ; W4 (36)

    MV W0, #36
    ADD W8, W6, W0

    LOAD W4, [W8]

    ; W5 (40)

    MV W0, #40
    ADD W8, W6, W0

    LOAD W5, [W8]

    ; W7 (48) -- restaura antes de W6 (W6 ainda tem endereco do PCB)

    MV W0, #48
    ADD W8, W6, W0

    LOAD W7, [W8]

    ; W8 (52) -- usa W0 como temporario para nao perder o endereco

    MV W0, #52
    ADD W0, W6, W0

    LOAD W8, [W0]

    ; W6 (44) -- restaurado por ultimo entre W1-W8

    MV W0, #44
    ADD W0, W6, W0

    LOAD W6, [W0]

    ; W0 <- scratch (ultimo registrador antes do MRET)

    LDD W0, [RESTORE_W0_SCRATCH_ADDRESS]

    MRET