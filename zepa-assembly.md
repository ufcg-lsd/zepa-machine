
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

#### Example:
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

## CMP (Compare)
Compares the values of two registers and updates the flags in the status register (SR).

**Syntax:**
CMP <Op1>, <Op2>

**Example:**
```
MV W1, #3
MV W2, #2
CMP W1, W2    ; Sets the G flag (W1 > W2)
```

### JUMP (Unconditional Jump)
Unconditionally jumps to a specific address or label, modifying the Program Counter (PC).

**Syntax:**
```
JUMP <Label/Address>
```

**Example:**
```
JUMP _LOOP    ; Jump to the _LOOP label
```

## JZ (Jump if Zero)
Jumps to a specific address if the Z flag is set (if two operands are equal).

**Syntax:**
```
JZ <Label/Address>
```

**Example:**
```
MV W1, #2
MV W2, #2
CMP W1, W2    ; Sets the Z flag (W1 == W2)
JZ _LESS      ; Jumps to _END if Z == 1
```

## JG (Jump if Greater)
Jumps if the G flag is set (if the first operand is greater than the second).

**Syntax:**
```
JG <Label/Address>
```

**Example:**
```
MV W1, #5
MV W2, #3
CMP W1, W2    ; Sets the G flag (W1 > W2)
JG _GREATER  ; Jumps to _CONTINUE if G == 1
```

## JL (Jump if Less)
Jumps if the L flag is set (if the first operand is less than the second).

**Syntax:**
```
JL <Label/Address>
```

**Example:**
```
MV W1, #1
MV W2, #3
CMP W1, W2    ; Sets the L flag (W1 < W2)
JL _EQUAL     ; Jumps to _RETRY if L == 1
```

**Example of JUMP and Control Flow:**
```
_start:
    MV W0, #10         ; W0 = 10
    MV W1, #20         ; W1 = 20
    CMP W0, W1         ; Compare W0 with W1

    JG _GREATER        ; Jump to _GREATER if W0 > W1 (G flag set)
    JL _LESS           ; Jump to _LESS if W0 < W1 (L flag set)

_EQUAL:
    MV W2, #0          ; If W0 == W1, W2 = 0
    JUMP _END          ; Jump to _END

_GREATER:
    MV W2, #1          ; If W0 > W1, W2 = 1
    JUMP _END          ; Jump to _END

_LESS:
    MV W2, #2          ; If W0 < W1, W2 = 2

_END:
    ; End of program
```

