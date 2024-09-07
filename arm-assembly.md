
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

#### Examples:
- **Adding 3 different values and storing the result in the W0 register:**
    ```
    ADD W2, #3, #4
    ADD W0, W2, #5
    ```

### Macros
Macros allow the reuse of repetitive code, reducing duplication and simplifying complex operations. A macro can be defined using the **_macro** flag and ends with the **_endmacro** flag.

#### General Macro Syntax:
```
_macro <name>, <operand1>, <operand2>, ...
    <opcode> <operand1>, <operand2>, ...
    ...
_endmacro
```

#### Example:
- **Macro Definition:**
    ```
    _macro MoveAdd reg1, reg2, value, value2
        MV reg1, value
        ADD reg2, reg1, value2
    _endmacro
    ```
- **Usage:**
    ```
    MoveAdd W0, W2, #2, #3
    ```