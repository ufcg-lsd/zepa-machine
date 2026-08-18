
# Assembly for Zepa Machine Architecture

## Introduction
This document defines the assembly programming language for the instructions of the Zepa machine. Every program written in this assembly needs to start with a **_start** flag, marking the entry point of the execution.

## Syntax

### General Syntax
The general syntax follows the structure:

```
<opcode> <operand1>, <operand2>, ...
```

Where the **opcode** is the operation code (instruction), and the **operands** are registers or constant values that participate in the operation.

**Example:**
- **Adding 3 different values and storing the result in the W0 register:**
    ```
    ADD W2, #3, #4
    ADD W0, W2, #5
    ```
## Literals (Immediate Values)
In zepa-assembly, literals are constant values prefixed with # and can be used directly in instructions. These immediate values can be moved to registers or used in arithmetic operations.

**Syntax:**
```
#<Value>
```

**Example:**
```
MV W1, #3    ; Move the literal value 3 into register W1
```

### MV (Move)
Moves an immediate value or the contents of one register to another register.

**Syntax:**
```
MV <Dest Reg.>, #<Value> or <Source Reg.>
```
**Example:**
```
MV W1, #5    ; Move the value 5 into register W1
MV W2, W1    ; Move the value from W1 into W2
```

## Arithmetic and Logical Operations

### AND (And)
Does bitwise AND to the values of two registers and stores the result in a destination register.

**Syntax:**
```
AND <Dest Reg.>, <Op1>, <Op2>
```

**Example:**
```
MV W1, #3
MV W2, #2
AND W0, W1, W2    ; W0 = 3 & 2
```

### OR (Or)
Does bitwise OR to the values of two registers and stores the result in a destination register.

**Syntax:**
```
OR <Dest Reg.>, <Op1>, <Op2>
```

**Example:**
```
MV W1, #3
MV W2, #2
OR W0, W1, W2    ; W0 = 3 | 2
```

### XOR (Xor)
Does bitwise XOR to the values of two registers and stores the result in a destination register.

**Syntax:**
```
XOR <Dest Reg.>, <Op1>, <Op2>
```

**Example:**
```
MV W1, #3
MV W2, #2
XOR W0, W1, W2    ; W0 = 3 ^ 2
```

### SHL (Shift Logical)
Performs a bidirectional logical shift on a register based on the shift amount in a second register. Shifts left if the amount is positive, and performs a logical right shift (padding with 0s) if the amount is negative.

**Syntax:**
```
SHL <Dest Reg.>, <Op1>, <Op2>
```

**Example:**
```
MV W1, #4
MV W2, #2
SHL W0, W1, W2    ; W0 = 4 << 2 (Logical shift left by 2)
```

### SHA (Shift Arithmetic)
Performs a bidirectional arithmetic shift on a register based on the shift amount in a second register. Shifts left if the amount is positive, and performs an arithmetic right shift (padding with the sign bit) if the amount is negative.

**Syntax:**
```
SHA <Dest Reg.>, <Op1>, <Op2>
```

**Example:**
```
MV W1, #-8
MV W2, #-2
SHA W0, W1, W2    ; W0 = -8 >> 2 (Arithmetic shift right by 2)
```

### ADD (Add)
Adds the values of two registers and stores the result in a destination register.

**Syntax:**
```
ADD <Dest Reg.>, <Op1>, <Op2>
```

**Example:**
```
MV W1, #3
MV W2, #2
ADD W0, W1, W2    ; W0 = 3 + 2
```

### SUB (Subtract)
Subtracts the value of the second register from the first.

**Syntax:**
```
SUB <Dest Reg.>, <Op1>, <Op2>
```

**Example:**
```
MV W1, #5
MV W2, #2
SUB W0, W1, W2    ; W0 = 5 - 2
```

### MUL (Multiply)
Multiplies the values of two registers.

**Syntax:**
```
MUL <Dest Reg.>, <Op1>, <Op2>
```

**Example:**
```
MV W1, #4
MV W2, #3
MUL W0, W1, W2    ; W0 = 4 * 3
```

### UDIV (Unsigned Divide)
Divides the unsigned values of two registers, the remainder is discarded.

**Syntax:**
```
UDIV <Dest Reg.>, <Op1>, <Op2>
```

**Example:**
```
MV W1, #6
MV W2, #3
UDIV W0, W1, W2    ; W0 = 6 / 3
```

### SDIV (Signed Divide)
Divides the signed values of two registers, the remainder is discarded.

**Syntax:**
```
SDIV <Dest Reg.>, <Op1>, <Op2>
```

**Example:**
```
MV W1, #-5
MV W2, #2
SDIV W0, W1, W2    ; W0 = (-5) / 2
```

## Control Flow Operations

### CMP (Compare)
Compares the values of two registers and updates the flags in the status register (SR).

**Syntax:**
```
CMP <Op1>, <Op2>
```

