package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

var __OPCODES = map[byte]string{
	0x00: "STOP", 0x01: "ADD", 0x02: "MULL", 0x03: "SUB", 0x04: "DIV",
	0x05: "SDIV", 0x06: "MOD", 0x07: "SMOD", 0x08: "ADDMOD",
	0x09: "MULLMOD", 0x0A: "EXP", 0x0B: "SIGNEXTEND",

	0x10: "LT", 0x11: "GT", 0x12: "SLT", 0x13: "SGT", 0x14: "EQ",
	0x15: "ISZERO", 0x16: "AND", 0x17: "OR", 0x18: "XOR", 0x19: "NOT",
	0x1A: "BYTE", 0x1B: "SHL", 0x1C: "SHR", 0x1D: "SAR",

	0x20: "SHA3",

	0x30: "ADDRESS", 0x31: "BALANCE", 0x32: "ORIGIN", 0x33: "CALLER",
	0x34: "CALLVALUE", 0x35: "CALLDATALOAD", 0x36: "CALLDATASIZE",
	0x37: "CALLDATACOPY", 0x38: "CODESIZE", 0x39: "CODECOPY", 0x3A: "GASPRICE",
	0x3B: "EXTCODESIZE", 0x3C: "EXTCODECOPY", 0x3D: "RETURNDATASIZE",
	0x3E: "RETURNDATACOPY", 0x3F: "EXTCODEHASH",

	0x40: "BLOCKHASH", 0x41: "COINBASE", 0x42: "TIMESTAMP", 0x43: "NUMBER",
	0x44: "DIFFICULTY", 0x45: "GASLIMIT", 0x46: "CHAINID",
	0x47: "SELFBALANCE",

	0x50: "POP", 0x51: "MLOAD", 0x52: "MSTORE", 0x53: "MSTORE8",
	0x54: "SLOAD", 0x55: "SSTORE", 0x56: "JUMP",
	0x57: "JUMPI", 0x58: "PC", 0x59: "MSIZE", 0x5A: "GAS",
	0x5B: "JUMPDEST",

	0x80: "DUP1", 0x81: "DUP2", 0x82: "DUP3", 0x83: "DUP4", 0x84: "DUP5",
	0x85: "DUP6", 0x86: "DUP7", 0x87: "DUP8", 0x88: "DUP9", 0x89: "DUP10",
	0x8A: "DUP11", 0x8B: "DUP12", 0x8C: "DUP13", 0x8D: "DUP14", 0x8E: "DUP15",
	0x8F: "DUP16",

	0x90: "SWAP1", 0x91: "SWAP2", 0x92: "SWAP3", 0x93: "SWAP4",
	0x94: "SWAP5", 0x95: "SWAP6", 0x96: "SWAP7", 0x97: "SWAP8", 0x98: "SWAP9",
	0x99: "SWAP10", 0x9A: "SWAP11", 0x9B: "SWAP12", 0x9C: "SWAP13",
	0x9D: "SWAP14", 0x9E: "SWAP15", 0x9F: "SWAP16",

	0xA0: "LOG0", 0xA1: "LOG1", 0xA2: "LOG2", 0xA3: "LOG3",
	0xA4: "LOG4",

	0xF0: "CREATE", 0xF1: "CALL", 0xF2: "CALLCODE", 0xF3: "RETURN",
	0xF4: "DELEGATECALL", 0xF5: "CREATE2", 0xFA: "STATICCALL", 0xFD: "REVERT",
	0xFE: "INVALID", 0xFF: "SELFDESTRUCT",
}

func print_bytecode(bytecode *[]byte, pc uint64, full bool) {

	out := ""
	for i := pc; i < uint64(len(*bytecode)); {
		opcode := (*bytecode)[i]
		if opcode >= 0x60 && opcode <= 0x7f { // PUSH instructions
			start := i
			takes := opcode - 0x5F // how many bytes PUSH takes?
			out += fmt.Sprintf("PUSH%d ", takes)
			i += 1             // opcode it self
			i += uint64(takes) // taken bytes
			out += hex.EncodeToString((*bytecode)[start+1 : i])
			out += fmt.Sprintf("\t%d...%d\n", start, i-1)
		} else {
			if v, ok := __OPCODES[opcode]; ok {
				out += fmt.Sprintf("%s", v)
				out += fmt.Sprintf("\t%d\n", i)
			}

			if !full {
				if opcode == JUMP || opcode == JUMPI || opcode == STOP || opcode == REVERT || opcode == SELFDESTRUCT || opcode == RETURN {
					fmt.Println((*bytecode)[:i+1])
					break
				}
			}

			i += 1
		}
	}

	fmt.Println(out)
}

func vtx_label(bytecode *[]byte, vtx *vertex) string {

	out := "{"

	for i := vtx.start; i < vtx.stop; {

		opcode := (*bytecode)[i]
		flag := i+1 < vtx.stop

		if opcode >= 0x60 && opcode <= 0x7f { // PUSH instructions
			takes := opcode - 0x5F // how many bytes PUSH takes?
			out += fmt.Sprintf("%X", opcode)
			pc_i := i // pc starting point
			i += 1    // opcode it self
			out += fmt.Sprintf(" - PUSH%d\\l", takes)
			i += uint64(takes) // taken bytes
			out += fmt.Sprintf("%d...%d\\r", pc_i, i-1)
		} else {
			str_opcode := ""
			write_jumpdest := opcode == JUMP || opcode == JUMPI
			if val, ok := __OPCODES[opcode]; ok {
				str_opcode = val
			}
			i += 1 // opcode it self
			if str_opcode != "" {
				if write_jumpdest { // if opcode is JUMP or JUMPI
					out += fmt.Sprintf("%02x - %s (dest: %d)\\l%d\\r", opcode, str_opcode, vtx.jump_dest, i-1)
				} else {
					if opcode == JUMPDEST {
						out += fmt.Sprintf("%02X - %s\\l%d\\r", opcode, str_opcode, i-1)
					} else {
						out += fmt.Sprintf("%02X - %s\\l%d\\r", opcode, str_opcode, i-1)
					}

				}
			}
		}

		if flag {
			out += "|"
		}
	}

	out += "}"
	return out
}

func write_vtx(f io.Writer, bytecode *[]byte, vtx *vertex) {
	if DOT_FLAG == 1 {
		label := vtx_label(bytecode, vtx)
		out := fmt.Sprintf("block_%d [shape=\"record\" label=\"%s\"]", vtx.start, label)
		fmt.Fprintln(f, out)

		if vtx.direction > -1 {
			// direction of a node
			// if 0 (FALSE) means we are creating left node
			var taillabel string

			if vtx.direction == 0 {
				taillabel = "[taillabel=\"FALSE\"]"
			}

			if vtx.direction == 1 {
				if vtx.parentID == vtx.start {
					taillabel = "[label=\"TRUE\" dir=back]"
				} else {
					taillabel = "[taillabel=\"TRUE\"]"
				}

			}

			fmt.Fprintf(f, "block_%d -> block_%d %s\n", vtx.parentID, vtx.start, taillabel)
		}
	}
}

func make_dot_file() *os.File {
	if DOT_FLAG == 1 {
		file_name := ""
		if BLOCK_NUMBER > -1 && TXN_IDX > -1 {
			file_name += strconv.Itoa(BLOCK_NUMBER)
			file_name += "_" + strconv.Itoa(TXN_IDX)

			f, err := os.Create(file_name + "_jumps.dot")
			if err != nil {
				log.Fatal(err)
			}

			fmt.Fprint(f, "digraph bytecode_graph {\n")

			return f
		}

		return nil
	}
	return nil
}
