package main

import (
	"fmt"

	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon/common"
	"golang.org/x/crypto/sha3"
)

/* -------------- 0s: Stop and Arithmetic Operations -------------- */

func g_STOP(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	return 0
}

func g_ADD(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y := callContext.stack.Pop(), callContext.stack.Peek()
	y.Add(&x, y)
	return 0
}

func g_MUL(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y := callContext.stack.Pop(), callContext.stack.Peek()
	y.Mul(&x, y)
	return 0
}

func g_SUB(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y := callContext.stack.Pop(), callContext.stack.Peek()
	y.Sub(&x, y)
	return 0
}

func g_DIV(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y := callContext.stack.Pop(), callContext.stack.Peek()
	y.Div(&x, y)
	return 0
}

func g_SDIV(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y := callContext.stack.Pop(), callContext.stack.Peek()
	y.SDiv(&x, y)
	return 0
}

func g_MOD(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y := callContext.stack.Pop(), callContext.stack.Peek()
	y.Mod(&x, y)
	return 0
}

func g_SMOD(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y := callContext.stack.Pop(), callContext.stack.Peek()
	y.SMod(&x, y)
	return 0
}

func g_ADDMOD(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y, z := callContext.stack.Pop(), callContext.stack.Pop(), callContext.stack.Peek()
	if z.IsZero() {
		z.Clear()
	} else {
		z.AddMod(&x, &y, z)
	}
	return 0
}

func g_MULMOD(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y, z := callContext.stack.Pop(), callContext.stack.Pop(), callContext.stack.Peek()
	if z.IsZero() {
		z.Clear()
	} else {
		z.MulMod(&x, &y, z)
	}
	return 0
}

func g_EXP(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	base, exponent := callContext.stack.Pop(), callContext.stack.Peek()
	switch {
	case exponent.IsZero():
		// x ^ 0 == 1
		exponent.SetOne()
	case base.IsZero():
		// 0 ^ y, if y != 0, == 0
		exponent.Clear()
	case exponent.LtUint64(2): // exponent == 1
		// x ^ 1 == x
		exponent.Set(&base)
	case base.LtUint64(2): // base == 1
		// 1 ^ y == 1
		exponent.SetOne()
	case base.LtUint64(3): // base == 2
		if exponent.LtUint64(256) {
			n := uint(exponent.Uint64())
			exponent.SetOne()
			exponent.Lsh(exponent, n)
		} else {
			exponent.Clear()
		}
	default:
		exponent.Exp(&base, exponent)
	}
	return 0
}

func g_SIGNEXTEND(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	back, num := callContext.stack.Pop(), callContext.stack.Peek()
	num.ExtendSign(num, &back)
	return 0
}

/* -------------- 10s: Comparison & Bitwise Logic Operations -------------- */

func g_LT(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y := callContext.stack.Pop(), callContext.stack.Peek()
	if x.Lt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return 0
}

func g_GT(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y := callContext.stack.Pop(), callContext.stack.Peek()
	if x.Gt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return 0
}

func g_SLT(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y := callContext.stack.Pop(), callContext.stack.Peek()
	if x.Slt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return 0
}

func g_SGT(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y := callContext.stack.Pop(), callContext.stack.Peek()
	if x.Sgt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return 0
}

func g_EQ(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y := callContext.stack.Pop(), callContext.stack.Peek()
	if x.Eq(y) {
		y.SetOne()
	} else {
		y.Clear()
	}
	return 0
}

func g_ISZERO(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x := callContext.stack.Peek()
	if x.IsZero() {
		x.SetOne()
	} else {
		x.Clear()
	}
	return 0
}

func g_AND(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y := callContext.stack.Pop(), callContext.stack.Peek()
	y.And(&x, y)
	return 0
}

func g_OR(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y := callContext.stack.Pop(), callContext.stack.Peek()
	y.Or(&x, y)
	return 0
}

func g_XOR(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x, y := callContext.stack.Pop(), callContext.stack.Peek()
	y.Xor(&x, y)
	return 0
}

