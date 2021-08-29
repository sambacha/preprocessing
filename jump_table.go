package main

type (
	exec_func     func(pc *uint64, in *interpreter, ctx *callCtx) uint64
	mem_size_func func(*Stack) (size uint64, overflow bool)
)

type operation struct {
	execute  exec_func
	mem_size mem_size_func

	min_stack int
	max_stack int
	num_pop   int
	num_push  int

	halts   bool
	jumps   bool
	writes  bool
	reverts bool
	returns bool

	// is_push   bool
	// is_swap   bool
	// is_dup    bool
}

type jump_table [256]*operation

func new_op(exec exec_func, num_pop, num_push int) *operation {
	return &operation{
		execute:   exec,
		min_stack: minStack(num_pop, num_push),
		max_stack: maxStack(num_pop, num_push),
		num_pop:   num_pop,
		num_push:  num_push,
	}
}

func new_dup_op(exec exec_func, n int) *operation {
	return &operation{
		execute:   exec,
		min_stack: minDupStack(n),
		max_stack: maxDupStack(n),
		num_pop:   0,
		num_push:  1,
	}
}

func new_swap_op(exec exec_func, n int) *operation {
	return &operation{
		execute:   exec,
		min_stack: minSwapStack(n),
		max_stack: maxSwapStack(n),
		num_pop:   1,
		num_push:  1,
	}
}

func (op *operation) with_mem(mem_size mem_size_func) *operation {
	op.mem_size = mem_size
	return op
}

func (op *operation) _halts() *operation {
	op.halts = true
	return op
}

func (op *operation) _jumps() *operation {
	op.jumps = true
	return op
}

func (op *operation) _writes() *operation {
	op.writes = true
	return op
}

func (op *operation) _reverts() *operation {
	op.reverts = true
	return op
}

func (op *operation) _returns() *operation {
	op.returns = true
	return op
}

