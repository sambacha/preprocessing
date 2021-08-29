package main

import (
	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/crypto"
)

// ContractRef is a reference to the contract's backing object
type ContractRef interface {
	Address() common.Address
}

type AccountRef common.Address

// Address casts AccountRef to a Address
func (ar AccountRef) Address() common.Address { return (common.Address)(ar) }

type Contract struct {
	CallerAddress common.Address
	caller        ContractRef
	self          ContractRef

	Code     []byte
	CodeHash common.Hash
	CodeAddr *common.Address
	Input    []byte

	value *uint256.Int

	// jumpdests map[common.Hash][]uint64 // Aggregated result of JUMPDEST
	// analysis  []uint64                 // Locally cached result of JUMPDEST analysis
}

// NewContract returns a new contract environment for the execution of EVM.
func new_contract(caller ContractRef, object ContractRef, value *uint256.Int) *Contract {
	c := &Contract{CallerAddress: caller.Address(), caller: caller, self: object}

	c.value = value

	return c
}

// Address returns the contracts address
func (c *Contract) Address() common.Address {
	return c.self.Address()
}

// Caller returns the caller of the contract.
//
// Caller will recursively call caller when the contract is a delegate
// call, including that of caller's caller.
func (c *Contract) Caller() common.Address {
	return c.CallerAddress
}

// SetCallCode sets the code of the contract and address of the backing data
// object
func (c *Contract) set_call_code(addr *common.Address, hash common.Hash, code []byte) {
	c.Code = code
	c.CodeHash = hash
	c.CodeAddr = addr
}

// SetCodeOptionalHash can be used to provide code, but it's optional to provide hash.
// In case hash is not provided, the jumpdest analysis will not be saved to the parent context
func (c *Contract) set_code_hash(addr *common.Address, codeAndHash *codeAndHash) {
	c.Code = codeAndHash.code
	c.CodeHash = codeAndHash.hash
	c.CodeAddr = addr
}

func (c *Contract) is_jumpable(dest *uint256.Int) bool {
	udest, overflow := dest.Uint64WithOverflow()
	// PC cannot go beyond len(code) and certainly can't be bigger than 64bits.
	// Don't bother checking for JUMPDEST in that case.
	if overflow || udest >= uint64(len(c.Code)) {
		return false
	}

	return true
}

type codeAndHash struct {
	code []byte
	hash common.Hash
}

func (c *codeAndHash) Hash() common.Hash {
	if c.hash == (common.Hash{}) {
		c.hash = crypto.Keccak256Hash(c.code)
	}
	return c.hash
}
