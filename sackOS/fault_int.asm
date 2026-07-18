fault_int:
  MV W0, #pcb_v        ; pcb_v initial address
  LDD W9, #running_pid ; running_pid
  MV W1, #84           ; pcb_size

  MUL W1, W9, W1       ; W1 = running_pid * pcb_size
  ADD W0, W0, W1       ; W0 = pcb_v[running_pid] initial address

  MV W1, #56           ; pcb W9 offset
  ADD W0, W0, W1       ; W0 = pcb_v[running_pid].W9 initial address

  MV W1, #1
  STORE W1, W0         ; pcb_v[running_pid].W9 = 1

  JUMP kill            ; kill(running_pid)