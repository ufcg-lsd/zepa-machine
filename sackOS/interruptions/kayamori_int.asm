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

    MV W0, #pcb_v               ; set pcb_v adress
    LDD W1, running_pid         ; current running pid
    MV W2, #84                  ; size of pcb
    MUL W1, W1, W2              ; running pid * 84
    ADD W0, W0, W1              ; pcb_v + (running pid * 84)

    MV W3, #56                  ; offset of pc in pcb to get W9
    ADD W0, W0, W3              ; pcb_v + (running pid * 84) + 56

    MV W3, #-1                  ; value -1
    STORE W3, W0                ; pcb[running_pid].w9 = -1

    JUMP schedule

  
syscall_fork:
    MV W4, #0                   ; W4 = pid = 0 (contador do loop)
    LDD W5, PARTITION_NUMBER    ; W5 = PARTITION_NUMBER (limite do loop)
    MV W2, #1                   ; W2 = 1 (passo do loop)

    fork_loop:
        CMP W4, W5              ; Compara pid (W4) com PARTITION_NUMBER (W5)
        BEQ fork_end            ; Se pid == limite, sai do loop

        MV W0, #pcb_v           ; W0 = endereço base do vetor pcb_v
        MV W3, #84              ; W3 = tamanho de cada PCB (84 bytes)
        MUL W1, W4, W3          ; W1 = pid * 84
        ADD W0, W0, W1          ; W0 = endereço de pcb_v[pid]
        
        ; 1. LER as flags primeiro (offset 80 do PCB)
        MV W3, #80              ; W3 = offset das flags
        ADD W8, W0, W3          ; W8 = endereço das flags
        LDB W6, W8              ; W6 = byte de flags de pcb_v[pid]

        ; 2. Verificar se is_mapped (bit 0 das flags) == 0
        MV W3, #0               ; W3 = 0
        MV W1, #1               ; W1 = 1 (máscara para bit 0)
        AND W6, W6, W1          ; Filtra bit 0 de W6
        CMP W6, W3              ; compara is_mapped com 0
        BGT next_iteration      ; se is_mapped > 0 (1), pula para a próxima iteração

        ; --- Se is_mapped == 0 (Processo Livre): Inicializar o PCB ---

        ; a. Gravar as flags = 9 (is_mapped = 1, ready = 1) no offset 80
        MV W6, #9               ; W6 = 9
        STRB W6, W8             ; memory[pcb_v[pid].flags] = 9

        ; b. Assign o parent_pid para o running_pid (offset 0)
        LDD W7, running_pid     ; W7 = running_pid
        STORE W7, W0            ; memory[pcb_v[pid].parent_pid] = running_pid

        ; c. Assign os outros PIDs e status_addr como -1 (W7 = -1)
        MV W7, #-1              ; W7 = -1

        ; child_pid como -1 (offset 4)
        MV W3, #4
        ADD W8, W0, W3          ; W8 = endereço do child
        STORE W7, W8            ; memory[pcb_v[pid].child] = -1

        ; prev_sibling como -1 (offset 8)
        MV W3, #8
        ADD W8, W0, W3          ; W8 = endereço do prev_sibling
        STORE W7, W8            ; memory[pcb_v[pid].prev_sibling] = -1

        ; next_sibling como -1 (offset 12)
        MV W3, #12
        ADD W8, W0, W3          ; W8 = endereço do next_sibling
        STORE W7, W8            ; memory[pcb_v[pid].next_sibling] = -1

        ; status_addr como -1 (offset 16)
        MV W3, #16
        ADD W8, W0, W3          ; W8 = endereço do status_addr
        STORE W7, W8            ; memory[pcb_v[pid].status_addr] = -1

        ; d. Calcular endereço do PCB pai (running_pid)
        LDD W1, running_pid     ; W1 = parent_pid
        MV W3, #84              ; W3 = 84 (tamanho de cada PCB)
        MUL W1, W1, W3          ; W1 = parent_pid * 84
        MV W7, #pcb_v           ; W7 = base pcb_v
        ADD W7, W7, W1          ; W7 = endereço base do PCB pai

        ; e. Retorno do fork no pai: pcb_v[running_pid].w9 = pid (W4)
        MV W3, #56              ; offset de W9 no PCB
        ADD W8, W7, W3          ; W8 = endereço de pcb_v[running_pid].w9
        STORE W4, W8            ; pcb_v[running_pid].w9 = pid

        ; f. Retorno do fork no filho: pcb_v[pid].w9 = 0
        MV W3, #56              ; offset de W9 no PCB
        ADD W8, W0, W3          ; W8 = endereço de pcb_v[pid].w9
        MV W1, #0               ; W1 = 0
        STORE W1, W8            ; pcb_v[pid].w9 = 0

        ; g. Ajustar ponteiros de família: pcb_v[pid].next_sibling = pcb_v[running_pid].child
        MV W3, #4               ; offset do child
        ADD W8, W7, W3          ; W8 = endereço de pcb_v[running_pid].child
        LOAD W1, W8              ; W1 = pcb_v[running_pid].child
        
        MV W3, #12              ; offset do next_sibling
        ADD W8, W0, W3          ; W8 = endereço de pcb_v[pid].next_sibling
        STORE W1, W8            ; pcb_v[pid].next_sibling = pcb_v[running_pid].child

        ; h. if pcb_v[running_pid].child != -1:
        MV W3, #-1              ; W3 = -1
        CMP W1, W3              ; compara child com -1
        BEQ skip_prev_sibling   ; se for -1, pula
        
        ; pcb_v[pcb_v[running_pid].child].prev_sibling = pid
        MV W3, #84
        MUL W1, W1, W3          ; W1 = anterior_child * 84
        MV W3, #pcb_v
        ADD W1, W1, W3          ; W1 = endereço de pcb_v[anterior_child]
        
        MV W3, #8               ; offset de prev_sibling
        ADD W1, W1, W3          ; W1 = endereço de pcb_v[anterior_child].prev_sibling
        STORE W4, W1            ; pcb_v[anterior_child].prev_sibling = pid

    skip_prev_sibling:
        ; pcb_v[running_pid].child = pid (W4)
        MV W3, #4               ; offset de child
        ADD W8, W7, W3          ; W8 = endereço de pcb_v[running_pid].child
        STORE W4, W8            ; pcb_v[running_pid].child = pid

        ; i. Copiar registradores do Pai para o Filho (exceto BASE, LIMIT e W9 que é 0)
        ; W0 = Filho, W7 = Pai
        
        ; Copia W0 (offset 20)
        MV W3, #20
        ADD W8, W7, W3
        LOAD W1, W8
        ADD W8, W0, W3
        STORE W1, W8

        ; Copia W1 (offset 24)
        MV W3, #24
        ADD W8, W7, W3
        LOAD W1, W8
        ADD W8, W0, W3
        STORE W1, W8

        ; Copia W2 (offset 28)
        MV W3, #28
        ADD W8, W7, W3
        LOAD W1, W8
        ADD W8, W0, W3
        STORE W1, W8

        ; Copia W3 (offset 32)
        MV W3, #32
        ADD W8, W7, W3
        LOAD W1, W8
        ADD W8, W0, W3
        STORE W1, W8

        ; Copia W4 (offset 36)
        MV W3, #36
        ADD W8, W7, W3
        LOAD W1, W8
        ADD W8, W0, W3
        STORE W1, W8

        ; Copia W5 (offset 40)
        MV W3, #40
        ADD W8, W7, W3
        LOAD W1, W8
        ADD W8, W0, W3
        STORE W1, W8

        ; Copia W6 (offset 44)
        MV W3, #44
        ADD W8, W7, W3
        LOAD W1, W8
        ADD W8, W0, W3
        STORE W1, W8

        ; Copia W7 (offset 48)
        MV W3, #48
        ADD W8, W7, W3
        LOAD W1, W8
        ADD W8, W0, W3
        STORE W1, W8

        ; Copia W8 (offset 52)
        MV W3, #52
        ADD W8, W7, W3
        LOAD W1, W8
        ADD W8, W0, W3
        STORE W1, W8

        ; Copia PC (offset 60)
        MV W3, #60
        ADD W8, W7, W3
        LOAD W1, W8
        ADD W8, W0, W3
        STORE W1, W8

        ; Copia SP (offset 64)
        MV W3, #64
        ADD W8, W7, W3
        LOAD W1, W8
        ADD W8, W0, W3
        STORE W1, W8

        ; Copia SR (offset 68)
        MV W3, #68
        ADD W8, W7, W3
        LOAD W1, W8
        ADD W8, W0, W3
        STORE W1, W8

        ; j. Copiar a memória do Pai para o Filho
        ; Obter BASE do pai
        MV W3, #72              ; offset de BASE
        ADD W8, W7, W3
        LOAD W1, W8              ; W1 = BASE_pai

        ; Obter BASE do filho
        ADD W8, W0, W3
        LOAD W3, W8              ; W3 = BASE_filho

        LDD W6, PARTITION_SIZE  ; W6 = size = PARTITION_SIZE

        ; Loop de cópia de memória
        MV W8, #0               ; W8 = index = 0
        MV W2, #4               ; W2 = 4 (passo)
        
    mem_copy_loop:
        CMP W8, W6              ; compara index com size (W6)
        BEQ mem_copy_end        ; se index == size, terminou a cópia
        
        ADD W7, W1, W8          ; W7 = BASE_pai + index
        LOAD W5, W7              ; W5 = memory[BASE_pai + index]
        
        ADD W7, W3, W8          ; W7 = BASE_filho + index
        STORE W5, W7            ; memory[BASE_filho + index] = W5
        
        ADD W8, W8, W2          ; index++
        JUMP mem_copy_loop
        
    mem_copy_end:
        JUMP schedule           ; Fork concluído com sucesso, chama o escalonador

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


