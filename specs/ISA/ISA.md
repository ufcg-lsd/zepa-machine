# ISA
## Introduction
The objective of this Instruction Set Architecture (ISA) is to define the registers and assembly instructions for the machine, for now mainly working with memory manipulation. For creating this document, the ISAs of ARM64 and RISC-V were used as reference.
## Definition of Registers
### General Purpose Registers
The general-purpose registers are flexible and can be used in different situations, as temporarily storing values, used to assist in operations, and storing memory addresses.

Initially, this machine has 6 registers, which are named W0 to W5, each of them capable of storing 32 bits.

- **W0, W1, W2, W3, W4, and W5 [31:0]**

### Special Registers
The special registers have specific purposes and exist to handle essential functions for the machine's operation.
For the specific case of this machine, six registers were defined, mainly to assist in memory manipulation, all storing values up to 32 bits.

- **Program Counter (PC) [31:0]**: Stores the address of the next instruction to be executed. Is automaticaly incremented after every instruction cycle, unless modified by a JUMP instruction.
- **Instruction Register (IR) [31:0]**: Contains the current instruction being decoded and executed.
- **Memory Data Register (MDR) [31:0]**: Holds the data being transferred from or to memory.
- **Stack Pointer (SP) [31:0]**: Points to the top of the stack, used to manage function calls and local variable storage.
- **Memory Address Register (MAR) [31:0]**: Stores the memory address where reading or writing operations will be executed.
- **Status Register (SR) [31:0]**: Stores flags that indicate the result of test operations executed. The first bits are reserved for the G, L and Z flags, and the last ones are flexible.

## Encoding
For this machine, the word size, instruction size, and register size were defined to be 32 bits.
### R-Type (Register type) format

| opcode  | rd  | rs1  | rs2  | funct5  | funct6  |
|-----------------|-------------|--------------|--------------|----------------|-----------------|
| 6 bits          | 5 bits      | 5 bits       | 5 bits       | 5 bits         | 6 bits          |

- **Description**: Used for instructions that only involve registers.
- **Fields**:
    - **opcode**: Operation code, such as ADD, SUB..
    - **rd**: Destination register.
    - **rs1**: Source register 1.
    - **rs2**: Source register 2.
    - **funct5** and **funct6**: Opcode extensions.


### I-Type (Immediate type) format


| opcode  | rs1/rd  | constant/address  | funct5  
|-----------------|-------------|--------------|--------------|
| 6 bits          | 5 bits      | 16 bits       | 5 bits       |

- **Description**: Used for instructions that also involve a constant or address.
- **Fields**:
    - **opcode**: Operation code, such as LOAD, STORE, etc..
    - **rs1/rd**: Destination or source register.
    - **constant/address**: Constant or address value
    - **funct5**: Opcode extension.

## Definition of Instructions

### Attribution
**MV**:
- **Description**: Moves a constant to a specific register.
- **Syntax**: MV \<Destination Reg.> \#\<Constant>
- **Example**: MV W1 #5
- **Format**: I-Type
- **Opcode (decimal)**: 12

### Arithmetic and Logical Operations
**ADD**:
- **Description**: Adds the values of two registers and saves the result in a third one.
- **Syntax**: ADD \<Destination Reg.>, \<Input Reg.>, \<Input Reg.>
- **Example**: ADD W0, W1, W0
- **Format**: R-Type
- **Opcode (decimal)**: 13

**SUB**:
- **Description**: Subtracts the value of one register from another and stores the result in a third one.
- **Syntax**: SUB \<Destination Reg.>, \<Input Reg.>, \<Input Reg.>
- **Example**: SUB W0, W1, W0
- **Format**: R-Type
- **Opcode (decimal)**: 14

### Test Instructions
**CMP**:
- **Description**: Compares the values of two registers and stores the flag in the SR.
- **Syntax**: CMP \<Input Reg.>, \<Input Reg.>
- **Example**: CMP W0, W1
- **Format**: R-Type
- **Opcode (decimal)**: 15

#### Z, L and G Test Flags
When a test instruction, such as CMP, is executed, the SR register is updated, and its value can be used by other instructions to change the program's flow. Each flag is represented by a bit, and the flag being set indicates that the bit value is 1.
Considering the SR register [31:0], the following flags can be set in their respective bits:
- **Z**, bit [0] - Set if the result of the last comparison is 0, meaning the operands are equal.
- **L**, bit [1] - Set if the result of the last comparison indicates that the first operand is less than the second.
- **G**, bit [2] - Set if the result of the last comparison indicates that the first operand is greater than the second.

These flags can be used by instructions to make decisions that can change the program flow.

### Control Flow Operations
**JUMP**:
- **Description**: Change the value of the Program Counter (PC) register, updating the program's execution flow.
- **Syntax**: JUMP \[<Address>]
- **Example**: JUMP 0x14
- **Format**: I-Type
- **Opcode (decimal)**: 16

### Load and Store Operations with Addresses
**LOAD**:
- **Description**: Loads the content stored at a specific memory address into a specific register.
- **Syntax**: LOAD \<Destination Reg.>, [\<Address>]
- **Example**: LOAD W0, 0x68DB00AD
- **Format**: I-Type
- **Opcode (decimal)**: 17

**STORE**:
- **Description**: Stores the value of a register to memory.
- **Syntax**: STORE \<Source Reg.>, [\<Address>]
- **Example**: STORE W1, 0x68DB00AD
- **Format**: I-Type
- **Opcode (decimal)**: 18

### Processor Execution Cycle
**FETCH**
- **Description**: Get the next instruction from memory using the address stored in the Program Counter (PC) and load it into the Instruction Register (IR).
- **Syntax and Example**: FETCH
- **Format**: I-Type
- **Opcode (decimal)**: 19

### Zepa Machine Instruction Encoding Table

| **Instruction** | **Format** | **opcode** | **rd** | **rs1** | **rs2** | **funct5** | **funct6** |
|-----------------|------------|------------|---------|---------|------------|--------|------------|
| **ADD**         | R-Type         | 001101    | reg     | reg     | reg        | 00000    | 000000    |
| **SUB**         | R-Type          | 001110    | reg     | reg     | reg        | 00000    | 000000    |
| **CMP**         | R-Type          | 001111    | 00000     | reg     | reg        | 00000    | 000000    |



| **Instruction** | **Format** | **opcode** | **rs1/rd** | **immediate** | **funct5** |
|-----------------|------------|---------------|---------|------------|--------|
| **MV**          | I-Type          | 001100      | reg     | 16bit constant         | 00000    |
| **JUMP**        | I-Type          | 010000      | 00000     | 16bit address        | 00000    |
| **LOAD**        | I-Type          | 010001       | reg     | 16bit address        | 00000    |
| **STORE**        | I-Type          | 010010       | reg     | 16bit address        | 00000    |
| **FETCH**        | I-Type          | 010011       | 00000     | 0000000000000000        | 00000    |


## References
- [Bit by Bit: Processadores Cortex-M0+ - Arquitetura do Conjunto de Instruções (ISA)](https://bit-by-bit.gitbook.io/embedded-systems/processadores-cortex-m0+/arquitetura-do-conjunto-de-instrucoes-isa)
- [ARM Developer: ARMv8-M Architecture Reference Manual](https://developer.arm.com/documentation/102374/0101)
- [RISC-V International: RISC-V Specifications](https://riscv.org/technical/specifications/)
- [Go Programming Language Specification](https://go.dev/ref/spec)
- [GeeksforGeeks: Essential Registers for Instruction Execution](https://www.geeksforgeeks.org/essential-registers-for-instruction-execution/)
###

