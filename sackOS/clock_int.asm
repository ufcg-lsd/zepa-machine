clock_int:
  LDD W0, #clock_interrupt_count    ; w0 = clock_interrupt_count

  MV W1, #1
  ADD W0, W0, W1                    ; clock_interrupt_count += 1

  LDD W1, #TIME_SLICE
  CMP W0, W1
  BLT #2                            ; if clock_interrupt_count < TIME_SLICE
  JUMP clk_no_preempt

  STRD W0, #clock_interrupt_count
  LDD W0, #SCRATCH_0
  LDD W1, #SCRATCH_1
  MRET

  clk_no_preempt:
    MV W0, #0
    STRD W0, #clock_interrupt_count ; clock_interrupt_count = 0

    LDD W0, #running_pid
    MV W1, #-1
    CMP W0, W1
    BEQ end_clock                   ; if running_pid != -1 

    LDD W1, #running_pid    
    MV W0, #84                      ; pcb_size
    MUL W1, W0, W1                  ; W1 = running_pid * pcb_size
    
    MV W0, #pcb_v
    ADD W1, W0, W1                  ; W1 = pcb_v[running_pid] initial address

    MV W0, #68
    ADD W1, W0, W1
    STORE ESR, W1                   ; saving SR

    MV W0, #4
    SUB W1, W1, W0
    STORE SP, W1                    ; saving SP

    SUB W1, W1, W0
    STORE EPC, W1                   ; saving PC

    SUB W1, W1, W0
    STORE W9, W1                    ; saving W9

    SUB W1, W1, W0
    STORE W8, W1                    ; saving W8

    SUB W1, W1, W0
    STORE W7, W1                    ; saving W7
    
    SUB W1, W1, W0
    STORE W6, W1                    ; saving W6
    
    SUB W1, W1, W0
    STORE W5, W1                    ; saving W5

    SUB W1, W1, W0
    STORE W4, W1                    ; saving W4

    SUB W1, W1, W0
    STORE W3, W1                    ; saving W3

    SUB W1, W1, W0
    STORE W2, W1                    ; saving W2

    LDD W3, #SCRATCH_1
    SUB W1, W1, W0
    STORE W3, W1                    ; saving W1

    LDD W3, #SCRATCH_0
    SUB W1, W1, W0
    STORE W3, W1                    ; saving W0

  end_clock:
    JUMP schedule