func new_jt() jump_table {
	return jump_table{
		/* 0s: Stop and Arithmetic Operations */
		STOP:       new_op(op_STOP, 0, 0)._halts(),
		ADD:        new_op(op_ADD, 2, 1),
		MUL:        new_op(op_MUL, 2, 1),
		SUB:        new_op(op_SUB, 2, 1),
		DIV:        new_op(op_DIV, 2, 1),
		SDIV:       new_op(op_SDIV, 2, 1),
		MOD:        new_op(op_MOD, 2, 1),
		SMOD:       new_op(op_SMOD, 2, 1),
		ADDMOD:     new_op(op_ADDMOD, 3, 1),
		MULMOD:     new_op(op_MULMOD, 3, 1),
		EXP:        new_op(op_EXP, 2, 1),
		SIGNEXTEND: new_op(op_SIGNEXTEND, 2, 1),

		/* 10s: Comparison & Bitwise Logic Operations */
		LT:     new_op(op_LT, 2, 1),
		GT:     new_op(op_GT, 2, 1),
		SLT:    new_op(op_SLT, 2, 1),
		SGT:    new_op(op_SGT, 2, 1),
		EQ:     new_op(op_EQ, 2, 1),
		ISZERO: new_op(op_ISZERO, 1, 1),
		AND:    new_op(op_AND, 2, 1),
		OR:     new_op(op_OR, 2, 1),
		XOR:    new_op(op_XOR, 2, 1),
		NOT:    new_op(op_NOT, 2, 1),
		BYTE:   new_op(op_BYTE, 2, 1),
		SHL:    new_op(op_SHL, 2, 1),
		SHR:    new_op(op_SHR, 2, 1),
		SAR:    new_op(op_SAR, 2, 1),

		/* 20s: SHA3 */
		SHA3: new_op(op_SAR, 2, 1).with_mem(memorySha3),

		/* 30s: Environmental Information */
		ADDRESS:        new_op(op_ADDRESS, 0, 1),
		BALANCE:        new_op(op_BALANCE, 1, 1),
		ORIGIN:         new_op(op_ORIGIN, 0, 1),
		CALLER:         new_op(op_CALLER, 0, 1),
		CALLVALUE:      new_op(op_CALLVALUE, 0, 1),
		CALLDATALOAD:   new_op(op_CALLDATALOAD, 1, 1),
		CALLDATASIZE:   new_op(op_CALLDATASIZE, 0, 1),
		CALLDATACOPY:   new_op(op_CALLDATACOPY, 3, 0).with_mem(memoryCallDataCopy),
		CODESIZE:       new_op(op_CODESIZE, 0, 1),
		CODECOPY:       new_op(op_CODECOPY, 3, 0).with_mem(memoryCodeCopy),
		GASPRICE:       new_op(op_GASPRICE, 0, 1),
		EXTCODESIZE:    new_op(op_EXTCODESIZE, 1, 1),
		EXTCODECOPY:    new_op(op_EXTCODECOPY, 4, 0).with_mem(memoryExtCodeCopy),
		RETURNDATASIZE: new_op(op_RETURNDATASIZE, 0, 1),
		RETURNDATACOPY: new_op(op_RETURNDATACOPY, 3, 0).with_mem(memoryReturnDataCopy),
		EXTCODEHASH:    new_op(op_EXTCODEHASH, 1, 1),

		/* 40s: Block Information */
		BLOCKHASH:   new_op(op_BLOCKHASH, 1, 1),
		COINBASE:    new_op(op_COINBASE, 0, 1),
		TIMESTAMP:   new_op(op_TIMESTAMP, 0, 1),
		NUMBER:      new_op(op_NUMBER, 0, 1),
		DIFFICULTY:  new_op(op_DIFFICULTY, 0, 1),
		GASLIMIT:    new_op(op_GASLIMIT, 0, 1),
		CHAINID:     new_op(op_CHAINID, 0, 1),
		SELFBALANCE: new_op(op_SELFBALANCE, 0, 1),

		/* 50s: Stack, Memory, Storage and Flow Operations */
		POP:      new_op(op_POP, 1, 0),
		MLOAD:    new_op(op_MLOAD, 1, 1).with_mem(memoryMLoad),
		MSTORE:   new_op(op_MSTORE, 2, 0).with_mem(memoryMStore),
		MSTORE8:  new_op(op_MSTORE8, 2, 0).with_mem(memoryMStore8),
		SLOAD:    new_op(op_SLOAD, 1, 1),
		SSTORE:   new_op(op_SSTORE, 2, 0)._writes(),
		JUMP:     new_op(op_JUMP, 1, 0)._jumps(),
		JUMPI:    new_op(op_JUMPI, 2, 0)._jumps(),
		PC:       new_op(op_PC, 0, 1),
		MSIZE:    new_op(op_MSIZE, 0, 1),
		GAS:      new_op(op_GAS, 0, 1),
		JUMPDEST: new_op(op_JUMPDEST, 0, 0),

		/* 60s & 70s: Push Operations */
		PUSH1:  new_op(op_PUSH1, 0, 1),
		PUSH2:  new_op(makePush(2, 2), 0, 1),
		PUSH3:  new_op(makePush(3, 3), 0, 1),
		PUSH4:  new_op(makePush(4, 4), 0, 1),
		PUSH5:  new_op(makePush(5, 5), 0, 1),
		PUSH6:  new_op(makePush(6, 6), 0, 1),
		PUSH7:  new_op(makePush(7, 7), 0, 1),
		PUSH8:  new_op(makePush(8, 8), 0, 1),
		PUSH9:  new_op(makePush(9, 9), 0, 1),
		PUSH10: new_op(makePush(10, 10), 0, 1),
		PUSH11: new_op(makePush(11, 11), 0, 1),
		PUSH12: new_op(makePush(12, 12), 0, 1),
		PUSH13: new_op(makePush(13, 13), 0, 1),
		PUSH14: new_op(makePush(14, 14), 0, 1),
		PUSH15: new_op(makePush(15, 15), 0, 1),
		PUSH16: new_op(makePush(16, 16), 0, 1),
		PUSH17: new_op(makePush(17, 17), 0, 1),
		PUSH18: new_op(makePush(18, 18), 0, 1),
		PUSH19: new_op(makePush(19, 19), 0, 1),
		PUSH20: new_op(makePush(20, 20), 0, 1),
		PUSH21: new_op(makePush(21, 21), 0, 1),
		PUSH22: new_op(makePush(22, 22), 0, 1),
		PUSH23: new_op(makePush(23, 23), 0, 1),
		PUSH24: new_op(makePush(24, 24), 0, 1),
		PUSH25: new_op(makePush(25, 25), 0, 1),
		PUSH26: new_op(makePush(26, 26), 0, 1),
		PUSH27: new_op(makePush(27, 27), 0, 1),
		PUSH28: new_op(makePush(28, 28), 0, 1),
		PUSH29: new_op(makePush(29, 29), 0, 1),
		PUSH30: new_op(makePush(30, 30), 0, 1),
		PUSH31: new_op(makePush(31, 31), 0, 1),
		PUSH32: new_op(makePush(32, 32), 0, 1),

		/* 80s: Duplication Operations */
		DUP1:  new_dup_op(makeDup(1), 1),
		DUP2:  new_dup_op(makeDup(2), 2),
		DUP3:  new_dup_op(makeDup(3), 3),
		DUP4:  new_dup_op(makeDup(4), 4),
		DUP5:  new_dup_op(makeDup(5), 5),
		DUP6:  new_dup_op(makeDup(6), 6),
		DUP7:  new_dup_op(makeDup(7), 7),
		DUP8:  new_dup_op(makeDup(8), 8),
		DUP9:  new_dup_op(makeDup(9), 9),
		DUP10: new_dup_op(makeDup(10), 10),
		DUP11: new_dup_op(makeDup(11), 11),
		DUP12: new_dup_op(makeDup(12), 12),
		DUP13: new_dup_op(makeDup(13), 13),
		DUP14: new_dup_op(makeDup(14), 14),
		DUP15: new_dup_op(makeDup(15), 15),
		DUP16: new_dup_op(makeDup(16), 16),

		/* 90s: Exchange Operations */
		SWAP1:  new_swap_op(makeSwap(1), 2),
		SWAP2:  new_swap_op(makeSwap(2), 3),
		SWAP3:  new_swap_op(makeSwap(3), 4),
		SWAP4:  new_swap_op(makeSwap(4), 5),
		SWAP5:  new_swap_op(makeSwap(5), 6),
		SWAP6:  new_swap_op(makeSwap(6), 7),
		SWAP7:  new_swap_op(makeSwap(7), 8),
		SWAP8:  new_swap_op(makeSwap(8), 9),
		SWAP9:  new_swap_op(makeSwap(9), 10),
		SWAP10: new_swap_op(makeSwap(10), 11),
		SWAP11: new_swap_op(makeSwap(11), 12),
		SWAP12: new_swap_op(makeSwap(12), 13),
		SWAP13: new_swap_op(makeSwap(13), 14),
		SWAP14: new_swap_op(makeSwap(14), 15),
		SWAP15: new_swap_op(makeSwap(15), 16),
		SWAP16: new_swap_op(makeSwap(16), 17),

		/* a0s: Logging Operations */
		LOG0: new_op(makeLog(0), 2, 0).with_mem(memoryLog)._writes(),
		LOG1: new_op(makeLog(1), 3, 0).with_mem(memoryLog)._writes(),
		LOG2: new_op(makeLog(2), 4, 0).with_mem(memoryLog)._writes(),
		LOG3: new_op(makeLog(3), 5, 0).with_mem(memoryLog)._writes(),
		LOG4: new_op(makeLog(4), 6, 0).with_mem(memoryLog)._writes(),

		/* f0s: System operations */
		CREATE: new_op(op_CREATE, 3, 1).with_mem(memoryCreate)._writes()._returns(),

		CALL:         new_op(op_CALL, 7, 1).with_mem(memoryCall)._returns(),
		CALLCODE:     new_op(op_CALLCODE, 7, 1).with_mem(memoryCall)._returns(),
		RETURN:       new_op(op_RETURN, 2, 0).with_mem(memoryReturn)._halts(),
		DELEGATECALL: new_op(op_DELEGATECALL, 6, 1).with_mem(memoryDelegateCall)._returns(),

		CREATE2: new_op(op_CREATE2, 4, 1).with_mem(memoryCreate2)._writes()._returns(),

		REVERT: new_op(op_REVERT, 2, 0).with_mem(memoryRevert)._reverts()._returns(),

		INVALID:      new_op(op_INVALID, 0, 0),
		SELFDESTRUCT: new_op(op_SELFDESTRUCT, 1, 0)._writes()._halts(),
	}
}