func g_NOT(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x := callContext.stack.Peek()
	x.Not(x)
	return 0
}

func g_BYTE(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	th, val := callContext.stack.Pop(), callContext.stack.Peek()
	val.Byte(&th)
	return 0
}

func g_SHL(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	shift, value := callContext.stack.Pop(), callContext.stack.Peek()
	if shift.LtUint64(256) {
		value.Lsh(value, uint(shift.Uint64()))
	} else {
		value.Clear()
	}
	return 0
}

func g_SHR(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	shift, value := callContext.stack.Pop(), callContext.stack.Peek()
	if shift.LtUint64(256) {
		value.Rsh(value, uint(shift.Uint64()))
	} else {
		value.Clear()
	}
	return 0
}

func g_SAR(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	shift, value := callContext.stack.Pop(), callContext.stack.Peek()
	if shift.GtUint64(255) {
		if value.Sign() >= 0 {
			value.Clear()
		} else {
			// Max negative shift: all bits set
			value.SetAllOne()
		}
	}
	n := uint(shift.Uint64())
	value.SRsh(value, n)
	return 0
}

/* -------------- 20s: SHA3 -------------- */

func g_SHA3(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	offset, size := callContext.stack.Pop(), callContext.stack.Peek()
	data := callContext.memory.GetPtr(offset.Uint64(), size.Uint64())
	// data := callContext.fixedMem.load(offset.Uint64(), size.Uint64())

	if intrprtr.hasher == nil {
		intrprtr.hasher = sha3.NewLegacyKeccak256().(keccakState)
	} else {
		intrprtr.hasher.Reset()
	}
	intrprtr.hasher.Write(data)
	if _, err := intrprtr.hasher.Read(intrprtr.hasherBuf[:]); err != nil {
		panic(err)
	}

	size.SetBytes(intrprtr.hasherBuf[:])
	return 0
}

/* -------------- 30s: Environmental Information -------------- */

func g_ADDRESS(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	callContext.stack.Push(new(uint256.Int).SetBytes(callContext.contract.Address().Bytes()))
	return 0
}

func g_BALANCE(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	slot := callContext.stack.Peek()
	address := common.Address(slot.Bytes20())

	// report access points
	// r := intrprtr.evm.report
	// initr := callContext.contract.Address().Hash()
	// t := newTuple(r.getLevel(), "BALANCE", initr, address.Hash())
	// r.add(t)

	slot.Set(intrprtr.evm.state.GetBalance(address))
	return 0
}

func g_ORIGIN(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	callContext.stack.Push(new(uint256.Int).SetBytes(intrprtr.evm.origin.Bytes()))
	return 0
}

func g_CALLER(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	callContext.stack.Push(new(uint256.Int).SetBytes(callContext.contract.Caller().Bytes()))
	return 0
}

func g_CALLVALUE(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	callContext.stack.Push(callContext.contract.value)
	return 0
}

func g_CALLDATALOAD(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	x := callContext.stack.Peek()
	if offset, overflow := x.Uint64WithOverflow(); !overflow {
		data := getData(callContext.contract.Input, offset, 32)
		x.SetBytes(data)
	} else {
		x.Clear()
	}
	return 0
}

func g_CALLDATASIZE(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	callContext.stack.Push(new(uint256.Int).SetUint64(uint64(len(callContext.contract.Input))))
	return 0
}

func g_CALLDATACOPY(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	var (
		memOffset  = callContext.stack.Pop()
		dataOffset = callContext.stack.Pop()
		length     = callContext.stack.Pop()
	)
	dataOffset64, overflow := dataOffset.Uint64WithOverflow()
	if overflow {
		dataOffset64 = 0xffffffffffffffff
	}
	// These values are checked for overflow during gas cost calculation
	memOffset64 := memOffset.Uint64()
	length64 := length.Uint64()

	callContext.memory.Set(memOffset64, length64, getData(callContext.contract.Input, dataOffset64, length64))
	return 0
}

