package main

import (
	"hash"

	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/common/math"
)

const (
	REVERTS = 0xFFFFFFFFFFFFFF00 + iota
	HALTS
	INVALID_OP
	STACK_UNDERFLOW
	STACK_OVERFLOW
	END_OF_LOOP
	GAS_UNIT_OVERFLOW
	GAS_CONST_ERR
	TOO_LARGE_MEM_ERR
)

// keccakState wraps sha3.state. In addition to the usual hash methods, it also supports
// Read to get a variable amount of data from the hash state. Read is faster than Sum
// because it doesn't copy the internal state, but also modifies the internal state.
type keccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}

type interpreter struct {
	evm       *evm
	jt        *jump_table
	g_jt      *jump_table
	lp_jt     *jump_table
	hasher    keccakState // Keccak256 hasher instance shared across opcodes
	hasherBuf common.Hash // Keccak256 hasher result array shared across opcodes
}

type callCtx struct {
	memory   *Memory
	stack    *Stack
	p_stack  *p_stack
	contract *Contract
}

func (ctx *callCtx) copy() *callCtx {
	new_ctx := &callCtx{
		memory:   ctx.memory._copy(),
		stack:    ctx.stack._copy(),
		p_stack:  ctx.p_stack,
		contract: ctx.contract,
	}

	return new_ctx
}

func new_interpreter(_evm *evm) *interpreter {
	jt := new_jt()
	g_jt := new_graph_jt()
	lp_jt := new_loop_jt()
	in := &interpreter{
		evm:   _evm,
		jt:    &jt,
		g_jt:  &g_jt,
		lp_jt: &lp_jt,
	}
	return in
}

func (in *interpreter) g_run(pc *uint64, callCtx *callCtx) (uint64, bool) {

	bytecode := callCtx.contract.Code
	stack := callCtx.stack
	memory := callCtx.memory

	l := uint64(len(bytecode))

	for *pc < l {

		op := bytecode[*pc]
		operation := in.g_jt[op]

		if operation == nil || op == INVALID {
			// invalid operation
			return INVALID_OP, false
		}

		if stack_len := stack.Len(); stack_len < operation.min_stack {
			// stack underflow
			return STACK_UNDERFLOW, false
		} else if stack_len > operation.max_stack {
			// stack overflow
			return STACK_OVERFLOW, false
		}

		var memory_size uint64
		if operation.mem_size != nil {
			mem_size, overflow := operation.mem_size(stack)
			if overflow {
				// gas unit overflow
				return GAS_UNIT_OVERFLOW, false
			}

			if memory_size, overflow = math.SafeMul(toWordSize(mem_size), 32); overflow {
				// gas unit overflow
				return GAS_UNIT_OVERFLOW, false
			}
		}

		if memory_size > 0 {
			if memory_size > GAS_CONST {
				return GAS_CONST_ERR, false
			}
			if memory_size > MAX_MEM_SIZE {
				return TOO_LARGE_MEM_ERR, false
			}
			memory.Resize(memory_size)
		}

		jumpdest := operation.execute(pc, in, callCtx)

		switch {
		case operation.jumps:
			return jumpdest, true
		case operation.reverts:
			return REVERTS, false
		case operation.halts:
			return HALTS, false
		case !operation.jumps:
			*pc++
		}
	}

	return END_OF_LOOP, false
}

func (in *interpreter) run(pc *uint64, ctx *callCtx, bytecode *[]byte, code_size *uint64) (uint64, bool) {

	stack := ctx.stack

	for *pc < *code_size && !in.evm.abort {

		op := (*bytecode)[*pc]
		operation := in.jt[op]

		if operation == nil || op == INVALID {
			return INVALID_OP, false
		}

		if stack_len := stack.Len(); stack_len < operation.min_stack {
			return STACK_UNDERFLOW, false
		} else if stack_len > operation.max_stack {
			return STACK_OVERFLOW, false
		}

		var memory_size uint64
		if operation.mem_size != nil {
			mem_size, overflow := operation.mem_size(stack)
			if overflow {
				return GAS_UNIT_OVERFLOW, false
			}

			if memory_size, overflow = math.SafeMul(toWordSize(mem_size), 32); overflow {
				return GAS_UNIT_OVERFLOW, false
			}
		}

		if memory_size > 0 {
			if memory_size > GAS_CONST {
				return GAS_CONST_ERR, false
			}
			if memory_size > MAX_MEM_SIZE {
				return TOO_LARGE_MEM_ERR, false
			}
			ctx.memory.Resize(memory_size)
		}

		jump_dest := operation.execute(pc, in, ctx)

		switch {
		case operation.jumps:
			return jump_dest, true
		case operation.reverts:
			return REVERTS, false
		case operation.halts:
			return HALTS, false
		case !operation.jumps:
			*pc++
		}

	}

	return END_OF_LOOP, false
}

func (in *interpreter) lp_run(pc *uint64, ctx *callCtx, bytecode *[]byte, code_size *uint64) uint64 {
	stack := ctx.stack
	for *pc < *code_size {
		op := (*bytecode)[*pc]
		operation := in.lp_jt[op]

		if operation == nil || op == INVALID {
			return INVALID_OP
		}

		if stack_len := stack.Len(); stack_len < operation.min_stack {
			return STACK_UNDERFLOW
		} else if stack_len > operation.max_stack {
			return STACK_OVERFLOW
		}

		var memory_size uint64
		if operation.mem_size != nil {
			mem_size, overflow := operation.mem_size(stack)
			if overflow {
				return GAS_UNIT_OVERFLOW
			}

			if memory_size, overflow = math.SafeMul(toWordSize(mem_size), 32); overflow {
				return GAS_UNIT_OVERFLOW
			}
		}

		if memory_size > 0 {
			if memory_size > GAS_CONST {
				return GAS_CONST_ERR
			}
			if memory_size > MAX_MEM_SIZE {
				return TOO_LARGE_MEM_ERR
			}
			ctx.memory.Resize(memory_size)
		}

		jump_dest := operation.execute(pc, in, ctx)

		switch {
		case operation.jumps:
			return jump_dest
		case operation.reverts:
			return REVERTS
		case operation.halts:
			return HALTS
		case !operation.jumps:
			*pc++
		}
	}

	return END_OF_LOOP
}