func new_graph_jt() jump_table {
	return jump_table{
		/* 0s: Stop and Arithmetic Operations */
		STOP:       new_op(g_STOP, 0, 0)._halts(),
		ADD:        new_op(g_ADD, 2, 1),
		MUL:        new_op(g_MUL, 2, 1),
		SUB:        new_op(g_SUB, 2, 1),
		DIV:        new_op(g_DIV, 2, 1),
		SDIV:       new_op(g_SDIV, 2, 1),
		MOD:        new_op(g_MOD, 2, 1),
		SMOD:       new_op(g_SMOD, 2, 1),
		ADDMOD:     new_op(g_ADDMOD, 3, 1),
		MULMOD:     new_op(g_MULMOD, 3, 1),
		EXP:        new_op(g_EXP, 2, 1),
		SIGNEXTEND: new_op(g_SIGNEXTEND, 2, 1),

		/* 10s: Comparison & Bitwise Logic Operations */
		LT:     new_op(g_LT, 2, 1),
		GT:     new_op(g_GT, 2, 1),
		SLT:    new_op(g_SLT, 2, 1),
		SGT:    new_op(g_SGT, 2, 1),
		EQ:     new_op(g_EQ, 2, 1),
		ISZERO: new_op(g_ISZERO, 1, 1),
		AND:    new_op(g_AND, 2, 1),
		OR:     new_op(g_OR, 2, 1),
		XOR:    new_op(g_XOR, 2, 1),
		NOT:    new_op(g_NOT, 2, 1),
		BYTE:   new_op(g_BYTE, 2, 1),
		SHL:    new_op(g_SHL, 2, 1),
		SHR:    new_op(g_SHR, 2, 1),
		SAR:    new_op(g_SAR, 2, 1),

		/* 20s: SHA3 */
		SHA3: new_op(g_SAR, 2, 1).with_mem(memorySha3),

		/* 30s: Environmental Information */
		ADDRESS:        new_op(g_ADDRESS, 0, 1),
		BALANCE:        new_op(g_BALANCE, 1, 1),
		ORIGIN:         new_op(g_ORIGIN, 0, 1),
		CALLER:         new_op(g_CALLER, 0, 1),
		CALLVALUE:      new_op(g_CALLVALUE, 0, 1),
		CALLDATALOAD:   new_op(g_CALLDATALOAD, 1, 1),
		CALLDATASIZE:   new_op(g_CALLDATASIZE, 0, 1),
		CALLDATACOPY:   new_op(g_CALLDATACOPY, 3, 0).with_mem(memoryCallDataCopy),
		CODESIZE:       new_op(g_CODESIZE, 0, 1),
		CODECOPY:       new_op(g_CODECOPY, 3, 0).with_mem(memoryCodeCopy),
		GASPRICE:       new_op(g_GASPRICE, 0, 1),
		EXTCODESIZE:    new_op(g_EXTCODESIZE, 1, 1),
		EXTCODECOPY:    new_op(g_EXTCODECOPY, 4, 0).with_mem(memoryExtCodeCopy),
		RETURNDATASIZE: new_op(g_RETURNDATASIZE, 0, 1),
		RETURNDATACOPY: new_op(g_RETURNDATACOPY, 3, 0).with_mem(memoryReturnDataCopy),
		EXTCODEHASH:    new_op(g_EXTCODEHASH, 1, 1),

		/* 40s: Block Information */
		BLOCKHASH:   new_op(g_BLOCKHASH, 1, 1),
		COINBASE:    new_op(g_COINBASE, 0, 1),
		TIMESTAMP:   new_op(g_TIMESTAMP, 0, 1),
		NUMBER:      new_op(g_NUMBER, 0, 1),
		DIFFICULTY:  new_op(g_DIFFICULTY, 0, 1),
		GASLIMIT:    new_op(g_GASLIMIT, 0, 1),
		CHAINID:     new_op(g_CHAINID, 0, 1),
		SELFBALANCE: new_op(g_SELFBALANCE, 0, 1),

		/* 50s: Stack, Memory, Storage and Flow Operations */
		POP:      new_op(g_POP, 1, 0),
		MLOAD:    new_op(g_MLOAD, 1, 1).with_mem(memoryMLoad),
		MSTORE:   new_op(g_MSTORE, 2, 0).with_mem(memoryMStore),
		MSTORE8:  new_op(g_MSTORE8, 2, 0).with_mem(memoryMStore8),
		SLOAD:    new_op(g_SLOAD, 1, 1),
		SSTORE:   new_op(g_SSTORE, 2, 0)._writes(),
		JUMP:     new_op(g_JUMP, 1, 0)._jumps(),
		JUMPI:    new_op(g_JUMPI, 2, 0)._jumps(),
		PC:       new_op(g_PC, 0, 1),
		MSIZE:    new_op(g_MSIZE, 0, 1),
		GAS:      new_op(g_GAS, 0, 1),
		JUMPDEST: new_op(g_JUMPDEST, 0, 0),

		/* 60s & 70s: Push Operations */
		PUSH1:  new_op(op_PUSH1, 0, 1),
		PUSH2:  new_op(g_makePush(2, 2), 0, 1),
		PUSH3:  new_op(g_makePush(3, 3), 0, 1),
		PUSH4:  new_op(g_makePush(4, 4), 0, 1),
		PUSH5:  new_op(g_makePush(5, 5), 0, 1),
		PUSH6:  new_op(g_makePush(6, 6), 0, 1),
		PUSH7:  new_op(g_makePush(7, 7), 0, 1),
		PUSH8:  new_op(g_makePush(8, 8), 0, 1),
		PUSH9:  new_op(g_makePush(9, 9), 0, 1),
		PUSH10: new_op(g_makePush(10, 10), 0, 1),
		PUSH11: new_op(g_makePush(11, 11), 0, 1),
		PUSH12: new_op(g_makePush(12, 12), 0, 1),
		PUSH13: new_op(g_makePush(13, 13), 0, 1),
		PUSH14: new_op(g_makePush(14, 14), 0, 1),
		PUSH15: new_op(g_makePush(15, 15), 0, 1),
		PUSH16: new_op(g_makePush(16, 16), 0, 1),
		PUSH17: new_op(g_makePush(17, 17), 0, 1),
		PUSH18: new_op(g_makePush(18, 18), 0, 1),
		PUSH19: new_op(g_makePush(19, 19), 0, 1),
		PUSH20: new_op(g_makePush(20, 20), 0, 1),
		PUSH21: new_op(g_makePush(21, 21), 0, 1),
		PUSH22: new_op(g_makePush(22, 22), 0, 1),
		PUSH23: new_op(g_makePush(23, 23), 0, 1),
		PUSH24: new_op(g_makePush(24, 24), 0, 1),
		PUSH25: new_op(g_makePush(25, 25), 0, 1),
		PUSH26: new_op(g_makePush(26, 26), 0, 1),
		PUSH27: new_op(g_makePush(27, 27), 0, 1),
		PUSH28: new_op(g_makePush(28, 28), 0, 1),
		PUSH29: new_op(g_makePush(29, 29), 0, 1),
		PUSH30: new_op(g_makePush(30, 30), 0, 1),
		PUSH31: new_op(g_makePush(31, 31), 0, 1),
		PUSH32: new_op(g_makePush(32, 32), 0, 1),

		/* 80s: Duplication Operations */
		DUP1:  new_dup_op(g_makeDup(1), 1),
		DUP2:  new_dup_op(g_makeDup(2), 2),
		DUP3:  new_dup_op(g_makeDup(3), 3),
		DUP4:  new_dup_op(g_makeDup(4), 4),
		DUP5:  new_dup_op(g_makeDup(5), 5),
		DUP6:  new_dup_op(g_makeDup(6), 6),
		DUP7:  new_dup_op(g_makeDup(7), 7),
		DUP8:  new_dup_op(g_makeDup(8), 8),
		DUP9:  new_dup_op(g_makeDup(9), 9),
		DUP10: new_dup_op(g_makeDup(10), 10),
		DUP11: new_dup_op(g_makeDup(11), 11),
		DUP12: new_dup_op(g_makeDup(12), 12),
		DUP13: new_dup_op(g_makeDup(13), 13),
		DUP14: new_dup_op(g_makeDup(14), 14),
		DUP15: new_dup_op(g_makeDup(15), 15),
		DUP16: new_dup_op(g_makeDup(16), 16),

		/* 90s: Exchange Operations */
		SWAP1:  new_swap_op(g_makeSwap(1), 2),
		SWAP2:  new_swap_op(g_makeSwap(2), 3),
		SWAP3:  new_swap_op(g_makeSwap(3), 4),
		SWAP4:  new_swap_op(g_makeSwap(4), 5),
		SWAP5:  new_swap_op(g_makeSwap(5), 6),
		SWAP6:  new_swap_op(g_makeSwap(6), 7),
		SWAP7:  new_swap_op(g_makeSwap(7), 8),
		SWAP8:  new_swap_op(g_makeSwap(8), 9),
		SWAP9:  new_swap_op(g_makeSwap(9), 10),
		SWAP10: new_swap_op(g_makeSwap(10), 11),
		SWAP11: new_swap_op(g_makeSwap(11), 12),
		SWAP12: new_swap_op(g_makeSwap(12), 13),
		SWAP13: new_swap_op(g_makeSwap(13), 14),
		SWAP14: new_swap_op(g_makeSwap(14), 15),
		SWAP15: new_swap_op(g_makeSwap(15), 16),
		SWAP16: new_swap_op(g_makeSwap(16), 17),

		/* a0s: Logging Operations */
		LOG0: new_op(g_makeLog(0), 2, 0).with_mem(memoryLog)._writes(),
		LOG1: new_op(g_makeLog(1), 3, 0).with_mem(memoryLog)._writes(),
		LOG2: new_op(g_makeLog(2), 4, 0).with_mem(memoryLog)._writes(),
		LOG3: new_op(g_makeLog(3), 5, 0).with_mem(memoryLog)._writes(),
		LOG4: new_op(g_makeLog(4), 6, 0).with_mem(memoryLog)._writes(),

		/* f0s: System operations */
		CREATE: new_op(g_CREATE, 3, 1).with_mem(memoryCreate)._writes()._returns(),

		CALL:         new_op(g_CALL, 7, 1).with_mem(memoryCall)._returns(),
		CALLCODE:     new_op(g_CALLCODE, 7, 1).with_mem(memoryCall)._returns(),
		RETURN:       new_op(g_RETURN, 2, 0).with_mem(memoryReturn)._halts(),
		DELEGATECALL: new_op(g_DELEGATECALL, 6, 1).with_mem(memoryDelegateCall)._returns(),

		CREATE2: new_op(g_CREATE2, 4, 1).with_mem(memoryCreate2)._writes()._returns(),

		REVERT: new_op(g_REVERT, 2, 0).with_mem(memoryRevert)._reverts()._returns(),

		INVALID:      new_op(g_INVALID, 0, 0),
		SELFDESTRUCT: new_op(g_SELFDESTRUCT, 1, 0)._writes()._halts(),
	}
}

