# ISA 
## Introdução
Para a criação da Instruction Set Architecture (ISA) referente à VM, foram utilizadas como base as ISAs do ARM e do RISC-V. A VM projetada, inicialmente, tem componentes focados em manipulação de memória.
## Definição dos registradores
### Registradores de Propósito Geral
Os registradores de propósito geral (GPRs) são registradores flexíveis, utilizados para armazenar dados temporários, operandos de operações aritméticas/lógicas, e endereços de memória. Nesta ISA, foram definidos 6 registradores de 8 bits, nomeados de W0 a W5.

- **W0, W1, W2, W3, W4 E W5 [7:0]**

### Registradores especiais
Os registradores especiais têm propósitos específicos e existem para lidar com funções essenciais para a operação da máquina.
Para o caso específico da VM, foram escolhidos 6 registrados essenciais e que existem principalmente para auxiliar na manipulação de memória.

- **Program Counter (PC)**: Armazena o endereço da próxima instrução a ser executada. É automaticamente incrementado após cada ciclo de instrução, a menos que seja modificado por uma instrução de salto.
- **Instruction Register (IR)**: Contém a instrução atual que está sendo decodificada e executada. É carregado através do FETCH.
- **Memory Data Register (MDR)**: Mantém os dados sendo transferidos para ou a partir da memória.
- **Stack Pointer (SP)**: Aponta para o topo da Stack, utilizado para gerenciar chamadas de função e armazenamento de variáveis locais.
- **Memory Address Register (MAR)**: Armazena o endereço da memória onde a operação de leitura ou escrita será realizada.
- **Status Register (SR)**: Armazena flags que indicam o resultado de operações realizadas, como o flag zero (Z) ou carry (C), utilizados para decisões de controle de fluxo.

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
- **Sintaxe**: CMP \<Reg. de entrada>, \<Reg. de entrada>
- **Exemplo**: CMP W0, W1

### Operações de controle de fluxo
**JUMP**:
- **Descrição**: Modifica o valor do registrador Program Counter (PC) para um novo endereço de memória, alterando o fluxo de execução do programa.
- **Sintaxe**: JUMP \[<Endereço/Label>]
- **Exemplo**: JUMP START_LOOP


### Operações para Load e Store com endereços
**LOAD**:
- **Descrição**: Carrega o conteúdo armazenado em um determinado endereço da memória para um registrador.
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