func g_CODESIZE(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	l := new(uint256.Int)
	l.SetUint64(uint64(len(callContext.contract.Code)))
	callContext.stack.Push(l)
	return 0
}

func g_CODECOPY(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	var (
		memOffset  = callContext.stack.Pop()
		codeOffset = callContext.stack.Pop()
		length     = callContext.stack.Pop()
	)
	uint64CodeOffset, overflow := codeOffset.Uint64WithOverflow()
	if overflow {
		uint64CodeOffset = 0xffffffffffffffff
	}
	codeCopy := getData(callContext.contract.Code, uint64CodeOffset, length.Uint64())
	callContext.memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)
	return 0
}

func g_GASPRICE(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	v, overflow := uint256.FromBig(intrprtr.evm.gasprice)
	if overflow {
		// TODO
	}
	callContext.stack.Push(v)
	return 0
}

func g_EXTCODESIZE(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	slot := callContext.stack.Peek()

	address := common.Address(slot.Bytes20())

	// report access points
	// r := intrprtr.evm.report
	// initr := callContext.contract.Address().Hash()
	// t := newTuple(r.getLevel(), "EXTCODESIZE", initr, address.Hash())
	// r.add(t)

	slot.SetUint64(uint64(intrprtr.evm.state.GetCodeSize(address)))
	return 0
}

func g_EXTCODECOPY(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	var (
		stack      = callContext.stack
		a          = stack.Pop()
		memOffset  = stack.Pop()
		codeOffset = stack.Pop()
		length     = stack.Pop()
	)
	addr := common.Address(a.Bytes20())

	// report access points
	// r := intrprtr.evm.report
	// initr := callContext.contract.Address().Hash()
	// t := newTuple(r.getLevel(), "EXTCODECOPY", initr, addr.Hash())
	// r.add(t)

	len64 := length.Uint64()
	codeCopy := getDataBig(intrprtr.evm.state.GetCode(addr), &codeOffset, len64)
	callContext.memory.Set(memOffset.Uint64(), len64, codeCopy)
	return 0
}

func g_RETURNDATASIZE(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	// TODO
	// callContext.stack.Push(new(uint256.Int).SetUint64(uint64(len(intrprtr.returnData))))
	return 0
}

func g_RETURNDATACOPY(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	// TODO
	// var (
	// 	memOffset  = callContext.stack.Pop()
	// 	dataOffset = callContext.stack.Pop()
	// 	length     = callContext.stack.Pop()
	// )

	// offset64, overflow := dataOffset.Uint64WithOverflow()
	// if overflow {
	// 	// return nil, ErrReturnDataOutOfBounds
	// }
	// // we can reuse dataOffset now (aliasing it for clarity)
	// end := dataOffset
	// _, overflow = end.AddOverflow(&dataOffset, &length)
	// if overflow {
	// 	// return nil, ErrReturnDataOutOfBounds
	// }

	// end64, overflow := end.Uint64WithOverflow()
	// if overflow || uint64(len(intrprtr.returnData)) < end64 {
	// 	return nil, _errReturnDataOutOfBounds
	// }
	// callContext.memory.Set(memOffset.Uint64(), length.Uint64(), intrprtr.returnData[offset64:end64])
	return 0
}

func g_EXTCODEHASH(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	slot := callContext.stack.Peek()
	address := common.Address(slot.Bytes20())

	// report access points
	// r := intrprtr.evm.report
	// initr := callContext.contract.Address().Hash()
	// t := newTuple(r.getLevel(), "EXTCODEHASH", initr, address.Hash())
	// r.add(t)

	if intrprtr.evm.state.Empty(address) {
		slot.Clear()
	} else {
		slot.SetBytes(intrprtr.evm.state.GetCodeHash(address).Bytes())
	}
	return 0
}

/* -------------- 40s: Block Information -------------- */

