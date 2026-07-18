_exit:

	LDD W9, [_running_pid]   ; W9 = running_pid
	MV W1, #84               ; W1 = pcb_size
	MUL W0, W9, W1           ; W0 = running_pid * pcb_size

	MV W1, #pcb_v            ; W1 = pcb_v initial address
	ADD W0, W0, W1           ; W0 = pcb_v[running_pid] initial address

	MV W1, #52
	ADD W0, W0, W1           ; W0 = pcb_v[running_pid].W8 address
	LOAD W2, W0              ; W2 = pcb_v[running_pid].W8 (status_code)

	MV W1, #4
	ADD W0, W0, W1           ; W0 = pcb_v[running_pid].W9 address
	STORE W2, W0             ; pcb_v[running_pid].W9 = pcb_v[running_pid].W8

	JUMP _kill