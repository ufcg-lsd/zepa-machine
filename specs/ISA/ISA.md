# ISA
## Introduction
The objective of this Instruction Set Architecture (ISA) is to define the registers and assembly instructions for the machine, for now mainly working with memory manipulation. For creating this document, the ISAs of ARM64 and RISC-V were used as reference.
## Definition of Registers
### General Purpose Registers
The general-purpose registers are flexible and can be used in different situations, as temporarily storing values, used to assist in operations, and storing memory addresses.

Initially, this machine has 10 registers, which are named W0 to W9, each of them capable of storing 32 bits.

- **W0, W1, W2, W3, W4, W5, W6, W7, W8 and W9 [31:0]**

### Special Registers
The special registers have specific purposes and exist to handle essential functions for the machine's operation.
For the specific case of this machine, six registers were defined, mainly to assist in memory manipulation, all storing values up to 32 bits.

- **Program Counter (PC) [31:0]**: Stores the address of the next instruction to be executed. Is automaticaly incremented after every instruction cycle, unless modified by a JUMP instruction.
- **Instruction Register (IR) [31:0]**: Contains the current instruction being decoded and executed.
- **Memory Data Register (MDR) [31:0]**: Holds the data being transferred from or to memory.
- **Stack Pointer (SP) [31:0]**: Points to the top of the stack, used to manage function calls and local variable storage.
- **Memory Address Register (MAR) [31:0]**: Stores the memory address where reading or writing operations will be executed.
- **Status Register (SR) [31:0]**: Stores flags that indicate the result of test operations executed and current CPU flags. The first three bits are reserved for the G, L and Z flags, bit 3 indicates whether the CPU is in kernel mode (0) or in user mode (1), and bit 4 indicates if interruptions are disabled (0) or enabled (1).
- **Exception Cause Register (ECR) [31:0]**: Stores a specific hardware code indicating the reason the exception or interrupt was triggered (e.g., an invalid instruction, a system call, or a hardware timer interrupt).
- **Exception Supervisor Address (ESA) [31:0]**: Stores the base memory address of the exception supervisor routine. When an exception occurs, the CPU automatically jumps to this address so the supervisor can route execution to the appropriate specific handler.
- **Exception Status Register (ESR) [31:0]**: Backs up the exact state of the Status Register (SR) at the moment the exception occurred.
- **Exception Program Counter (EPC) [31:0]**: Stores the value of the Program Counter (PC) at the exact instruction where the exception occurred.
- **Base (BASE) [31:0]**: Stores the starting physical address position of the running user process, to be managed by the MMU.
- **Limit (LIMIT) [31:0]**: Stores the final physical address position of the running user process, to be managed by the MMU.

## Interruptions

For this machine, when interruptions occur, the cause of the interruption is saved in the ecr, the status is saved in the esr, the pc is saved in the epc and is set as the esa. This machine contains 4 types of interruptions implemented:

#### Clock - ID 0

Triggered by an internal instruction counter. When the counter reaches a predefined number, the CPU triggers a clock interruption.

#### Input - ID 1
Triggered by a inputFlag. This interruption expects data from the buffer.

#### Kill - ID 2 
Triggered by a killFlag. This interruption expects a process identifier from the buffer.

#### System Call - ID 3

Triggered by the syscall instruction, its id is expected to be at w8.

#### Fault - ID 4

Triggered by access outside base limit, use of privileged instruction/register when in user mode.

## Buffer
Buffer is a part of the memory designed to receive outside data. Buffer is defined as the last 64KB of the memory and receives the byte size of the input in the first 32 bits of the buffer.

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
- **Opcode (decimal)**: 0

### Arithmetic and Logical Operations
**AND**:
- **Description**: Does bitwise AND to the values of two registers and stores the result in a destination register.
- **Syntax**: AND \<Destination Reg.>, \<Input Reg.>, \<Input Reg.>
- **Example**: AND W0, W1, W0
- **Format**: R-Type
- **Opcode (decimal)**: 1

**OR**:
- **Description**: Does bitwise OR to the values of two registers and stores the result in a destination register.
- **Syntax**: OR \<Destination Reg.>, \<Input Reg.>, \<Input Reg.>
- **Example**: OR W0, W1, W0
- **Format**: R-Type
- **Opcode (decimal)**: 2

