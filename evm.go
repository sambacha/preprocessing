package main

import (
	"math/big"

	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/core/state"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/crypto"
	"github.com/ledgerwatch/erigon/params"
)

const (
	CREATE_ = 1 + iota
	CREATE2_

	// since gas is unlimited some contracts can call to itself infinite number
	// of times this number limits recursive calls
	MAX_RECURSIONS = 10
)

type access struct {
	mode    int // -1 reads, 1 writes, 0 unknown
	address common.Address
}

type evm struct {
	block       *types.Block
	state       *state.IntraBlockState
	mstate      *mock_state
	chainCfg    params.ChainConfig
	interpreter *interpreter
	origin      common.Address
	gasprice    *big.Int

	abort  bool
	result bool // general analysis result

	rw_set      *rw_set     // set of read/write
	return_data *byte_set   // return data of every exec frame (all calls)
	create_addr *create_set // return addresses for CREATE and CREATE2
	level       int         // currently executing frame level (recursion depth)
	suicide     bool        // if there is possible selfdestruct

	frame_errs map[int][]uint64
}

func new_evm(block *types.Block, state *state.IntraBlockState, chainCfg params.ChainConfig, msg types.Message) *evm {
	origin := msg.From()
	gasprice := msg.GasPrice().ToBig()
	msg.Gas()
	mstate := new_mock_state()
	frame_errs := make(map[int][]uint64)

	// create_addr := make(map[int]common.Address)
	_evm := evm{
		block: block, state: state, mstate: &mstate,
		chainCfg: chainCfg, origin: origin,
		gasprice: gasprice, level: -1,
		frame_errs:  frame_errs,
		return_data: new_byte_set(),
		create_addr: new_create_set(),
		rw_set:      new_set_addr(),
		result:      true, // true by default
	}

	interpreter := new_interpreter(&_evm)
	_evm.interpreter = interpreter
	return &_evm
}

func (evm *evm) call(caller ContractRef, addr common.Address, input []byte, value *uint256.Int) {
	code := evm.state.GetCode(addr)
	addrCopy := addr
	codehash := evm.state.GetCodeHash(addrCopy)
	contract := new_contract(caller, AccountRef(addrCopy), value)
	contract.set_call_code(&addrCopy, codehash, code)

	evm.level += 1
	if GRAPH {
		new_graph(evm, contract, input)
	}

	if TREE {
		new_tree(evm, contract, input)
	}
	evm.level -= 1
}

func (evm *evm) call_code(caller ContractRef, addr common.Address, input []byte, value *uint256.Int) {
	code := evm.state.GetCode(addr)
	addrCopy := addr
	codehash := evm.state.GetCodeHash(addrCopy)
	contract := new_contract(caller, AccountRef(addrCopy), value)
	contract.set_call_code(&addrCopy, codehash, code)

	evm.level += 1
	if GRAPH {
		new_graph(evm, contract, input)
	}

	if TREE {
		new_tree(evm, contract, input)
	}
	evm.level -= 1
}

func (evm *evm) delegate_call(caller ContractRef, addr common.Address, input []byte) {
	code := evm.state.GetCode(addr)
	addrCopy := addr
	codehash := evm.state.GetCodeHash(addrCopy)
	contract := new_contract(caller, AccountRef(addrCopy), new(uint256.Int))
	contract.set_call_code(&addrCopy, codehash, code)

	evm.level += 1
	if GRAPH {
		new_graph(evm, contract, input)
	}

	if TREE {
		new_tree(evm, contract, input)
	}
	evm.level -= 1
}

func (evm *evm) static_call(caller ContractRef, addr common.Address, input []byte) {
	code := evm.state.GetCode(addr)
	addrCopy := addr
	codehash := evm.state.GetCodeHash(addrCopy)
	contract := new_contract(caller, AccountRef(addrCopy), new(uint256.Int))
	contract.set_call_code(&addrCopy, codehash, code)

	evm.level += 1
	if GRAPH {
		new_graph(evm, contract, input)
	}

	if TREE {
		new_tree(evm, contract, input)
	}
	evm.level -= 1
}

func (evm *evm) _create(caller ContractRef, codeAndHash *codeAndHash, value *uint256.Int, address common.Address, calltype int) {
	contract := new_contract(caller, AccountRef(address), value)
	contract.set_code_hash(&address, codeAndHash)

	evm.level += 1
	evm.create_addr.renew(evm.level)
	evm.create_addr.set(evm.level, address)
	if GRAPH {
		new_graph(evm, contract, nil)
	}

	if TREE {
		new_tree(evm, contract, nil)
	}
	evm.level -= 1
}

func (evm *evm) create(caller ContractRef, code []byte, value *uint256.Int) {
	contractAddr := crypto.CreateAddress(caller.Address(), evm.state.GetNonce(caller.Address()))
	evm._create(caller, &codeAndHash{code: code}, value, contractAddr, CREATE_)
}

func (evm *evm) create2(caller ContractRef, code []byte, endowment *uint256.Int, salt *uint256.Int) {
	codeAndHash := &codeAndHash{code: code}
	contractAddr := crypto.CreateAddress2(caller.Address(), common.Hash(salt.Bytes32()), codeAndHash.Hash().Bytes())
	evm._create(caller, codeAndHash, endowment, contractAddr, CREATE2_)
}
