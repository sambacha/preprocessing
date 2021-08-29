package main

import (
	"fmt"

	"github.com/ledgerwatch/erigon/common"
)

// set of read/write addresses of every transaction
type rw_set struct {
	read_set  map[common.Address]bool
	write_set map[common.Address]bool
	// contains indexes of transactions
	// that write to the same address as this transaction read
	cross_set []int
}

func new_set_addr() *rw_set {
	read_set := make(map[common.Address]bool)
	write_set := make(map[common.Address]bool)
	cross_set := make([]int, 0)
	return &rw_set{read_set, write_set, cross_set}
}

func (set *rw_set) add_cross(txn_idx int, write_set *map[common.Address]bool) {

	for addr := range *write_set {
		// if read set of this transaction has an address
		// that other transaction writes to it
		if _, ok := set.read_set[addr]; ok {
			set.cross_set = append(set.cross_set, txn_idx)
		}
	}
}

func (set *rw_set) add(addr common.Address, mode int) {
	if mode == READ {
		set.read_set[addr] = true
		return
	}

	if mode == WRITE {
		set.write_set[addr] = true
		return
	}

	panic("Invalid mode. Possible modes are: READ and WRITE\n")
}

func (set *rw_set) has(addr common.Address, mode int) bool {
	if mode == READ {
		if _, ok := set.read_set[addr]; ok {
			return true
		}
		return false
	}

	if mode == WRITE {
		if _, ok := set.write_set[addr]; ok {
			return true
		}
		return false
	}

	panic("Invalid mode. Possible modes are: READ and WRITE\n")
}

func (set *rw_set) print(idx int) {
	fmt.Printf("\n**** transaction: %d ****\n", idx)
	fmt.Println("read set: ")
	if len(set.read_set) > 0 {
		for addr := range set.read_set {
			fmt.Println(addr)
		}
	} else {
		fmt.Println("-- empty --")
	}
	fmt.Println()
	fmt.Println("write set: ")

	if len(set.write_set) > 0 {
		for addr := range set.write_set {
			fmt.Println(addr)
		}
	} else {
		fmt.Println("-- empty --")
	}

}

/* ---------------------------------------------------- */

// container of unique byte slices for each exec frame
// used in REVERT and RETURN as well as call instructions
type byte_set struct {
	store map[int][][]byte
}

func new_byte_set() *byte_set {
	return &byte_set{store: make(map[int][][]byte)}
}

func (set *byte_set) add(level int, data []byte) {
	if container, ok := set.store[level]; !ok {
		container = append(container, data)
		set.store[level] = container
	} else {
		d_size := len(data)
		for _, existing := range container {
			e_size := len(existing)

			greater := e_size > d_size
			less := e_size < d_size

			if greater || less {
				continue
			}

			// equal size case
			exist := true
			for i := 0; i < e_size; i++ {
				if existing[i] != data[i] {
					exist = false
					break
				}
			}
			if exist {
				return
			}

		}
		container = append(container, data)
		set.store[level] = container
	}
}

func (set *byte_set) get(level int) [][]byte {
	return set.store[level]
}

func (set *byte_set) renew(level int) {
	set.store[level] = set.store[level][:0]
}

/* ---------------------------------------------------- */

// container of unique addresses for each exec frame
// used in CREATE and CREATE2
type create_set struct {
	data map[int]common.Address
}

func new_create_set() *create_set {
	return &create_set{data: make(map[int]common.Address)}
}

func (set *create_set) renew(level int) {
	delete(set.data, level)
}

func (set *create_set) set(level int, addr common.Address) {
	if _, ok := set.data[level]; !ok {
		set.data[level] = addr
	} else {
		panic("ADDRESS AT LEVEL EXISTS")
	}
}

func (set *create_set) get(level int) common.Address {
	return set.data[level]
}

/* ---------------------------------------------------- */

type set_u64 struct {
	data []uint64
}

func new_set_u64() *set_u64 {
	return &set_u64{data: make([]uint64, 0)}
}

func (set *set_u64) add(n uint64) {
	for _, d := range set.data {
		if d == n {
			return
		}
	}
	set.data = append(set.data, n)
}