func new_loop_jt() jump_table {
	return jump_table{
		/* 0s: Stop and Arithmetic Operations */
		STOP:       new_op(lp_STOP, 0, 0)._halts(),
		ADD:        new_op(lp_ADD, 2, 1),
		MUL:        new_op(lp_MUL, 2, 1),
		SUB:        new_op(lp_SUB, 2, 1),
		DIV:        new_op(lp_DIV, 2, 1),
		SDIV:       new_op(lp_SDIV, 2, 1),
		MOD:        new_op(lp_MOD, 2, 1),
		SMOD:       new_op(lp_SMOD, 2, 1),
		ADDMOD:     new_op(lp_ADDMOD, 3, 1),
		MULMOD:     new_op(lp_MULMOD, 3, 1),
		EXP:        new_op(lp_EXP, 2, 1),
		SIGNEXTEND: new_op(lp_SIGNEXTEND, 2, 1),

		/* 10s: Comparison & Bitwise Logic Operations */
		LT:     new_op(lp_LT, 2, 1),
		GT:     new_op(lp_GT, 2, 1),
		SLT:    new_op(lp_SLT, 2, 1),
		SGT:    new_op(lp_SGT, 2, 1),
		EQ:     new_op(lp_EQ, 2, 1),
		ISZERO: new_op(lp_ISZERO, 1, 1),
		AND:    new_op(lp_AND, 2, 1),
		OR:     new_op(lp_OR, 2, 1),
		XOR:    new_op(lp_XOR, 2, 1),
		NOT:    new_op(lp_NOT, 2, 1),
		BYTE:   new_op(lp_BYTE, 2, 1),
		SHL:    new_op(lp_SHL, 2, 1),
		SHR:    new_op(lp_SHR, 2, 1),
		SAR:    new_op(lp_SAR, 2, 1),

		/* 20s: SHA3 */
		SHA3: new_op(lp_SAR, 2, 1).with_mem(memorySha3),

		/* 30s: Environmental Information */
		ADDRESS:        new_op(lp_ADDRESS, 0, 1),
		BALANCE:        new_op(lp_BALANCE, 1, 1),
		ORIGIN:         new_op(lp_ORIGIN, 0, 1),
		CALLER:         new_op(lp_CALLER, 0, 1),
		CALLVALUE:      new_op(lp_CALLVALUE, 0, 1),
		CALLDATALOAD:   new_op(lp_CALLDATALOAD, 1, 1),
		CALLDATASIZE:   new_op(lp_CALLDATASIZE, 0, 1),
		CALLDATACOPY:   new_op(lp_CALLDATACOPY, 3, 0).with_mem(memoryCallDataCopy),
		CODESIZE:       new_op(lp_CODESIZE, 0, 1),
		CODECOPY:       new_op(lp_CODECOPY, 3, 0).with_mem(memoryCodeCopy),
		GASPRICE:       new_op(lp_GASPRICE, 0, 1),
		EXTCODESIZE:    new_op(lp_EXTCODESIZE, 1, 1),
		EXTCODECOPY:    new_op(lp_EXTCODECOPY, 4, 0).with_mem(memoryExtCodeCopy),
		RETURNDATASIZE: new_op(lp_RETURNDATASIZE, 0, 1),
		RETURNDATACOPY: new_op(lp_RETURNDATACOPY, 3, 0).with_mem(memoryReturnDataCopy),
		EXTCODEHASH:    new_op(lp_EXTCODEHASH, 1, 1),

		/* 40s: Block Information */
		BLOCKHASH:   new_op(lp_BLOCKHASH, 1, 1),
		COINBASE:    new_op(lp_COINBASE, 0, 1),
		TIMESTAMP:   new_op(lp_TIMESTAMP, 0, 1),
		NUMBER:      new_op(lp_NUMBER, 0, 1),
		DIFFICULTY:  new_op(lp_DIFFICULTY, 0, 1),
		GASLIMIT:    new_op(lp_GASLIMIT, 0, 1),
		CHAINID:     new_op(lp_CHAINID, 0, 1),
		SELFBALANCE: new_op(lp_SELFBALANCE, 0, 1),

		/* 50s: Stack, Memory, Storage and Flow Operations */
		POP:      new_op(lp_POP, 1, 0),
		MLOAD:    new_op(lp_MLOAD, 1, 1).with_mem(memoryMLoad),
		MSTORE:   new_op(lp_MSTORE, 2, 0).with_mem(memoryMStore),
		MSTORE8:  new_op(lp_MSTORE8, 2, 0).with_mem(memoryMStore8),
		SLOAD:    new_op(lp_SLOAD, 1, 1),
		SSTORE:   new_op(lp_SSTORE, 2, 0)._writes(),
		JUMP:     new_op(lp_JUMP, 1, 0)._jumps(),
		JUMPI:    new_op(lp_JUMPI, 2, 0)._jumps(),
		PC:       new_op(lp_PC, 0, 1),
		MSIZE:    new_op(lp_MSIZE, 0, 1),
		GAS:      new_op(lp_GAS, 0, 1),
		JUMPDEST: new_op(lp_JUMPDEST, 0, 0),

		/* 60s & 70s: Push Operations */
		PUSH1:  new_op(lp_PUSH1, 0, 1),
		PUSH2:  new_op(lp_makePush(2, 2), 0, 1),
		PUSH3:  new_op(lp_makePush(3, 3), 0, 1),
		PUSH4:  new_op(lp_makePush(4, 4), 0, 1),
		PUSH5:  new_op(lp_makePush(5, 5), 0, 1),
		PUSH6:  new_op(lp_makePush(6, 6), 0, 1),
		PUSH7:  new_op(lp_makePush(7, 7), 0, 1),
		PUSH8:  new_op(lp_makePush(8, 8), 0, 1),
		PUSH9:  new_op(lp_makePush(9, 9), 0, 1),
		PUSH10: new_op(lp_makePush(10, 10), 0, 1),
		PUSH11: new_op(lp_makePush(11, 11), 0, 1),
		PUSH12: new_op(lp_makePush(12, 12), 0, 1),
		PUSH13: new_op(lp_makePush(13, 13), 0, 1),
		PUSH14: new_op(lp_makePush(14, 14), 0, 1),
		PUSH15: new_op(lp_makePush(15, 15), 0, 1),
		PUSH16: new_op(lp_makePush(16, 16), 0, 1),
		PUSH17: new_op(lp_makePush(17, 17), 0, 1),
		PUSH18: new_op(lp_makePush(18, 18), 0, 1),
		PUSH19: new_op(lp_makePush(19, 19), 0, 1),
		PUSH20: new_op(lp_makePush(20, 20), 0, 1),
		PUSH21: new_op(lp_makePush(21, 21), 0, 1),
		PUSH22: new_op(lp_makePush(22, 22), 0, 1),
		PUSH23: new_op(lp_makePush(23, 23), 0, 1),
		PUSH24: new_op(lp_makePush(24, 24), 0, 1),
		PUSH25: new_op(lp_makePush(25, 25), 0, 1),
		PUSH26: new_op(lp_makePush(26, 26), 0, 1),
		PUSH27: new_op(lp_makePush(27, 27), 0, 1),
		PUSH28: new_op(lp_makePush(28, 28), 0, 1),
		PUSH29: new_op(lp_makePush(29, 29), 0, 1),
		PUSH30: new_op(lp_makePush(30, 30), 0, 1),
		PUSH31: new_op(lp_makePush(31, 31), 0, 1),
		PUSH32: new_op(lp_makePush(32, 32), 0, 1),

		/* 80s: Duplication Operations */
		DUP1:  new_dup_op(lp_makeDup(1), 1),
		DUP2:  new_dup_op(lp_makeDup(2), 2),
		DUP3:  new_dup_op(lp_makeDup(3), 3),
		DUP4:  new_dup_op(lp_makeDup(4), 4),
		DUP5:  new_dup_op(lp_makeDup(5), 5),
		DUP6:  new_dup_op(lp_makeDup(6), 6),
		DUP7:  new_dup_op(lp_makeDup(7), 7),
		DUP8:  new_dup_op(lp_makeDup(8), 8),
		DUP9:  new_dup_op(lp_makeDup(9), 9),
		DUP10: new_dup_op(lp_makeDup(10), 10),
		DUP11: new_dup_op(lp_makeDup(11), 11),
		DUP12: new_dup_op(lp_makeDup(12), 12),
		DUP13: new_dup_op(lp_makeDup(13), 13),
		DUP14: new_dup_op(lp_makeDup(14), 14),
		DUP15: new_dup_op(lp_makeDup(15), 15),
		DUP16: new_dup_op(lp_makeDup(16), 16),

		/* 90s: Exchange Operations */
		SWAP1:  new_swap_op(lp_makeSwap(1), 2),
		SWAP2:  new_swap_op(lp_makeSwap(2), 3),
		SWAP3:  new_swap_op(lp_makeSwap(3), 4),
		SWAP4:  new_swap_op(lp_makeSwap(4), 5),
		SWAP5:  new_swap_op(lp_makeSwap(5), 6),
		SWAP6:  new_swap_op(lp_makeSwap(6), 7),
		SWAP7:  new_swap_op(lp_makeSwap(7), 8),
		SWAP8:  new_swap_op(lp_makeSwap(8), 9),
		SWAP9:  new_swap_op(lp_makeSwap(9), 10),
		SWAP10: new_swap_op(lp_makeSwap(10), 11),
		SWAP11: new_swap_op(lp_makeSwap(11), 12),
		SWAP12: new_swap_op(lp_makeSwap(12), 13),
		SWAP13: new_swap_op(lp_makeSwap(13), 14),
		SWAP14: new_swap_op(lp_makeSwap(14), 15),
		SWAP15: new_swap_op(lp_makeSwap(15), 16),
		SWAP16: new_swap_op(lp_makeSwap(16), 17),

		/* a0s: Logging Operations */
		LOG0: new_op(lp_makeLog(0), 2, 0).with_mem(memoryLog)._writes(),
		LOG1: new_op(lp_makeLog(1), 3, 0).with_mem(memoryLog)._writes(),
		LOG2: new_op(lp_makeLog(2), 4, 0).with_mem(memoryLog)._writes(),
		LOG3: new_op(lp_makeLog(3), 5, 0).with_mem(memoryLog)._writes(),
		LOG4: new_op(lp_makeLog(4), 6, 0).with_mem(memoryLog)._writes(),

		/* f0s: System operations */
		CREATE: new_op(lp_CREATE, 3, 1).with_mem(memoryCreate)._writes()._returns(),

		CALL:         new_op(lp_CALL, 7, 1).with_mem(memoryCall)._returns(),
		CALLCODE:     new_op(lp_CALLCODE, 7, 1).with_mem(memoryCall)._returns(),
		RETURN:       new_op(lp_RETURN, 2, 0).with_mem(memoryReturn)._halts(),
		DELEGATECALL: new_op(lp_DELEGATECALL, 6, 1).with_mem(memoryDelegateCall)._returns(),

		CREATE2: new_op(lp_CREATE2, 4, 1).with_mem(memoryCreate2)._writes()._returns(),

		REVERT: new_op(lp_REVERT, 2, 0).with_mem(memoryRevert)._reverts()._returns(),

		INVALID:      new_op(lp_INVALID, 0, 0),
		SELFDESTRUCT: new_op(lp_SELFDESTRUCT, 1, 0)._writes()._halts(),
	}
}
