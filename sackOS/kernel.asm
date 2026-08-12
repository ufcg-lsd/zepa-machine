setup:
    MV W0 #2048  
    MV W1 #4096
    MUL W1, W0, W1   ; 8MB of full kernel memory
    STRD W1 #0x1008 ; kernel_max_memory
    STRD W1, #0x1020 ; initilizes the kernel stack pointer as the kernel_max_memory
    LDD W2 #0x100C ; buffer_size
    STRD LIMIT #0x1010 ; memory_size
    ADD W3 W1 W2 ; W3 -> KERNEL_MAX_MEMORY + BUFFER

    SUB W0 LIMIT W3 ; W0 -> USER_MEMORY = limit - (KERNEL_MAX_MEMORY + BUFFER)
    LDD W2 #0x1000 ; partition_size

    UDIV W3 W0 W2 ; W3 -> NUM_PARTITIONS = user_memory / PARTITION_SIZE
    STRD W3, #0x1014 ; partition_number

    MV W0, #-1
    STRD W0, #0x1018 ; running_pid = -1

    MV W4 #0x102C ; pcb_v

    MV W8 #72
    ADD W4 W4 W8 ; gets W4 to pcb_v[0].BASE address

   ; W1 will acumulate PARTITION_SIZE * i through the loop

    MV W6 #0 ; W6 -> pid = 0
    ; for (int pid = 0, pid < NUM_PARTITIONS; pid++)

    CMP W6 W3 
    BEQ setup_registers ; loop conditions

        STORE W1 W4 ; pcb_v[pid].BASE = KERNEL_MAX_MEMORY + pid * PARTITION_SIZE
        
        MV W8 #4
        ADD W4 W4 W8 ; W4 = pcb_v[pid].LIMIT address

        ADD W7 W1 W2 ; W7 = BASE + PARTITION_SIZE

        STORE W7 W4 ; pcb_v[pid].LIMIT = BASE + PARTITION_SIZE

        MV W8 #1
        ADD W6 W6 W8 ; pid += 1 


        ; setup variables to next Iteration

        MV W8 #80
        ADD W4 W4 W8 ; W4 = pcb_v[pid+1].BASE address
        ADD W1 W1 W2 ; W1 += PARTITION_SIZE

        JUMP #-12


        setup_registers:
            LDD SP #0x1020 ; kernel_stack_pointer
            MV ESA #0x94 ; the exception_supervisor initial address
            MV EPC #0x90 ; the infinite loop below
            MV ESR #16 ; enable interruptions
            
            MRET ; go to infinite loop, waiting for program inputs

JUMP #0

exception_supervisor:
    STRD W0 #0x1024 ; scratch_space_0
    STRD W1 #0x1028 ; scratch_space_1

    MV W0 #0
    CMP ECR W0 ; if ECR = 0 (clock interruption)
    BEQ clock_int

    MV W0 #-1
    LDD W1 #0x1018 ; running_pid

    CMP W0 W1
    BEQ jumpToHandler

    MV W0 #84 ; pcb_size
    MUL W1 W1 W0 ; W1 = RUNNING_PID * pcb_size

    MV W0 #0x102C ; pcb_v
    ADD W0 W0 W1 ; W0 = pcb[RUNNING_PID] address

    MV W1 #28
    ADD W0 W0 W1 ;  W0 = pcb[RUNNING_PID].w2 address

    MV W1 #4 ; uses W1 to jump to next register address
    ; save registers

    STORE W2 W0
    ADD W0 W1 W0

    STORE W3 W0
    ADD W0 W1 W0

    STORE W4 W0
    ADD W0 W1 W0

    STORE W5 W0
    ADD W0 W1 W0

    STORE W6 W0
    ADD W0 W1 W0

    STORE W7 W0
    ADD W0 W1 W0

    STORE W8 W0
    ADD W0 W1 W0

    STORE W9 W0
    ADD W0 W1 W0

    STORE EPC W0
    ADD W0 W1 W0

    STORE SP W0
    ADD W0 W1 W0

    STORE ESR W0
  
    MV W3 #48
    SUB W0 W0 W3 ; w0 points to pcb[RUNNING_PID].w0

    LDD W2 #0x1024 ; scratch_space_0

    STORE W2 W0

    MV W8 #4
    ADD W0 W0 W8 ; w0 points to pcb[RUNNING_PID].w1

    LDD W2 #0x1028 ; scratch_space_1
    STORE W2 W0

    jumpToHandler:
        MV W0 #1
        MV W1 #1 

        CMP ECR W0
        BEQ input_int

        ADD W0 W0 W1

        CMP ECR W0
        BEQ kill_int

        ADD W0 W0 W1 
        
        CMP ECR W0
        BEQ syscall_int

        ADD W0 W0 W1 
        
        CMP ECR W0
        BEQ fault_int    


