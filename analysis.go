package main

import (
	"context"
	"fmt"
	"log"
	"time"

	// "time"

	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
	"github.com/ledgerwatch/erigon/core/rawdb"
	"github.com/ledgerwatch/erigon/core/state"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/params"
	"github.com/ledgerwatch/erigon/tests"
	log_ "github.com/ledgerwatch/log/v3"
)

// used for dot files
var (
	BLOCK_NUMBER = -1
	TXN_IDX      = -1
)

func analize(block *types.Block, ibs *state.IntraBlockState, chainCfg *params.ChainConfig) []*evm {

	blockN := block.NumberU64()
	BLOCK_NUMBER = int(blockN)

	var result []*evm

	// fmt.Printf("---------------- BLOCK %d ----------------\n", blockN)

	for i, txn := range block.Transactions() {
		// fmt.Printf("-- TXNID %d --\n", i)
		TXN_IDX = i

		chainCfg.ChainID = txn.GetChainID().ToBig()
		signer := types.MakeSigner(chainCfg, blockN)
		msg, err := txn.AsMessage(*signer, block.BaseFee())

		if err != nil {
			log.Fatal(err)
		}

		contractCreation := msg.To() == nil
		sender := AccountRef(msg.From())
		evm := new_evm(block, ibs, *chainCfg, msg)
		input := txn.GetData() // in case of contract creation it's a code
		value := txn.GetValue()

		if contractCreation {
			// create contract
			evm.create(sender, input, value)
		} else {
			// message call
			evm.call(sender, *msg.To(), input, value)
		}

		TXN_IDX = -1
		result = append(result, evm)
	}
	BLOCK_NUMBER = -1
	return result
}

func handle_results(results []*evm, block_number int) bool {

	all := true // did all analysis finished successfully?
	for _, _evm := range results {

		// if at least one of them failed, there is no way
		// we can confirm that transactions either depend on each other or not
		if !_evm.result {
			all = false
			break
		}
	}

	if all { // if all transactions finished successfully
		// check read/write sets
		all_read := true
		for _, _evm := range results {
			if len(_evm.rw_set.write_set) > 0 {
				all_read = false
				break
			}
		}

		if all_read { // if all they do is read - they are independent
			return true
		}

		length := len(results)

		for i := 0; i < length; i++ {
			for j := 0; j < length; j++ {

				if i == j {
					continue
				}

				results[i].rw_set.add_cross(j, &results[j].rw_set.write_set)

			}
		}

		for _, _evm := range results {
			// there is at least one crossing point
			// if this was executed independently
			// there is a risk of messing the state
			if len(_evm.rw_set.cross_set) > 0 {
				return false
			}
		}

		// no crossing points, so it can be executed independently
		return true
	}
	return false
}

// goes over each block from start untill encounters an error.
// analizes this block
func analize_blocks(start int) {
	_log := log_.New()

	db := mdbx.NewMDBX(_log).Path(*CHAINDATA_PATH).MustOpen()
	defer db.Close()

	tx, err := db.BeginRo(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	chainCfg, _, _ := tests.GetChainConfig("London")

	DOT_FLAG = -1

	for i := start; ; i++ {

		block, err := rawdb.ReadBlockByNumber(tx, uint64(i))
		if err != nil {
			log.Fatalf("Error reading block %d: %s\n", i, err)
		}

		reader := state.NewPlainState(tx, block.NumberU64())
		dbstate := state.New(reader)

		results := analize(block, dbstate, chainCfg)
		result := handle_results(results, i)
		fmt.Printf("\nIndependent execution for block #%d: %t\n", i, result)
		if result && len(results) > 1 {
			fmt.Println("Number of transactions: ", len(results))
			for i, _evm := range results {
				_evm.rw_set.print(i)
			}

			time.Sleep(time.Second * 2)
		}
		fmt.Println()
	}
}

// reads single block at block_number.
// analizes this block
func analize_block(block_number int) bool {
	_log := log_.New()
	db := mdbx.NewMDBX(_log).Path(*CHAINDATA_PATH).MustOpen()

	defer db.Close()
	tx, err := db.BeginRo(context.Background())

	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	chainCfg, _, _ := tests.GetChainConfig("London")

	block, err := rawdb.ReadBlockByNumber(tx, uint64(block_number))
	if err != nil {
		log.Fatalln("Error reading block: ", err)
	}

	reader := state.NewPlainState(tx, block.NumberU64())
	dbstate := state.New(reader)

	results := analize(block, dbstate, chainCfg)
	result := handle_results(results, block_number)
	fmt.Printf("\nIndependent execution for block #%d: %t\n", block_number, result)
	fmt.Println("Number of transactions: ", len(results))
	for i, _evm := range results {
		_evm.rw_set.print(i)
	}
	fmt.Println()
	return result
}
