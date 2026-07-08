syscall_int:
    MV W1, #0
    CMP W9, W1
    BEQ syscall_fork

    MV W1, #1
    CMP W9, W1
    BEQ syscall_wait

    MV W1, #2
    CMP W9, W1
    BEQ syscall_exit

    MV W1, #3
    CMP W9, W1
    BEQ syscall_getPID

    MV W1, #4
    CMP W9, W1
    BEQ syscall_rele

    JUMP schedule
    
    MV W0, #pcb_v               ; set pcb_v adress
    LDD W1, running_pid         ; current running pid
    MV W2, #84                  ; size of pcb
    MUL W1, W1, W2              ; running pid * 84
    ADD W0, W0, W1              ; pcb_v + (running pid * 84)

    MV W3, #56                  ; offset of pc in pcb to get W9
    ADD W0, W0, W3              ; pcb_v + (running pid * 84) + 56

    MV W3, #-1                  ; value -1
    STORE W3, W0                ; pcb[running_pid].w9 = -1

  
syscall_fork:
    MV W4, #0                   ; W4 = pid = 0 (contador do loop)
    LDD W5, PARTITION_NUMBER    ; W5 = PARTITION_NUMBER (limite do loop)
    MV W2, #1                   ; W2 = 1 (passo do loop)

    fork_loop:
        CMP W4, W5              ; Compara pid (W4) com PARTITION_NUMBER (W5)
        BEQ fork_end           ; Se pid == limite, sai do loop

        MV W0, #pcb_v           ; W0 = endereço base do vetor pcb_v
        MV W1, W4               ; W1 = pid atual (copia W4)
        MV W3, #84              ; W3 = tamanho de cada PCB (84 bytes)
        MUL W1, W1, W3          ; W1 = pid * 84
        ADD W0, W0, W1          ; W0 = endereço de pcb_v[pid]

        ; flags está no offset 80 do PCB
        MV W3, #80              ; W3 = offset das flags
        ADD W5, W0, W3          ; W5 = endereço das flags
        LDB W6, W5              ; W6 = byte de flags de pcb_v[pid]

        ; código de verificação do is_mapped (bit 0 das flags)
        MV W3, #0               ; W3 = 0
        MV W1, #1               ; W1 = 1 (máscara para bit 0)
        AND W6, W6, W1          ; W0 = is_mapped
        CMP W6, W3              ; compara is_mapped com 0
        BGT next_iteration      ; se is_mapped > 0 (1), pula para a próxima iteração

        ; inicializar o PCB e carregar buffer
        MV W5, #9               ; inicializa flags
        LDD W7, running_pid     ;
        STORE W7, W0            ; assign o parent_pid para o running_pid
       
        LDD W7, 

    next_iteration:
        ADD W4, W4, W2          ; pid++ (W4 = W4 + W2)
        JUMP fork_loop

    fork_end:
        MV W0, #pcb_v               ; set pcb_v adress
        LDD W1, running_pid         ; current running pid
        MV W2, #84                  ; size of pcb
        MUL W1, W1, W2              ; running pid * 84
        ADD W0, W0, W1              ; pcb_v + (running pid * 84)

        MV W3, #56                  ; offset of pc in pcb to get W9
        ADD W0, W0, W3              ; pcb_v + (running pid * 84) + 56

        MV W3, #-1                  ; value -1
        STORE W3, W0                ; pcb[running_pid].w9 = -1