; INTERRUPTIONS


clock_int:
  LDD W0, #0x101C    ; w0 = clock_interrupt_count

  MV W1, #1
  ADD W0, W0, W1                    ; clock_interrupt_count += 1

  LDD W1, #0x1004 ; time_slice 
  CMP W0, W1
  BEQ clock_reset                   ; if clock_interrupt_count == TIME_SLICE, reset and check running

    STRD W0, #0x101C ; clock_interrupt_count
  clock_return:
    LDD W0, #0x1024 ; scratch_space_0
    LDD W1, #0x1028 ; scratch_space_1
    MRET

  clock_reset:
    MV W0, #0
    STRD W0, #0x101C ; clock_interrupt_count = 0

  LDD W0, #0x1018 ; running_pid
  MV W1, #-1
  CMP W0, W1
  BEQ clock_return                   ; if running_pid == -1, clock_return

  ; saving registers
  LDD W1, #0x1018 ; running_pid    
  MV W0, #84                      ; pcb_size
  MUL W1, W0, W1                  ; W1 = running_pid * pcb_size
  
  MV W0, #0x102C ; pcb_v
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

  LDD W3, #0x1028 ; scratch_space_1
  SUB W1, W1, W0
  STORE W3, W1                    ; saving W1

  LDD W3, #0x1024 ; scratch_space_0
  SUB W1, W1, W0
  STORE W3, W1                    ; saving W0

  JUMP schedule


input_int:
    ; Obter BUFFER_START em W3 (BUFFER_START = LIMIT - BUFFER)
    LDD W3, #0x1010     ; W3 = total memory size
    LDD W2, #0x100C ; W2 = BUFFER size constant
    SUB W3, W3, W2          ; W3 = BUFFER_START
    LOAD W3, W3 ; W3 = buffer_input_size, set at the start of the buffer

    LDD W4, #0x1000 ; W4 = partition_size
    
    CMP W3, W4
    BGT input_end ; if the buffer_input_size is greater than the partition size, just go back to the scheduler

    MV W4, #0                   ; W4 = pid = 0 (contador do loop)
    LDD W5, #0x1014    ; W5 = PARTITION_NUMBER (limite do loop)
    MV W2, #1                   ; W2 = 1 (passo do loop)

    input_loop:
        CMP W4, W5              ; Compara pid (W4) com PARTITION_NUMBER (W5)
        BEQ input_end           ; Se pid == limite, sai do loop

        MV W0, #0x102C          ; W0 = endereço base do vetor pcb_v
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
        LDD W1, #0x1000 ; partition_size
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
        LDD W3, #0x1010     ; W3 = total memory size
        LDD W2, #0x100C ; W2 = BUFFER size constant
        SUB W3, W3, W2          ; W3 = BUFFER_START

        ; Obter BASE do processo em W1
        MV W2, #72              ; offset de BASE
        ADD W8, W0, W2
        LOAD W1, W8             ; W1 = BASE do processo

        LDD W6, #0x1000 ; W6 = partition size

        ; Etapa 1: Copiar exatamente o tamanho do BUFFER
        MV W8, #0               ; W8 = index = 0
        MV W2, #4               ; W2 = passo (4 bytes)
        LOAD W4, W3             ; W4 = buffer_input_size
        ADD W3, W3, W2          ; Começa após o input_size

    input_copy_loop:
        CMP W8, W4              ; Compara index (W8) com tamanho do BUFFER (W4)
        BEQ input_copy_end      ; Se copiou o buffer inteiro, vai para a limpeza do resto
        BGT input_copy_end 
        ; Ler do buffer (BUFFER_START + index)
        ADD W7, W3, W8          ; W7 = BUFFER_START + index
        LOAD W5, W7             ; W5 = memory[BUFFER_START + index]

        ; Gravar na partição (BASE + index)
        ;   W1 = partition_size                         

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
        BGT input_clear_end
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