**XOR**:
- **Description**: Does bitwise XOR to the values of two registers and stores the result in a destination register.
- **Syntax**: XOR \<Destination Reg.>, \<Input Reg.>, \<Input Reg.>
- **Example**: XOR W0, W1, W0
- **Format**: R-Type
- **Opcode (decimal)**: 3

**SHL**:
- **Description**: Performs a bidirectional Logical Shift on a register based on the shift amount in a second register. Shifts left if the amount is positive, and performs a logical right shift (padding with 0s) if the amount is negative. Stores the result in a destination register.
- **Syntax**: SHL \<Destination Reg.>, \<Input Reg.>, \<Input Reg.>
- **Example**: SHL W0, W1, W2
- **Format**: R-Type
- **Opcode (decimal)**: 4

**SHA**:
- **Description**: Performs a bidirectional Arithmetic Shift on a register based on the shift amount in a second register. Shifts left if the amount is positive, and performs an arithmetic right shift (padding with the sign bit) if the amount is negative. Stores the result in a destination register.
- **Syntax**: SHA \<Destination Reg.>, \<Input Reg.>, \<Input Reg.>
- **Example**: SHA W0, W1, W2
- **Format**: R-Type
- **Opcode (decimal)**: 5

**ADD**:
- **Description**: Adds the values of two registers and saves the result in a third one.
- **Syntax**: ADD \<Destination Reg.>, \<Input Reg.>, \<Input Reg.>
- **Example**: ADD W0, W1, W0
- **Format**: R-Type
- **Opcode (decimal)**: 6

**SUB**:
- **Description**: Subtracts the value of one register from another and stores the result in a third one.
- **Syntax**: SUB \<Destination Reg.>, \<Input Reg.>, \<Input Reg.>
- **Example**: SUB W0, W1, W0
- **Format**: R-Type
- **Opcode (decimal)**: 7

**MUL**:
- **Description**: Multiplies the value of two registers and stores the result in a third one.
- **Syntax**: MUL \<Destination Reg.>, \<Input Reg.>, \<Input Reg.>
- **Example**: MUL W0, W1, W0
- **Format**: R-Type
- **Opcode (decimal)**: 8

**UDIV**:
- **Description**: Divides the unsigned value of one register by another and stores the result in a third one, discarding the remainder.
- **Syntax**: UDIV \<Destination Reg.>, \<Input Reg.>, \<Input Reg.>
- **Example**: UDIV W0, W1, W0
- **Format**: R-Type
- **Opcode (decimal)**: 9

**SDIV**:
- **Description**: Divides the signed value of one register by another and stores the result in a third one, discarding the remainder.
- **Syntax**: SDIV \<Destination Reg.>, \<Input Reg.>, \<Input Reg.>
- **Example**: SDIV W0, W1, W0
- **Format**: R-Type
- **Opcode (decimal)**: 10

### Test Instructions
**CMP**:
- **Description**: Compares the values of two registers and stores the flag in the SR.
- **Syntax**: CMP \<Input Reg.>, \<Input Reg.>
- **Example**: CMP W0, W1
- **Format**: R-Type
- **Opcode (decimal)**: 11

#### Z, L and G Test Flags
When a test instruction, such as CMP, is executed, the SR register is updated, and its value can be used by other instructions to change the program's flow. Each flag is represented by a bit, and the flag being set indicates that the bit value is 1.
Considering the SR register [31:0], the following flags can be set in their respective bits:
- **Z**, bit [0] - Set if the result of the last comparison is 0, meaning the operands are equal.
- **L**, bit [1] - Set if the result of the last comparison indicates that the first operand is less than the second.
- **G**, bit [2] - Set if the result of the last comparison indicates that the first operand is greater than the second.

These flags can be used by instructions to make decisions that can change the program flow.

### Control Flow Operations
**JUMP**:
- **Description**: Unconditionally jumps the Program Counter (PC) forward or backward by a specific instruction offset (PC-relative).
- **Syntax**: JUMP [\<Label/Offset\>]
- **Example**: JUMP 0x14
- **Format**: I-Type
- **Opcode (decimal)**: 12