func g_BLOCKHASH(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	num := callContext.stack.Peek()
	num64, overflow := num.Uint64WithOverflow()
	if overflow {
		num.Clear()
	}
	var upper, lower uint64
	upper = intrprtr.evm.block.NumberU64()
	if upper < 257 {
		lower = 0
	} else {
		lower = upper - 256
	}
	if num64 >= lower && num64 < upper {
		num.SetBytes(intrprtr.evm.block.Hash().Bytes())
	} else {
		num.Clear()
	}
	return 0
}

func g_COINBASE(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	callContext.stack.Push(new(uint256.Int).SetBytes(intrprtr.evm.block.Coinbase().Bytes()))
	return 0
}

func g_TIMESTAMP(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	v := new(uint256.Int).SetUint64(intrprtr.evm.block.Time())
	callContext.stack.Push(v)
	return 0
}

func g_NUMBER(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	v := new(uint256.Int).SetUint64(intrprtr.evm.block.NumberU64())
	callContext.stack.Push(v)
	return 0
}

func g_DIFFICULTY(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	v, overflow := uint256.FromBig(intrprtr.evm.block.Difficulty())
	if overflow {
		// return nil, fmt.Errorf("interpreter.evm.Context.Difficulty higher than 2^256-1")
	}
	callContext.stack.Push(v)
	return 0
}

func g_GASLIMIT(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	gaslimit := intrprtr.evm.block.GasLimit()
	callContext.stack.Push(new(uint256.Int).SetUint64(gaslimit))
	return 0
}

func g_CHAINID(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	fmt.Println("----------- NOT IMPLEMENTED(op_CHAINID) -------------")
	return 0
}

func g_SELFBALANCE(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	fmt.Println("----------- NOT IMPLEMENTED(op_SELFBALANCE) -------------")
	return 0
}

/* ----- 50s: Stack, Memory, Storage and Flow Operations ----- */

func g_POP(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	callContext.stack.Pop()
	return 0
}

func g_MLOAD(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	v := callContext.stack.Peek()
	offset := v.Uint64()
	v.SetBytes(callContext.memory.GetPtr(offset, 32))
	return 0
}

func g_MSTORE(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	mStart, val := callContext.stack.Pop(), callContext.stack.Pop()
	callContext.memory.Set32(mStart.Uint64(), &val)
	return 0
}

func g_MSTORE8(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	off, val := callContext.stack.Pop(), callContext.stack.Pop()
	callContext.memory.store[off.Uint64()] = byte(val.Uint64())
	return 0
}

func g_SLOAD(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	loc := callContext.stack.Peek()
	intrprtr.hasherBuf = loc.Bytes32()
	addr := callContext.contract.Address()

	// report access points
	// r := intrprtr.evm.report
	// initr := callContext.contract.Address().Hash()
	// t := newTuple(r.getLevel(), "SLOAD", initr, addr.Hash())
	// r.add(t)

	ok := intrprtr.evm.mstate.get_state(addr, &intrprtr.hasherBuf, loc)

	if !ok {
		intrprtr.evm.state.GetState(addr, &intrprtr.hasherBuf, loc)
	}
	return 0
}

func g_SSTORE(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	loc := callContext.stack.Pop()
	val := callContext.stack.Pop()
	intrprtr.hasherBuf = loc.Bytes32()
	addr := callContext.contract.Address()

	intrprtr.evm.mstate.set_state(addr, &intrprtr.hasherBuf, val)

	return 0
}

func g_JUMP(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	dest := callContext.stack.Pop()
	if callContext.contract.is_jumpable(&dest) {
		return dest.Uint64()
	}
	return 0
}

func g_JUMPI(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	// condition does not metter in this case
	// all we need is a jump destination
	dest, _ := callContext.stack.Pop(), callContext.stack.Pop()
	if callContext.contract.is_jumpable(&dest) {
		return dest.Uint64()
	}
	return 0
}

func g_PC(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	callContext.stack.Push(new(uint256.Int).SetUint64(*pc))
	return 0
}

