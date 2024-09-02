# ISA 
## Introdução
Para a criação da Instruction Set Architecture (ISA) referente à VM, foram utilizadas como base as ISAs do ARM e do RISC-V.
## Definição dos registradores
### Registradores de propósito geral
Os registradores de propósito geral tem como objetivo serem mais flexíveis, e podem ser utilizados em diversas situações, como guardar valores temporariamente, operações e endereços de memória. 

Foram escolhidos 6 registradores de propósito geral, nomeados de W0 até W5, cada um realizando operações de até 8 bits.

- **W0, W1, W2, W3, W4 E W5 [7:0]**

### Registradores especiais
Os registradores especiais têm propósitos específicos e existem para lidar com funções essenciais para a operação da máquina.
Para o caso específico da VM, foram escolhidos 6 registrados essenciais e que existem principalmente para auxiliar na manipulação de memória.

- **Program Counter (PC)**: Armazena o endereço da próxima instrução a ser executada.
- **Instruction Register (IR)**: Contém a instrução atual que está sendo decodificada e executada.
- **Memory Data Register (MDR)**: Mantém os dados sendo transferidos para ou a partir da memória.
- **Stack Pointer (SP)**: Aponta para o topo da pilha, utilizado para gerenciar chamadas de função e armazenamento de variáveis locais.
- **Memory Address Register (MAR)**: Armazena o endereço da memória onde a operação de leitura ou escrita será realizada.
- **Status Register (SR)**: Armazena flags que indicam o resultado de operações realizadas.

## Definição das instruções

### Operações aritméticas e lógicas
**ADD**: 
- **Descrição**: Soma os valores de dois registradores e salva em um terceiro.
- **Sintaxe**: ADD \<Reg. de destino>, \<Reg. de entrada> , \<Reg. de entrada>
- **Exemplo**: ADD W0, W1, W0

**SUB**: 
- **Descrição**: Subtrai o valor de um registrador do valor de outro e armazena em um terceiro.
- **Sintaxe**: SUB \<Reg. de destino>, \<Reg. de entrada> , \<Reg. de entrada>
- **Exemplo**: SUB W0, W1, W0

**CMP**:
- **Descrição**: Compara os valores de dois registradores e armazena a flag no registrador Status Register
- **Sintaxe**: CMP \<Reg. de entrada> , \<Reg. de entrada>
- **Exemplo**: CMP W0, W1

### Operações para Load e Store com endereços
**LOAD**:
- **Descrição**: Carrega o contúdo armazenado em um determinado endereço da memória para um registrador.
- **Sintaxe**: LOAD \<Reg. de destino>, \[<Endereço>]
- **Exemplo**: LOAD W0, 0x68DB00AD

**STORE**:
- **Descrição**: Armazena o valor de um registrador na memória.
- **Sintaxe**: STORE \<Reg. de origem>, \[<Endereço>]
- **Exemplo**: STORE W1, 0x68DB00AD

### Ciclo de execução do processador
**FETCH**
- **Descrição**: Recupera a próxima instrução da memória usando o endereço armazenado no Program Counter (PC) e carrega essa instrução no Instruction Register (IR).
- **Sintaxe e exemplo**: FETCH
### 