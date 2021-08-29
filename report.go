package main

import (
	"fmt"

	"github.com/ledgerwatch/erigon/common"
)

type tuple struct {
	depth       int
	instraction string
	initiator   string
	addr        string
}

type report struct {
	blockNumber  uint64
	txnIDX       int
	currentLevel int
	accessPoints []tuple
	depth        int
}

func newTuple(depth int, instraction string, initiator common.Hash, addr common.Hash) tuple {
	return tuple{depth, instraction, initiator.String(), addr.String()}
}

func newReport(bn uint64, idx int) *report {
	var ac []tuple
	r := &report{blockNumber: bn, txnIDX: idx, currentLevel: 0, accessPoints: ac, depth: 0}

	return r
}

func (t tuple) is_equal(rhs tuple) bool {
	return t.depth == rhs.depth && t.instraction == rhs.instraction && t.initiator == rhs.initiator && t.addr == rhs.addr
}

// increment current level by one
func (r *report) inc() {
	r.currentLevel++
	r.depth++
}

// decrement current level by one
func (r *report) dec() {
	r.currentLevel--
}

func (r *report) getLevel() int {
	return r.currentLevel
}

func (r *report) add(t_in tuple) {
	if len(r.accessPoints) > 0 {
		for _, t := range r.accessPoints {
			added := t.is_equal(t_in)
			if added {
				return
			}
		}
		r.accessPoints = append(r.accessPoints, t_in)

	} else {
		r.accessPoints = append(r.accessPoints, t_in)
	}

}

func (t tuple) print() {
	tabs := ""
	for i := 0; i < t.depth; i++ {
		tabs += "\t"
	}
	fmt.Printf("%v----\n", tabs)
	fmt.Printf("%vdepth-%d. instraction: %v\n", tabs, t.depth, t.instraction)
	fmt.Printf("%vinitiator: %s\n", tabs, t.initiator)
	fmt.Printf("%vaccess point: %s\n", tabs, t.addr)
	fmt.Printf("%v----\n", tabs)
	fmt.Println()
}

func (r *report) print_report() {
	flag := false
	if len(r.accessPoints) > 0 {
		flag = true
	}
	if flag {
		fmt.Println("--------------------------------------------")
		fmt.Printf("block number: %d, txn index: %d, recursions: %d\n", r.blockNumber, r.txnIDX, r.depth-1)
		fmt.Println("access points:")
		for _, t := range r.accessPoints {
			t.print()
		}

		// time.Sleep(time.Second * 2)
	}

}