**Example:**
```
MV W1, #3
MV W2, #2
CMP W1, W2    ; Sets the G flag (W1 > W2)
```


### JUMP (Unconditional Jump)
Unconditionally jumps the PC forward or backward by a specific instruction offset (PC-relative) or to a specific label.

**Syntax:**
```
JUMP <Label/Offset>
```

**Example:**
```
JUMP _loop    ; Jump to the _loop label
```

### JMPR (Jump Register)
Unconditionally sets the Program Counter (PC) to an absolute memory address currently stored in a register.

**Syntax:**
```
JMPR <Source Reg.>
```

**Example:**
```
MV W5, #0x0040      ; Store a return address in W5
JMPR W5             ; Jump exactly to the address stored in W5
```

### BEQ (Branch if Equal)
Conditionally jumps the PC forward or backward by a specific instruction offset (PC-relative) or to a specific label if the Z (Equal) flag in the Status Register is set.

**Syntax:**
```
BEQ <Label/Offset>
```

**Example:**
```
CMP W1, W2
BEQ _equal_logic    ; Jump to the _equal_logic label if W1 == W2
```

### BLT (Branch if Less Than)
Conditionally jumps the PC forward or backward by a specific instruction offset (PC-relative) or to a specific label if the L (Less Than) flag in the Status Register is set.

**Syntax:**
```
BLT <Label/Offset>
```

**Example:**
```
CMP W1, W2
BLT _less_logic     ; Jump to the _less_logic label if W1 < W2
```

### BGT (Branch if Greater Than)
Conditionally jumps the PC forward or backward by a specific instruction offset (PC-relative) or to a specific label if the G (Greater Than) flag in the Status Register is set.

**Syntax:**
```
BGT <Label/Offset>
```

**Example:**
```
CMP W1, W2
BGT _greater_logic  ; Jump to the _greater_logic label if W1 > W2
```

## Memory Operations

### LOAD (Load from Memory)
Loads a value from a memory address into a register.

**Syntax:**
```
LOAD <Dest Reg.>, [< Address Reg.>]
```

**Example:**
```
LOAD W1, [W2]    ; Load the value stored at memory address in W2 into register W1
```

### STORE (Store to Memory)
Stores the value from a register into a memory address.

**Syntax:**
```
STORE <Source Reg.>, [< Address Reg.>]
```

**Example:**
```
STORE W1, [W2]   ; Store the value from register W1 into memory address inside W2
```

### LDD (Load Direct)
Loads a value from a memory address into a register.

**Syntax:**
```
LDD <Dest Reg.>, [< Address >]
or LDD <Dest Reg.>, #<ADRESS>
```

**Example:**
```
LDD W1, [0x123]    ; Load the value stored at memory address 0x123 into register W1
```

### STRD (Store Direct)
Stores the value from a register into a memory address.

**Syntax:**
```
STRD <Source Reg.>, [< Address >]
or STRD <Source Reg.>, #< Address >
```

**Example:**
```
STRD W1, [0x123]   ; Store the value from register W1 into memory address 0x123
```

### LDB (Load Byte)
Loads a single 8-bit unsigned byte from a memory address into a register. The loaded byte is zero-extended to fill the 32-bit register.

**Syntax:**
```
LDB <Dest Reg.>, [< Address Reg.>]
```

**Example:**
```
LDB W1, [W2]    ; Load an unsigned byte from the memory address in W2 into register W1
```

### LDSB (Load Signed Byte)
Loads a single 8-bit signed byte from a memory address into a register. The loaded byte is sign-extended to fill the 32-bit register, preserving its negative or positive arithmetic value.

**Syntax:**
```
LDSB <Dest Reg.>, [< Address Reg.>]
```

**Example:**
```
LDSB W1, [W2]   ; Load a sign-extended byte from the memory address in W2 into register W1
```

### STRB (Store Byte)
Stores the lowest 8 bits (one byte) from a register into a specific memory address. The upper 24 bits of the source register are ignored, and adjacent memory blocks are left untouched.

**Syntax:**
```
STRB <Source Reg.>, [< Address Reg.>]
```

**Example:**
```
STRB W1, [W2]   ; Store only the lowest byte of register W1 into the memory address inside W2
```

### MRET (Machine Return)
Return from an exception. Restores the processor to its pre-exception state by copying the Exception Status Register (ESR) back into the Status Register (SR), and the Exception Program Counter (EPC) back into the Program Counter (PC).

**Syntax:**
```
MRET
```

**Example:**
```
MRET   ; Return from an exception / interruption
```

### SYSCALL (System Call)
Triggers a synchronous exception to transfer control to the operating system's exception to request privileged services, the immediate code is put into W9.

**Syntax:**
```
SYSCALL <CODE>
```

**Example:**
```
SYSCALL #2   ; Calls the kernel with the system call of code 2
```

## References

- [ARM Assembly](https://armasm.com/)