func g_MSIZE(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	callContext.stack.Push(new(uint256.Int).SetUint64(uint64(callContext.memory.Len())))
	return 0
}

func g_GAS(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	const _gas = uint64(100_000_000_000)
	callContext.stack.Push(new(uint256.Int).SetUint64(_gas))
	return 0
}

func g_JUMPDEST(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	return 0
}

/* ----- f0s: System operations ----- */

func g_CREATE(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	stack := callContext.stack
	for i := 0; i < 3; i++ {
		stack.Pop()
	}
	address := random_address()
	addr := new(uint256.Int).SetBytes(address.Bytes())
	stack.Push(addr)
	return 0
}

func g_CREATE2(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	stack := callContext.stack
	for i := 0; i < 4; i++ {
		stack.Pop()
	}

	address := random_address()
	addr := new(uint256.Int).SetBytes(address.Bytes())
	stack.Push(addr)
	return 0
}

func g_CALL(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	stack := callContext.stack
	for i := 0; i < 7; i++ {
		stack.Pop()
	}
	stack.Push(uint256.NewInt(1))
	return 0
}

func g_CALLCODE(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	stack := callContext.stack
	for i := 0; i < 7; i++ {
		stack.Pop()
	}
	stack.Push(uint256.NewInt(1))
	return 0
}

func g_DELEGATECALL(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	stack := callContext.stack
	for i := 0; i < 6; i++ {
		stack.Pop()
	}
	stack.Push(uint256.NewInt(1))
	return 0
}

func g_STATICCALL(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	stack := callContext.stack
	for i := 0; i < 6; i++ {
		stack.Pop()
	}
	stack.Push(uint256.NewInt(1))
	return 0
}

func g_RETURN(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	stack := callContext.stack
	for i := 0; i < 2; i++ {
		stack.Pop()
	}
	return 0
}

func g_REVERT(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	stack := callContext.stack
	for i := 0; i < 2; i++ {
		stack.Pop()
	}
	return 0
}

func g_INVALID(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	return 0
}

func g_SELFDESTRUCT(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	callContext.stack.Pop()
	return 0
}

/* -------------- PUSH, DUP, SWAP, LOG -------------- */

// opPush1 is a specialized version of pushN
func g_PUSH1(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
	var (
		codeLen = uint64(len(callContext.contract.Code))
		integer = new(uint256.Int)
	)
	*pc++
	if *pc < codeLen {
		callContext.stack.Push(integer.SetUint64(uint64(callContext.contract.Code[*pc])))
	} else {
		callContext.stack.Push(integer.Clear())
	}
	return 0
}

// make push instruction function
func g_makePush(size uint64, pushByteSize int) exec_func {
	return func(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
		codeLen := len(callContext.contract.Code)

		startMin := int(*pc + 1)
		if startMin >= codeLen {
			startMin = codeLen
		}
		endMin := startMin + pushByteSize
		if startMin+pushByteSize >= codeLen {
			endMin = codeLen
		}

		integer := new(uint256.Int)
		callContext.stack.Push(integer.SetBytes(common.RightPadBytes(
			// So it doesn't matter what we push onto the stack.
			callContext.contract.Code[startMin:endMin], pushByteSize)))

		*pc += size
		return 0
	}

}

// make dup instruction function
func g_makeDup(size int64) exec_func {
	return func(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
		callContext.stack.Dup(int(size))
		return 0
	}
}

// make swap instruction function
func g_makeSwap(size int64) exec_func {
	// switch n + 1 otherwise n would be swapped with n
	size++
	return func(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
		callContext.stack.Swap(int(size))
		return 0
	}
}

// make log instruction function, does not perform any logging
// just pushes off 2 + size items of the stuck
func g_makeLog(size int) exec_func {
	return func(pc *uint64, intrprtr *interpreter, callContext *callCtx) uint64 {
		stack := callContext.stack
		stack.Pop()
		stack.Pop()
		for i := 0; i < size; i++ {
			stack.Pop()
		}
		return 0
	}
}
