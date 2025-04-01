
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


### BEQ (Branch if equals)
Use comparison to go to a specific address if comparison are equal.

**Syntax:**
```
BEQ <Label/Address>
```

**Example:**
```
MV W1, #3
MV W2, #3
CMP W1, W2    ; Sets the Z flag (W1 = W2)
BEQ #10       ; Go to instruction at address 10
```

### BLT (Branch if lower than)
Use comparison to go to a specific address if comparison are lower.

**Syntax:**
```
BLT <Label/Address>
```

**Example:**
```
MV W1, #2
MV W2, #3
CMP W1, W2    ; Sets the Z flag (W1 < W2)
BLT #10       ; Go to instruction at address 10
```

### BGT (Branch if greater than)
Use comparison to go to a specific address if comparison are greater.

**Syntax:**
```
BGT <Label/Address>
```

**Example:**
```
MV W1, #3
MV W2, #2
CMP W1, W2    ; Sets the Z flag (W1 > W2)
BGT #10       ; Go to instruction at address 10
```

### JUMP (Unconditional Jump)
Unconditionally jumps to a specific address or label, modifying the Program Counter (PC).

**Syntax:**
```
JUMP <Label/Address>
```

**Example:**
```
JUMP _loop    ; Jump to the _loop label
```

### HALT (Halt)
Terminate program.

**Syntax:**
```
HALT
```

**Example:**
```
HALT
MV W0 #1 ; Not executed
```

### UDF (Undefined)
Raise an exception.

**Syntax:**
```
UDF
```

**Example:**
```
UDF      ; Show: Exception Raised
MV W0 #1 ; Not executed
```

## Memory Operations

### LOAD (Load from Memory)
Loads a value from a memory address into a register.

**Syntax:**
```
LOAD <Dest Reg.>, [<Memory Address>]
```

**Example:**
```
LOAD W1, [0x100]    ; Load the value stored at memory address 0x100 into register W1
```

### STORE (Store to Memory)
Stores the value from a register into a memory address.

**Syntax:**
```
STORE <Source Reg.>, [<Memory Address>]
```

**Example:**
```
STORE W1, [0x200]   ; Store the value from register W1 into memory address 0x200
```

## References

- [ARM Assembly](https://armasm.com/)