kill_int:

    LDD W9, #0x1010 ; memory_size
    LDD W7, #0x100C ; buffer_size
    SUB W9, W9, W7  ; initial buffer addr
    LOAD W9, W9     ; pid of kill_int

    LDD W1, #0x1014 ; partition_number

    CMP W9, W1
    BGT not_valid
    BEQ not_valid

    MV  W2, #84
    MUL W3, W9, W2
    MV W4, #0x102C ; pcb_v
    ADD W3, W3, W4

    MV  W5, #80
    ADD W5, W3, W5
    LDB W6, W5

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
    STORE W6, W5

    JUMP kill

    not_valid:
        JUMP schedule


syscall_int:
    MV W1, #0
    CMP W9, W1
    BEQ fork

    MV W1, #1
    CMP W9, W1
    BEQ wait

    MV W1, #2
    CMP W9, W1
    BEQ exit

    MV W1, #3
    CMP W9, W1
    BEQ getPID

    MV W1, #4
    CMP W9, W1
    BEQ rele

    MV W0, #0x102C              ; set pcb_v adress
    LDD W1, #0x1018             ; current running pid
    MV W2, #84                  ; size of pcb
    MUL W1, W1, W2              ; running pid * 84
    ADD W0, W0, W1              ; pcb_v + (running pid * 84)

    MV W3, #56                  ; offset of pc in pcb to get W9
    ADD W0, W0, W3              ; pcb_v + (running pid * 84) + 56

    MV W3, #-1                  ; value -1
    STORE W3, W0                ; pcb[running_pid].w9 = -1

    JUMP schedule


fault_int:
  MV W0, #0x102C       ; pcb_v initial address
  LDD W9, #0x1018      ; running_pid
  MV W1, #84           ; pcb_size

  MUL W1, W9, W1       ; W1 = running_pid * pcb_size
  ADD W0, W0, W1       ; W0 = pcb_v[running_pid] initial address

  MV W1, #56           ; pcb W9 offset
  ADD W0, W0, W1       ; W0 = pcb_v[running_pid].W9 initial address

  MV W1, #1
  STORE W1, W0         ; pcb_v[running_pid].W9 = 1

  JUMP kill            ; kill(running_pid)



; SYSCALLS

fork:
    MV W4, #0                   ; W4 = pid = 0 (contador do loop)
    LDD W5, #0x1014    ; W5 = PARTITION_NUMBER (limite do loop)
    MV W2, #1                   ; W2 = 1 (passo do loop)

    fork_loop:
        CMP W4, W5              ; Compara pid (W4) com PARTITION_NUMBER (W5)
        BEQ fork_end            ; Se pid == limite, sai do loop

        MV W0, #0x102C          ; W0 = endereço base do vetor pcb_v
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
        LDD W7, #0x1018     ; W7 = running_pid
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
        LDD W1, #0x1018         ; W1 = running_pid
        MV W3, #84              ; W3 = 84 (tamanho de cada PCB)
        MUL W1, W1, W3          ; W1 = running_pid * 84
        MV W7, #0x102C          ; W7 = base pcb_v
        ADD W7, W7, W1          ; W7 = endereço base do PCB pai

        ; e. Retorno do fork no pai: pcb_v[running_pid].w9 = pid (W4)
        MV W3, #56              ; offset de W9 no PCB
        ADD W8, W7, W3          ; W8 = endereço de pcb_v[running_pid].w9
        STORE W4, W8            ; pcb_v[running_pid].w9 = pid

        ; f. Retorno do fork no filho: pcb_v[pid].w9 = -2
        MV W3, #56              ; offset de W9 no PCB
        ADD W8, W0, W3          ; W8 = endereço de pcb_v[pid].w9
        MV W1, #-2              ; W1 = -2
        STORE W1, W8            ; pcb_v[pid].w9 = -2

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
            MV W3, #0x102C          ; pcb_v
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

        LDD W6, #0x1000 ; W6 = size = PARTITION_SIZE

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
        MV W0, #0x102C              ; set pcb_v adress
        LDD W1, #0x1018             ; current running pid
        MV W2, #84                  ; size of pcb
        MUL W1, W1, W2              ; running pid * 84
        ADD W0, W0, W1              ; pcb_v + (running pid * 84)

        MV W3, #56                  ; offset of pc in pcb to get W9
        ADD W0, W0, W3              ; pcb_v + (running pid * 84) + 56

        MV W3, #-1                  ; value -1
        STORE W3, W0                ; pcb[running_pid].w9 = -1
        JUMP schedule