**JMPR**:
- **Description**: Unconditionally sets the Program Counter (PC) to an absolute memory address currently stored in a register.
- **Syntax**: JMPR \<Source Reg.\>
- **Example**: JMPR W5
- **Format**: R-Type
- **Opcode (decimal)**: 13

**BEQ**:
- **Description**: Conditionally jumps the PC forward or backward by a specific instruction offset (PC-relative) if the Z (Equal) flag in the Status Register is set.
- **Syntax**: BEQ \<Label/Offset\>
- **Example**: BEQ 0x05
- **Format**: I-Type
- **Opcode (decimal)**: 14

**BLT**:
- **Description**: Conditionally jumps the PC forward or backward by a specific instruction offset (PC-relative) if the L (Less Than) flag in the Status Register is set.
- **Syntax**: BLT \<Label/Offset\>
- **Example**: BLT 0x03
- **Format**: I-Type
- **Opcode (decimal)**: 15

**BGT**:
- **Description**: Conditionally jumps the PC forward or backward by a specific instruction offset (PC-relative) if the G (Greater Than) flag in the Status Register is set.
- **Syntax**: BGT \<Label/Offset\>
- **Example**: BGT 0x02
- **Format**: I-Type
- **Opcode (decimal)**: 16

### Load and Store Operations with Addresses
**LOAD**:
- **Description**: Loads the content stored at a specific memory address into a specific register.
- **Syntax**: LOAD \<Destination Reg.>, [\<Address Reg.>]
- **Example**: LOAD W0, W1
- **Format**: R-Type
- **Opcode (decimal)**: 17

**STORE**:
- **Description**: Stores the value of a register to memory.
- **Syntax**: STORE \<Source Reg.>, [\<Address Reg.>]
- **Example**: STORE W1, W2
- **Format**: R-Type
- **Opcode (decimal)**: 18

**LDD**:
- **Description**: Loads the content stored at a specific memory address into a specific register.
- **Syntax**: LDD \<Destination Reg.>, [\<Address>]
- **Example**: LDD W0, #0x123
- **Format**: I-Type
- **Opcode (decimal)**: 19

**STRD**:
- **Description**: Stores the value of a register to memory.
- **Syntax**: STRD \<Source Reg.>, [\<Address>]
- **Example**: STRD W1, #0x123
- **Format**: I-Type
- **Opcode (decimal)**: 20

**LDB**:
- **Description**: Loads a single 8-bit unsigned byte from a memory address into a register. The loaded byte is zero-extended to fill the 32-bit register.
- **Syntax**: LDB \<Destination Reg.\>, [\<Address Reg.\>]
- **Example**: LDB W2, W1
- **Format**: R-Type
- **Opcode (decimal)**: 21

**LDSB** (Load Signed Byte):
- **Description**: Loads a single 8-bit signed byte from a memory address into a register. The loaded byte is sign-extended to fill the 32-bit register, preserving its arithmetic sign.
- **Syntax**: LDSB <Destination Reg.>, [<Address Reg.>]
- **Example**: LDSB W4, W1
- **Format**: R-Type
- **Opcode (decimal)**: 22

**STRB** (Store Byte):
- **Description**: Stores the lowest 8 bits (one byte) from a register into a specific memory address. The upper 24 bits of the source register are ignored.
- **Syntax**: STRB <Source Reg.>, [<Address Reg.>]
- **Example**: STRB W3, W1
- **Format**: R-Type
- **Opcode (decimal)**: 23

**MRET**
- **Description**: Return from an exception. Restores the processor to its pre-exception state by copying the Exception Status Register (ESR) back into the Status Register (SR), and the Exception Program Counter (EPC) back into the Program Counter (PC).
- **Syntax and Example**: MRET
- **Format**: I-Type
- **Opcode (decimal)**: 24

**SYSCALL**
- **Description**: Triggers a synchronous exception to transfer control to the operating system's exception to request privileged services, the immediate code is put into W9 and the syscall return is put back into W9.
- **Syntax and Example**: SYSCALL #2
- **Format**: I-Type
- **Opcode (decimal)**: 25

