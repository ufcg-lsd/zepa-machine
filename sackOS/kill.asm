kill:
  MV W0, #pcb_v           ; W0 = pcb_v initial address
  MV W1, #84              ; W1 = pcb_size
  MUL W2, W1, W9
  ADD W2, W0, W2          ; W2 = pcb_v[pid].parent_pid address

  LOAD W3, W2             ; W3 = pcb_v[pid].parent_pid

  MV W4, #-1
  CMP W3, W4
  BEQ kill_no_parent      ; if parent != -1

    MUL W3, W3, W1
    ADD W3, W3, W0        ; W3 = pcb_v[parent] initial address

    MV W4, #80
    ADD W3, W3, W4        ; W3 = pcb_v[parent].flags address

    LDB W4, W3           ; W4 = pcb_v[parent].flags
    MV W5, #4
    AND W4, W4, W5        ; W4 = pcb_v[parent].is_waiting

    CMP W4, W5
    BEQ #2                ; if pcb_v[parent].is_waiting
    JUMP kill_not_waiting

      MV W4, #9           ; mapped = 1, zombie = 0, waiting = 0, state = ready
      STRB W4, W3

      MV W4, #8
      SUB W3, W3, W4      ; W3 = pcb_v[parent].BASE address
      LOAD W4, W3         ; W4 = pcb_v[parent].BASE

      MV W5, #56
      SUB W3, W3, W5      ; W3 = pcb_v[parent].status_addr address
      LOAD W5, W3         ; W5 = pcb_v[parent].status_addr

      ADD W4, W4, W5      ; W4 = base+status_addr address

      MV W5, #56
      ADD W2, W2, W5      ; W2 = pcb_v[pid].W9 address
      LOAD W5, W2         ; W5 = pcb_v[pid].W9

      STORE W5, W4        ; memory[pcb_v[parent].BASE+pcb_v[parent].status_addr] = pcb_v[pid].w9

      MV W5, #12
      SUB W3, W3, W5      ; W3 = pcb_v[parent].child address
      LOAD W4, W3         ; W4 = pcb_v[parent].child

      CMP W4, W9
      BEQ #2              ; if pcb_v[parent].child = pid
      JUMP #6

        MV W5, # 44
        SUB W2, W2, W5    ; W2 = pcb_v[pid].next_sibling address
        LOAD W4, W2       ; W4 = pcb_v[pid].next_sibling
        STORE W4, W3      ; pcb_v[parent].child = pcb_v[pid].next_sibling
        ADD W2, W2, W5    ; W2 = pcb_v[pid].W9 address


      MV W5, #48
      SUB W2, W2, W5      ; W2 = pcb_v[pid].prev_sibling address
      LOAD W3, W2         ; W3 = pcb_v[pid].prev_sibling

      MV W5, #4
      ADD W2, W2, W5      ; W2 = pcb_v[pid].next_sibling address
      LOAD W4, W2         ; W4 = pcb_v[pid].next_sibling


      MV W5, #-1
      CMP W3, W5
      BEQ #6              ; if pcb_v[pid].prev_sibling != -1

        MUL W5, W3, W1
        ADD W5, W5, W0    ; W5 = pcb_v[pcb_v[pid].prev_sibling] initial address
        MV W6, #12
        ADD W5, W5, W6    ; W5 = pcb_v[pcb_v[pid].prev_sibling].next_sibling address

        STORE W4, W5      ; pcb_v[pcb_v[pid].prev_sibling].next_sibling = pcb_v[pid].next_sibling


      MV W5, #-1
      CMP W4, W5
      BEQ #6              ; if pcb_v[pid].next_sibling != -1

        MUL W5, W4, W1
        ADD W5, W5, W0    ; W5 = pcb_v[pcb_v[pid].next_sibling] initial address
        MV W6, #8
        ADD W5, W5, W6    ; W5 = pcb_v[pcb_v[pid].next_sibling].prev_sibling address

        STORE W3, W5      ; pcb_v[pcb_v[pid].next_sibling].prev_sibling = pcb_v[pid].prev_sibling

      
      MV W3, #68
      ADD W2, W2, W3      ; W2 = pcb_v[pid].flags address
      LDB W3, W2         ; W3 = pcb_v[pid].flags

      MV W4, #-2          ; everything 1 except the LSB
      AND W3, W3, W4      ; is_mapped = 0
      STRB W3, W2        ; pcb_v[pid].is_mapped = 0

      JUMP kill_orphanize ; execution after the elses

    kill_not_waiting:
      ; here W2 = pcb_v[pid].parent_pid address
      MV W3, #80
      ADD W2, W2, W3      ; W2 = pcb_v[pid].flags address
      LDB W3, W2         ; W3 = pcb_v[pid].flags

      MV W4, #2
      OR W3, W3, W4       ; is_zombie = 1
      STRB W3, W2        ; pcb_v[pid].is_zombie = 1
      JUMP kill_orphanize ; execution after the elses


  kill_no_parent:
    ; here W2 = pcb_v[pid].parent_pid address
    MV W3, #80
    ADD W2, W2, W3      ; W2 = pcb_v[pid].flags address
    LDB W3, W2         ; W3 = pcb_v[pid].flags

    MV W4, #-2          ; everything 1 except the LSB
    AND W3, W3, W4      ; is_mapped = 0
    STRB W3, W2        ; pcb_v[pid].is_mapped = 0

  
  kill_orphanize:
    ; here W2 = pcb_v[pid].flags address

    MV W3, #76
    SUB W2, W2, W3      ; W2 = pcb_v[pid].child address
    LOAD W3, W2          ; W3 = pcb_v[pid].child (curr_child)
    MV W7, #-1

    orphanize_loop:
      CMP W3, W7
      BEQ kill_end

      MUL W4, W3, W1
      ADD W4, W4, W0    ; W4 = pcb_v[curr_child].parent_pid address
      STORE W7, W4      ; pcb_v[curr_child].parent_pid = -1

      MV W6, #8
      ADD W4, W4, W6    ; W4 = pcb_v[curr_child].prev_sibling address
      STORE W7, W4      ; pcb_v[curr_child].prev_sibling = -1

      MV W6, #4
      ADD W4, W4, W6    ; W4 = pcb_v[curr_child].next_sibling address
      LOAD W5, W4       ; W5 = pcb_v[curr_child].next_sibling = next
      STORE W7, W4      ; pcb_v[curr_child].next_sibling = -1

      MV W8, #0
      ADD W3, W5, W8    ; W3 = curr_child = next
      JUMP orphanize_loop

  kill_end:
    JUMP schedule
