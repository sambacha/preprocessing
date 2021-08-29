package main

import (
	"fmt"
)

type graph struct {
	root *vertex
}

type vertex struct {
	parentID   uint64
	direction  int    // indicates whether this vertex is left of right, -1 root
	start      uint64 // starting point of the bytecode
	stop       uint64 // ending point
	jump_dest  uint64 // 0 means no jumpdest
	left_vtx   *vertex
	right_vtx  *vertex
	ctx        *callCtx
	stack_size int // initial stack size
	stack      Stack
}

func new_vtx(start, parentID uint64, direction int) *vertex {
	vtx := &vertex{
		start:     start,
		parentID:  parentID,
		direction: direction,
	}
	return vtx
}

func (vtx *vertex) set_ctx(ctx *callCtx) *vertex {
	vtx.ctx = ctx
	vtx.stack_size = ctx.stack.Len()
	vtx.stack = *ctx.stack
	return vtx
}

func (vtx *vertex) print() {
	fmt.Printf("vertex { parendID:%d, direction:%d, start:%d, stop: %d, jump_dest: %d }\n", vtx.parentID, vtx.direction, vtx.start, vtx.stop, vtx.jump_dest)
	vtx.stack.Print()
}

func (vtx *vertex) run(evm *evm, valid_jumpdests *[]bool, code_size *uint64) {

	pc := vtx.start
	jump_dest, _ := evm.interpreter.g_run(&pc, vtx.ctx)
	vtx.stop = pc + 1
	if jump_dest < *code_size && (*valid_jumpdests)[jump_dest] {
		// fmt.Println("VALID JUMP")
		vtx.jump_dest = jump_dest
	}
}

func new_graph(evm *evm, contract *Contract, input []byte) {
	fmt.Println("*******************CREATING A GRAPH*******************")

	bytecode := contract.Code
	code_size := uint64(len(bytecode))

	if code_size == 0 {
		// fmt.Println("0 length bytecode")
		return
	}

	valid_jumpdests := make_valid_jumpdests(&bytecode)

	callContext := &callCtx{
		memory:   NewMemory(),
		stack:    NewStack(),
		contract: contract,
	}

	f := make_dot_file()
	defer f.Close()

	visited_map := make(map[uint64]*vertex)

	root := new_vtx(0, 0, -1).set_ctx(callContext)
	// visited_map[0] = root
	vtxs := []*vertex{root}

	var vtx *vertex
	for len(vtxs) > 0 {
		vtx, vtxs = vtxs[0], vtxs[1:]

		if _, ok := visited_map[vtx.start]; !ok {
			// we have never visited this vertex
			visited_map[vtx.start] = vtx

			pc := vtx.start
			jump_dest, is_jump := evm.interpreter.g_run(&pc, vtx.ctx)
			vtx.stop = pc + 1

			jump_flag := false
			if jump_dest < code_size && valid_jumpdests[jump_dest] {
				jump_flag = true
				vtx.jump_dest = jump_dest
			}

			write_vtx(f, &bytecode, vtx)

			if jump_dest == REVERTS || jump_dest == HALTS ||
				jump_dest == INVALID_OP {
				continue
			}

			if is_jump {
				// code execution starting point (pc) of the parent vertex
				parentID := vtx.start
				if vtx.stop < code_size && bytecode[pc] == JUMPI {
					ctx := vtx.ctx.copy()
					left_vtx := new_vtx(vtx.stop, parentID, 0).set_ctx(ctx)
					vtx.left_vtx = left_vtx
					vtxs = append(vtxs, left_vtx)
				}

				if jump_flag {
					// vtx.jump_dest = jump_dest
					right_vtx := new_vtx(jump_dest, parentID, 1).set_ctx(vtx.ctx)
					vtx.right_vtx = right_vtx
					vtxs = append(vtxs, right_vtx)
				}
			}

		} else {

			vtx.run(evm, &valid_jumpdests, &code_size)

			write_vtx(f, &bytecode, vtx)
		}
	}

	if f != nil {
		fmt.Fprint(f, "}\n")
	}
}