### Processor Execution Cycle
**FETCH**
- **Description**: Get the next instruction from memory using the address stored in the Program Counter (PC) and load it into the Instruction Register (IR).
- **Syntax and Example**: FETCH
- **Format**: I-Type
- **Opcode (decimal)**: 26

### Zepa Machine Instruction Encoding Table

| **Instruction** | **Format** | **opcode** | **rd** | **rs1** | **rs2** | **funct5** | **funct6** |
|-----------------|------------|------------|--------|---------|---------|------------|------------|
| **AND**         | R-Type     | 000001     | reg    | reg     | reg     | 00000      | 000000     |
| **OR**          | R-Type     | 000010     | reg    | reg     | reg     | 00000      | 000000     |
| **XOR**         | R-Type     | 000011     | reg    | reg     | reg     | 00000      | 000000     |
| **SHL**         | R-Type     | 000100     | reg    | reg     | reg     | 00000      | 000000     |
| **SHA**         | R-Type     | 000101     | reg    | reg     | reg     | 00000      | 000000     |
| **ADD**         | R-Type     | 000110     | reg    | reg     | reg     | 00000      | 000000     |
| **SUB**         | R-Type     | 000111     | reg    | reg     | reg     | 00000      | 000000     |
| **MUL**         | R-Type     | 001000     | reg    | reg     | reg     | 00000      | 000000     |
| **UDIV**        | R-Type     | 001001     | reg    | reg     | reg     | 00000      | 000000     |
| **SDIV**        | R-Type     | 001010     | reg    | reg     | reg     | 00000      | 000000     |
| **CMP**         | R-Type     | 001011     | 00000  | reg     | reg     | 00000      | 000000     |
| **JMPR**        | R-Type     | 001101     | 00000  | reg     | 00000   | 00000      | 000000     |
| **LOAD**        | R-Type     | 010001     | 00000  | reg     | reg     | 00000      | 000000     |
| **STORE**       | R-Type     | 010010     | 00000  | reg     | reg     | 00000      | 000000     |
| **LDB**         | R-Type     | 010101     | 00000  | reg     | reg     | 00000      | 000000     |
| **LDSB**        | R-Type     | 010110     | 00000  | reg     | reg     | 00000      | 000000     |
| **STRB**        | R-Type     | 010111     | 00000  | reg     | reg     | 00000      | 000000     |


| **Instruction** | **Format** | **opcode** | **rs1/rd** | **immediate**    | **funct5** |
|-----------------|------------|------------|------------|------------------|------------|
| **MV**          | I-Type     | 000000     | reg        | 16bit constant   | 00000      |
| **JUMP**        | I-Type     | 001100     | 00000      | 16bit address    | 00000      |
| **BEQ**         | I-Type     | 001110     | 00000      | 16bit offset     | 00000      |
| **BLT**         | I-Type     | 001111     | 00000      | 16bit offset     | 00000      |
| **BGT**         | I-Type     | 010000     | 00000      | 16bit offset     | 00000      |
| **LDD**         | I-Type     | 010011     | reg        | 16bit address    | 00000      |
| **STRD**        | I-Type     | 010100     | reg        | 16bit address    | 00000      |
| **MRET**        | I-Type     | 011000     | 00000      | 0000000000000000 | 00000      |
| **SYSCALL**     | I-Type     | 011001     | 00000      | code             | 00000      |
| **FETCH**       | I-Type     | 011010     | 00000      | 0000000000000000 | 00000      |


## References
- [Bit by Bit: Processadores Cortex-M0+ - Arquitetura do Conjunto de Instruções (ISA)](https://bit-by-bit.gitbook.io/embedded-systems/processadores-cortex-m0+/arquitetura-do-conjunto-de-instrucoes-isa)
- [ARM Developer: ARMv8-M Architecture Reference Manual](https://developer.arm.com/documentation/102374/0101)
- [RISC-V International: RISC-V Specifications](https://riscv.org/technical/specifications/)
- [Go Programming Language Specification](https://go.dev/ref/spec)
- [GeeksforGeeks: Essential Registers for Instruction Execution](https://www.geeksforgeeks.org/essential-registers-for-instruction-execution/)
###

