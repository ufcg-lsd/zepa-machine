# ISA
## Introduction
The objective of this Instruction Set Architecture (ISA) is to define the registers and assembly instructions for the VM, for now mainly working with memory manipulation. For creating this document, the ISAs of ARM64 and RISC-V were used as reference.
## Definition of Registers
### General Purpose Registers
The general-purpose registers are flexible and can be used in different situations, as temporarily storing values, used to assist in operations, and storing memory addresses.

Initially, this VM has 6 registers, which are named W0 to W5, each of them capable of performing 8 bit

- **W0, W1, W2, W3, W4, and W5 [7:0]**

### Special Registers
The special registers have specific purposes and exist to handle essential functions for the machine's operation.
For the specific case of this VM, six registers were defined, mainly to assist in memory manipulation, all storing values up to 8 bits.

- **Program Counter (PC) [7:0]**: Stores the address of the next instruction to be executed. Is automaticaly incremented after every instruction cycle, unless modified by a JUMP instruction.
- **Instruction Register (IR) [7:0]**: Contains the current instruction being decoded and executed.
- **Memory Data Register (MDR) [7:0]**: Holds the data being transferred from or to memory.
- **Stack Pointer (SP) [7:0]**: Points to the top of the stack, used to manage function calls and local variable storage.
- **Memory Address Register (MAR) [7:0]**: Stores the memory address where reading or writing operations will be executed.
- **Status Register (SR) [7:0]**: Stores flags that indicate the result of test operations executed. The first bits are reserved for the N and Z flags, and the last ones are flexible.

## Definition of Instructions

### Attribution
**MV**:
- **Description**: Moves a constant to a specific register.
- **Syntax**: MV \<Destination Reg.> \#\<Constant>
- **Example**: MV W1 #5

### Arithmetic and Logical Operations
**ADD**:
- **Description**: Adds the values of two registers and saves the result in a third one.
- **Syntax**: ADD \<Destination Reg.>, \<Input Reg.>, \<Input Reg.>
- **Example**: ADD W0, W1, W0

**SUB**:
- **Description**: Subtracts the value of one register from another and stores the result in a third one.
- **Syntax**: SUB \<Destination Reg.>, \<Input Reg.>, \<Input Reg.>
- **Example**: SUB W0, W1, W0

### Test Instructions
**CMP**:
- **Description**: Compares the values of two registers and stores the flag in the SR.
- **Syntax**: CMP \<Input Reg.>, \<Input Reg.>
- **Example**: CMP W0, W1

#### N and Z Test Flags
When a test instruction, such as CMP, is executed, the SR register is updated, and its value can be used by other instructions to change the program's flow. Each flag is represented by a bit, and the flag being set indicates that the bit value is 1.
Considering the SR register [7:0], the following flags can be set in their respective bits:
- **N**, bit [7] - Set if the result of the last test was negative, meaning different.
- **Z**, bit [6] - Set if the result of the last test was 0, meaning equal.

These flags can be used by instructions to make decisions that can change the program flow.

### Control Flow Operations
**JUMP**:
- **Description**: Change the value of the Program Counter (PC) register, updating the program's execution flow.
- **Syntax**: JUMP \[<Address/Label>]
- **Example**: JUMP START_LOOP

### Load and Store Operations with Addresses
**LOAD**:
- **Description**: Loads the content stored at a specific memory address into a specific register.
- **Syntax**: LOAD \<Destination Reg.>, [\<Address>]
- **Example**: LOAD W0, 0x68DB00AD

**STORE**:
- **Description**: Stores the value of a register to memory.
- **Syntax**: STORE \<Source Reg.>, [\<Address>]
- **Example**: STORE W1, 0x68DB00AD

### Processor Execution Cycle
**FETCH**
- **Description**: Get the next instruction from memory using the address stored in the Program Counter (PC) and load it into the Instruction Register (IR).
- **Syntax and Example**: FETCH

## References
https://bit-by-bit.gitbook.io/embedded-systems/processadores-cortex-m0+/arquitetura-do-conjunto-de-instrucoes-isa
https://developer.arm.com/documentation/102374/0101
https://riscv.org/technical/specifications/
https://go.dev/ref/spec
###

