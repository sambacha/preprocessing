package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os/exec"
)

var (
	DEFAULT_PATH   = "/home/kairat/diskC/goerli/erigon/chaindata/"
	CHAINDATA_PATH = flag.String("chaindata", DEFAULT_PATH, "path to chaindata database")
	BLOCK_INDEX    = flag.Int("block", -1, "block number to run analisys on")
	GRAPHVIZ       = flag.Bool("graphviz", false, "generate graphviz files?")
	LOOP           = flag.Bool("loop", false, "to loop over all blocks starting from block index")

	TREE  bool = true
	GRAPH bool = false
	// generate dot files, 1 true any other number false
	// works only when parse_block() called
	DOT_FLAG int = 1
)

func generagte_svg() {
	items, _ := ioutil.ReadDir(".")
	for _, item := range items {
		if !item.IsDir() {
			file_name := item.Name()
			length := len(file_name)
			if file_name[length-4:] == ".dot" {
				svg_file := file_name[:length-4] + ".svg"
				cmd := exec.Command("dot", "-Tsvgz", "-o", svg_file, file_name)
				if _, err := cmd.CombinedOutput(); err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}

func main() {

	flag.Parse()

	if *GRAPHVIZ {
		GRAPH = true
	}

	if *BLOCK_INDEX < 0 {
		panic("Block index can not be negative number!")
	}

	if *LOOP {
		GRAPH = false
		analize_blocks(*BLOCK_INDEX)
	} else {
		analize_block(*BLOCK_INDEX)
	}

	if GRAPH {
		generagte_svg()
	}
}

// failed_block := 13528
// f_block := 12842
// large_code := 72003
// infinite_loop := 23371

// comment/uncomment one of these functions
// run one of them at a time
// start_block := 740164
// parse_blocks(238857)

// loop := 13526
// BLOCK_TODO := 748348
// block_index_out_of_range := 13526
// cap_out := 105492
// large_code2 := 156893
// many_many_calls := 156981
// to_check := 156893
// too_many_calls := 629654
// loop_stack_underflow := 238857
// non_loop_stack_underflow := 480879
// loop_false := 326800
// many txns := 2216369
// true return := 2216592