input_int:
    MV W4, #0                   ; W4 = pid = 0 (contador do loop)
    LDD W5, PARTITION_NUMBER    ; W5 = PARTITION_NUMBER (limite do loop)
    MV W2, #1                   ; W2 = 1 (passo do loop)

    input_loop:
        CMP W4, W5              ; Compara pid (W4) com PARTITION_NUMBER (W5)
        BEQ input_end           ; Se pid == limite, sai do loop

        MV W0, #pcb_v           ; W0 = endereço base do vetor pcb_v
        MV W3, #84              ; W3 = tamanho de cada PCB (84 bytes)
        MUL W1, W4, W3          ; W1 = pid * 84
        ADD W0, W0, W1          ; W0 = endereço de pcb_v[pid]
        
        ; 1. LER as flags primeiro (offset 80 do PCB)
        MV W3, #80              ; W3 = offset das flags
        ADD W8, W0, W3          ; W8 = endereço das flags
        LDB W6, W8              ; W6 = byte de flags de pcb_v[pid]

        ; 2. Verificar se is_mapped (bit 0 das flags) == 0
        MV W3, #0               ; W3 = 0
        MV W1, #1               ; W1 = 1 (máscara para bit 0)
        AND W6, W6, W1          ; Filtra bit 0 de W6
        CMP W6, W3              ; compara is_mapped com 0
        BGT next_input_iteration ; se is_mapped > 0 (1), pula para a próxima iteração

        ; Se is_mapped == 0 (Processo Livre): Inicializar o PCB

        ; a. Gravar as flags = 9 (is_mapped = 1, ready = 1) no offset 80
        MV W6, #9               ; W6 = 9
        STRB W6, W8             ; memory[pcb_v[pid].flags] = 9

        ; b. Assign os PIDs e status_addr como -1 (W7 = -1)
        MV W7, #-1              ; W7 = -1

        ; parent_pid como -1 (offset 0)
        STORE W7, W0            ; memory[pcb_v[pid].parent_pid] = -1

        ; child_pid como -1 (offset 4)
        MV W3, #4
        ADD W8, W0, W3          ; W8 = endereço do child
        STORE W7, W8            ; memory[pcb_v[pid].child] = -1

        ; prev_sibling como -1 (offset 8)
        MV W3, #8
        ADD W8, W0, W3          ; W8 = endereço do prev_sibling
        STORE W7, W8            ; memory[pcb_v[pid].prev_sibling] = -1

        ; next_sibling como -1 (offset 12)
        MV W3, #12
        ADD W8, W0, W3          ; W8 = endereço do next_sibling
        STORE W7, W8            ; memory[pcb_v[pid].next_sibling] = -1

        ; status_addr como -1 (offset 16)
        MV W3, #16
        ADD W8, W0, W3          ; W8 = endereço do status_addr
        STORE W7, W8            ; memory[pcb_v[pid].status_addr] = -1

        ; c. Inicializar registradores do processo no PCB (offsets 20 a 68)
        MV W7, #0               ; W7 = 0

        ; W0 (offset 20)
        MV W3, #20
        ADD W8, W0, W3
        STORE W7, W8

        ; W1 (offset 24)
        MV W3, #24
        ADD W8, W0, W3
        STORE W7, W8

        ; W2 (offset 28)
        MV W3, #28
        ADD W8, W0, W3
        STORE W7, W8

        ; W3 (offset 32)
        MV W3, #32
        ADD W8, W0, W3
        STORE W7, W8

        ; W4 (offset 36)
        MV W3, #36
        ADD W8, W0, W3
        STORE W7, W8

        ; W5 (offset 40)
        MV W3, #40
        ADD W8, W0, W3
        STORE W7, W8

        ; W6 (offset 44)
        MV W3, #44
        ADD W8, W0, W3
        STORE W7, W8

        ; W7 (offset 48)
        MV W3, #48
        ADD W8, W0, W3
        STORE W7, W8

        ; W8 (offset 52)
        MV W3, #52
        ADD W8, W0, W3
        STORE W7, W8

        ; W9 (offset 56)
        MV W3, #56
        ADD W8, W0, W3
        STORE W7, W8

        ; PC (offset 60) = 0
        MV W3, #60
        ADD W8, W0, W3
        STORE W7, W8

        ; SP (offset 64) = PARTITION_SIZE
        LDD W1, PARTITION_SIZE
        MV W3, #64
        ADD W8, W0, W3
        STORE W1, W8

        ; SR (offset 68) = 24 (User Mode + Habilitar Interrupções)
        MV W1, #24
        MV W3, #68
        ADD W8, W0, W3
        STORE W1, W8

        ; d. Copiar o buffer e limpar o restante da memória
        ; Obter BUFFER_START em W3 (BUFFER_START = LIMIT - BUFFER)
        LDD W3, MEMORY_SIZE     ; W3 = total memory size
        LDD W2, BUFFER          ; W2 = BUFFER size constant
        SUB W3, W3, W2          ; W3 = BUFFER_START

        ; Obter BASE do processo em W1
        MV W2, #72              ; offset de BASE
        ADD W8, W0, W2
        LOAD W1, W8             ; W1 = BASE do processo

        LDD W6, PARTITION_SIZE  ; W6 = partition size

        ; Etapa 1: Copiar exatamente o tamanho do BUFFER
        LDD W4, BUFFER          ; W4 = tamanho do BUFFER
        MV W8, #0               ; W8 = index = 0
        MV W2, #4               ; W2 = passo (4 bytes)

    input_copy_loop:
        CMP W8, W4              ; Compara index (W8) com tamanho do BUFFER (W4)
        BEQ input_copy_end      ; Se copiou o buffer inteiro, vai para a limpeza do resto

        ; Ler do buffer (BUFFER_START + index)
        ADD W7, W3, W8          ; W7 = BUFFER_START + index
        LOAD W5, W7             ; W5 = memory[BUFFER_START + index]

        ; Gravar na partição (BASE + index)
        ADD W7, W1, W8          ; W7 = BASE + index
        STORE W5, W7            ; memory[BASE + index] = W5

        ADD W8, W8, W2          ; index += 4
        JUMP input_copy_loop

    input_copy_end:
        ; Preencher o restante da partição com Zeros (W5 = 0)
        MV W5, #0               ; W5 = 0 (valor para limpar)

    input_clear_loop:
        CMP W8, W6              ; Compara index (W8) com tamanho total da partição (W6)
        BEQ input_clear_end     ; Se limpou toda a partição, termina

        ; Gravar zero na partição (BASE + index)
        ADD W7, W1, W8          ; W7 = BASE + index
        STORE W5, W7            ; memory[BASE + index] = 0

        ADD W8, W8, W2          ; index += 4
        JUMP input_clear_loop

    input_clear_end:
        JUMP schedule           ; Processo carregado, chama o escalonador

    next_input_iteration:
        MV W2, #1               ; Restaura passo = 1
        ADD W4, W4, W2          ; pid++
        JUMP input_loop

input_end:
    JUMP schedule