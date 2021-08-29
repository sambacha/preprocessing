package main

import (
	"math/rand"

	"github.com/holiman/uint256"
	"github.com/ledgerwatch/erigon/common"
)

func random_u256() uint256.Int {
	var n = [4]uint64{
		rand.Uint64(),
		rand.Uint64(),
		rand.Uint64(),
		rand.Uint64(),
	}

	return uint256.Int(n)
}

func random_address() common.Address {
	const size = common.AddressLength
	var result [size]byte

	for i := 0; i < size; i++ {
		result[i] = byte(rand.Intn(256))
	}
	return result
}