wait:
    MV W0 #0x102C ; w0 points to pcb_v[0] first byte
    MV W8 #52 
    ADD W0 W0 W8 ; w0 points to pcb_v[0].w8
    MV W8 #84 ; bytes size of each pcb

    LDD W1 #0x1018 ; running_pid
    MUL W1 W1 W8 ; W1 = RUNNING_PID * 84 bytes
    ADD W0 W0 W1 ; w0 points to pcb_v[RUNNING_PID].w8
    
    LOAD W9 W0 ; w9 = status_addr

    MV W8 #20
    ADD W0 W8 ; w0 points to pcb_v[running_pid].BASE

    LOAD W2 W0 ; w2 = pcb_v[RUNNING_PID].BASE
    
    MV W8 #4
    ADD W0 W0 W8 ; w0 points to pcb_v[RUNNING_PID].LIMIT
    
    LOAD W3 W0 ; w3 = pcb_v[RUNNING_PID].LIMIT

    ADD W7 W2 W9 ; w7 = pcb_v[running_pid].BASE + status_addr
    ADD W2 W2 W9 ; w2 = pcb_v[running_pid].BASE + status_addr
    
    MV W8 #4
    ADD W2 W2 W8 ; w2 = pcb_v[running_pid].BASE + status_addr + 4


    ; if  pcb_v[running_pid].BASE + status_addr + 4 > pcb_v[RUNNING_PID].LIMIT
    ; fault_int()


    CMP W2 W3
    BGT fault_int 


    MV W8 #72
    SUB W2 W0 W8 ; w2 points to pcb_v[RUNNING_PID].child 
    LOAD W6 W2 ; w6 = pcb_v[RUNNING_PID].child 

    MV W8 #-1 
    CMP W6 W8 ; if pcb_v[RUNNING_PID].child  == -1:
    BEQ #2

    JUMP #6

        MV W8 #20
        SUB W3 W0 W8 ; w3 points to pcb_v[RUNNING_PID].w9
        MV W8 #-1
        STORE W8 W3
        JUMP schedule

    LOAD W3 W2 ; w3 = curr_child = pcb_v[running_pid].childPID
    MV W0 #0x102C ;  w0 points to pcb_v[0] first byte

    waitLoop:
        ;calculate pcb_v[curr_child]
        ; w3 = curr_child

        MV W8 #84
        MUL W1 W3 W8

        ADD W1 W0 W1 ;W1 points to pcb_v[curr_child] first byte

        MV W8 #80
        ADD W1 W1 W8 ; w1 points to pcb_v[curr_child].flags
        LDB W6 W1 ; w6 = pcb_v[curr_child].flags

        MV W8 #2 ;0b00000000000000000000000000000010
        OR W8 W8 W6
        CMP W6 W8
        BEQ curr_child_is_zombie 

            ;if !pcb_v[curr_child].is_zombie:
            ; curr_child = pcb_v[curr_child].next_sibling

            MV W8 #68
            SUB W3 W1 W8 ; w3 points to pcb_v[curr_child].next_siblingPID
            LOAD W6 W3 ; w6 = pcb_v[curr_child].next_siblingPID

            MV W8 #-1
            CMP W6 W8 ; if pcb_v[curr_child].next_siblingPID == -1
            BEQ setWaitingStatus

            LOAD W3 W3 ; w3 = curr_child = pcb_v[curr_child].next_siblingPID
            JUMP waitLoop ; goes to next Iteration

        ; else (pcb_v[curr_child].is_zombie)

        curr_child_is_zombie:
            MV W8  #-2 ;0b11111111111111111111111111111110
            LDB W6 W1
            AND W8 W8 W6 
            STRB W8 W1 ; sets pcb_v[curr_child].is_mapped = 0


            
            ; w0 points to pcb_v[0] first byte
            ; w1 points to pcb_v[curr_child].flags
            ; w2 points to pcb_v[running_pid].child
            ; w3 = curr_child
            ; w7 = pcb_v[running_pid].BASE + status_addr

            MV W8 #52
            ADD W4 W2 W8 ; w4 points to pcb_v[running_pid].w9
            STORE W3 W4 ; pcb_v[running_pid].w9 = curr_child

            MV W8 #24 
            SUB W1 W1 W8 ; w1 points to pcb_v[curr_child].w9
            LOAD W6 W1 ; w6 = pcb_v[curr_child].w9

            STORE W6 W7 ; memory[pcb_v[running_pid].BASE+status_addr] = pcb_v[curr_child].w9

            MV W8 #44
            SUB W1 W1 W8 ; w1 points to pcb_v[curr_child].next_sibling

            LOAD W6 W2 ; w6 = pcb_v[running_pid].child
            CMP W6 W3 ; if curr_child == pcb_v[running_pid].child
            BEQ #2

            JUMP curr_child_is_not_first_children

            LOAD W6 W1 ; w6 = pcb_v[curr_child].next_sibling
            STORE W6 W2 ; pcb_v[running_pid].child = pcb_v[curr_child].next_sibling

            JUMP link_next_sibling
            
        

        curr_child_is_not_first_children:
            MV W8 #4
            SUB W4 W1 W8 ; w4 points to pcb_v[curr_child].prev_sibling
            LOAD W7 W4   ; w7 = pcb_v[curr_child].prev_sibling

            MV W6 #84 ; 84 bytes each pcb
            MUL W5 W7 W6 ; 
            ADD W5 W0 W5 ; w5 points to pcb_v[pcb_v[running_pid].prev_sibling] first byte

            MV W8 #12
            ADD W5 W5 W8 ; w5 points to pcb_v[pcb_v[running_pid].prev_sibling].next_sibling

            LOAD W7 W1 ; w7 = pcb_v[curr_child].next_sibling
            STORE W7 W5 ; pcb_v[pcb_v[running_pid].prev_sibling].next_sibling = pcb_v[curr_child].next_sibling

        
            ; else
        link_next_sibling:


            ; w1 points to pcb_v[curr_child].next_sibling
            LOAD W7 W1 ; w7 = pcb_v[curr_child].next_sibling
            MV W8 #-1
            CMP W7 W8 ; if pcb_v[curr_child].next_sibling == -1 
            BEQ schedule

            MV W6 #84 ; 84 bytes each pcb
            MUL W5 W7 W6 
            ADD W5 W0 W5 ; w5 points to pcb_v[pcb_v[running_pid].next_sibling] first byte

            MV W8 #8
            ADD W5 W5 W8 ; w5 points to pcb_v[pcb_v[running_pid].next_sibling].prev_sibling

            MV W8 #4
            SUB W1 W1 W8 ; w1 = pcb_v[curr_child].prev_sibling
            LOAD W8 W1
            STORE W8 W5 ; pcb_v[pcb_v[running_pid].next_sibling].prev_sibling = pcb_v[curr_child].prev_sibling

            JUMP schedule

    setWaitingStatus:
        ; w2 = pcb_v[running_pid].child

        MV W8 #12
        ADD W2 W2 W8 ; w2 = pcb_v[running_pid].status_add
        STORE W9 W2 ; pcb_v[running_pid].status_addr = status_addr

        MV W8 #64
        ADD W2 W2 W8 ; w2 points to pcb_v[running_pid].flags

        LDB W4 W2 ; w4 = pcb_v[running_pid].flags

        ; pcb_v[running_pid].scheduler_state = blocked
        ; pcb_v[running_pid].is_waiting = 1
        ; same to set flags byte to  0bXXX101XX

        
        ; set 1's 
        MV W8 #20 ; 0b00010100
        OR W4 W4 W8

        ; set 0
        MV W8 #-9 ; 0b11110111
        AND W4 W4 W8

        STRB W4 W2 ;

        JUMP schedule


