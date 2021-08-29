package main

import (
	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon/common"
)

type mock_state struct {
	state   map[common.Address]map[common.Hash]uint256.Int
	balance map[common.Address]uint256.Int
}

func new_mock_state() mock_state {
	state := make(map[common.Address]map[common.Hash]uint256.Int)
	balance := make(map[common.Address]uint256.Int)
	m := mock_state{state, balance}
	return m
}

func (m *mock_state) get_state(addr common.Address, key *common.Hash, val *uint256.Int) bool {
	if kvStorage, ok := m.state[addr]; ok {
		if value, ok := kvStorage[*key]; ok {
			*val = value
			return true
		}
		return false
	}
	return false
	// *val := value
}

func (m *mock_state) set_state(addr common.Address, key *common.Hash, val uint256.Int) {
	if kvStorage, ok := m.state[addr]; ok {
		kvStorage[*key] = val
	} else {
		kvStorage = make(map[common.Hash]uint256.Int)
		kvStorage[*key] = val
		m.state[addr] = kvStorage
	}
}

func (m *mock_state) get_balance(addr common.Address) *uint256.Int {
	if balance, ok := m.balance[addr]; ok {
		return &balance
	}
	return nil
}

func (m *mock_state) add_balance(addr common.Address, amount *uint256.Int) {
	if balance, ok := m.balance[addr]; ok {
		new_balance := balance.Add(&balance, amount)
		m.balance[addr] = *new_balance
	} else {
		m.balance[addr] = *amount
	}
}

func (m *mock_state) suicide(addr common.Address) {

}