exit:
	LDD W9, #0x1018        ; W9 = running_pid
	MV W1, #84               ; W1 = pcb_size
	MUL W0, W9, W1           ; W0 = running_pid * pcb_size

	MV W1, #0x102C           ; W1 = pcb_v initial address
	ADD W0, W0, W1           ; W0 = pcb_v[running_pid] initial address

	MV W1, #52
	ADD W0, W0, W1           ; W0 = pcb_v[running_pid].W8 address
	LOAD W2, W0              ; W2 = pcb_v[running_pid].W8 (status_code)

	MV W1, #4
	ADD W0, W0, W1           ; W0 = pcb_v[running_pid].W9 address
	STORE W2, W0             ; pcb_v[running_pid].W9 = pcb_v[running_pid].W8

	JUMP kill


getPID:
    LDD W0 #0x1018 ; running_pid
    MV W1 #0x102C ; pcb_v

    MV W8 #84
    MUL W2 W0 W8 ; RUNNING_PID * 84 get the offset of bytes to acess pcb[RUNNING_PID]
    
    ADD W1 W1 W2 ; stores on W1 the first address of pcb[RUNNING_PID]

    MV W8 #56
    ADD W1 W1 W8 ; W1 = pcb[RUNNING_PID].w9 address bytes

    STORE W0 W1 ; stores RUNNING_PID on pcb[RUNNING_PID].w9

    JUMP schedule


rele:
    JUMP schedule


; HELPERS


kill:
  MV W0, #0x102C          ; W0 = pcb_v initial address
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

      MV W4, #24
      SUB W3, W3, W4      ; W3 = pcb_v[parent].W9 address
      STORE W9, W3        ; pcb_v[parent].W9 = pid

      MV W4, #16
      ADD W3, W3, W4      ; W3 = pcb_v[parent].BASE address
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

        MV W5, #44
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



schedule:

    ; clock_interrupt_count = 0
    ; Every process selected by the scheduler starts with a new time slice

    MV W0, #0
    STRD W0, #0x101C ; clock_interrupt_count

    ; valida running_pid

    LDD W9, #0x1018 ; running_pid

    MV W0, #-1
    CMP W9, W0
    BEQ schedule_reset_running_pid

    ; valida running_pid

    LDD W0, #0x1014 ; partition_number

    CMP W9, W0
    BEQ schedule_reset_running_pid
    BGT schedule_reset_running_pid

    JUMP schedule_check_current_process


    schedule_reset_running_pid:

        MV W9, #0


    schedule_check_current_process:

        ; verifica se processo atual esta running
        ; W1 = pcb_v[running_pid]

        MV W0, #84
        MUL W1, W9, W0

        MV W0, #0x102C ; pcb_v
        ADD W1, W0, W1

        ; W2 = pcb_v[running_pid].flags

        MV W0, #80
        ADD W0, W1, W0

        LDB W2, W0

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

        STRB W2, W0


    schedule_calculate_limit_pid:

        ; limit_pid = running_pid + 1

        MV W0, #1
        ADD W3, W9, W0

        ; wrap limit_pid se necessario

        LDD W0, #0x1014 ; partition_number

        CMP W3, W0
        BEQ schedule_limit_pid_zero

        JUMP schedule_limit_pid_ready


    schedule_limit_pid_zero:

        MV W3, #0


    schedule_limit_pid_ready:

        ; curr_pid = limit_pid

        MV W4, #0
        ADD W4, W4, W3

    schedule_search_loop:

        ; busca processo ready
        ; W1 = &pcb_v[curr_pid]

        MV W0, #84
        MUL W1, W4, W0

        MV W0, #0x102C ; pcb_v
        ADD W1, W0, W1

        ; W2 = pcb_v[curr_pid].flags

        MV W0, #80
        ADD W0, W1, W0

        LDB W2, W0

        ; esta mapeado?

        MV W5, #0x01
        AND W5, W2, W5

        MV W0, #0x01
        CMP W5, W0
        BEQ schedule_check_zombie

        JUMP schedule_next_pid


    schedule_check_zombie:

        ; eh zombie?

        MV W5, #0x02
        AND W5, W2, W5

        MV W0, #0
        CMP W5, W0
        BEQ schedule_check_ready

        JUMP schedule_next_pid


    schedule_check_ready:

        ; esta ready?

        MV W5, #0x18
        AND W5, W2, W5

        MV W0, #0x08
        CMP W5, W0
        BEQ schedule_process_found

        JUMP schedule_next_pid


    schedule_process_found:

        ; processo encontrado, seta running e restaura

        MV W5, #-25
        AND W2, W2, W5

        MV W0, #80
        ADD W0, W1, W0

        STRB W2, W0

        ; W9 = curr_pid

        MV W9, #0
        ADD W9, W9, W4
        STRD W9, #0x1018 ; running_pid = curr_pid

        ; W6 = &pcb_v[running_pid]

        MV W6, #0
        ADD W6, W6, W1

        JUMP schedule_restore_context


    schedule_next_pid:

        ; curr_pid++

        MV W0, #1
        ADD W4, W4, W0

        ; wrap curr_pid se necessario

        LDD W0, #0x1014 ; partition_number

        CMP W4, W0
        BEQ schedule_wrap_pid

        JUMP schedule_check_search_end


    schedule_wrap_pid:

        MV W4, #0


    schedule_check_search_end:

        ; while curr_pid != limit_pid

        CMP W4, W3
        BEQ schedule_no_ready_process

        JUMP schedule_search_loop


    ; precisa de futuro review a partir daqui
    schedule_no_ready_process:

        ; nenhum pronto, idle loop

        MV W9, #-1
        STRD W9, #0x1018 ; running_pid = -1

        MV EPC, #0x90 ; infinite_loop

        MV ESR, #16

        MRET


    ; W6 = endereco inicial do PCB selecionado
    ; W9 = running_pid (nao restaurado do PCB)
    schedule_restore_context:

        ; carrega registros do PCB
        ; W0 (20) -> scratch

        MV W0, #20
        ADD W8, W6, W0

        LOAD W0, W8

        STRD W0, #0x1024 ; scratch_space_0

        ; PC (60)

        MV W0, #60
        ADD W8, W6, W0

        LOAD EPC, W8

        ; SP (64)

        MV W0, #64
        ADD W8, W6, W0

        LOAD SP, W8

        ; SR (68)

        MV W0, #68
        ADD W8, W6, W0

        LOAD ESR, W8

        ; BASE (72)

        MV W0, #72
        ADD W8, W6, W0

        LOAD BASE, W8

        ; LIMIT (76)

        MV W0, #76
        ADD W8, W6, W0

        LOAD LIMIT, W8

        ; W1 (24)

        MV W0, #24
        ADD W8, W6, W0

        LOAD W1, W8

        ; W2 (28)

        MV W0, #28
        ADD W8, W6, W0

        LOAD W2, W8

        ; W3 (32)

        MV W0, #32
        ADD W8, W6, W0

        LOAD W3, W8

        ; W4 (36)

        MV W0, #36
        ADD W8, W6, W0

        LOAD W4, W8

        ; W5 (40)

        MV W0, #40
        ADD W8, W6, W0

        LOAD W5, W8

        ; W7 (48) -- restaura antes de W6 (W6 ainda tem endereco do PCB)

        MV W0, #48
        ADD W8, W6, W0

        LOAD W7, W8

        ; W9 (56) -- restaura antes de W8 (W8 ainda tem endereco do registrador)

        MV W0, #56
        ADD W8, W6, W0

        LOAD W9, W8

        ; W8 (52) -- usa W0 como temporario para nao perder o endereco

        MV W0, #52
        ADD W0, W6, W0

        LOAD W8, W0

        ; W6 (44) -- restaurado por ultimo entre W1-W9

        MV W0, #44
        ADD W0, W6, W0

        LOAD W6, W0

        ; W0 <- scratch (ultimo registrador antes do MRET)

        LDD W0, 0x1024 ; scratch_space_0

        MRET