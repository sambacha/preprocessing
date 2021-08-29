package main

const (
	READ        int    = 0x1010
	WRITE       int    = 0x2020
	ROOT_PARENT uint64 = 0xABCDEF01
	NON         byte   = 0x2F
)

var (
	interested_opcodes = interesting_opcodes()
	halt_opcodes       = halting_opcodes()
	rw_modes           = rw_opcodes()
)

/* ----------------- Helper functions ----------------- */

// Returns slice of booleans that marks every instruction
// we are interested in
func interesting_opcodes() (result [256]bool) {
	for i := 0; i < 256; i++ {
		b := byte(i)

		if b == BALANCE || b == EXTCODESIZE || b == EXTCODECOPY ||
			b == EXTCODEHASH ||
			b == CREATE || b == CREATE2 || b == CALL ||
			b == CALLCODE || b == DELEGATECALL || b == STATICCALL ||
			b == JUMP || b == JUMPI {
			result[i] = true
		}
	}
	return
}

func halting_opcodes() (result [256]bool) {
	for i := 0; i < 256; i++ {
		b := byte(i)
		if b == REVERT || b == RETURN ||
			b == STOP || b == SELFDESTRUCT {
			result[i] = true
		}
	}
	return
}

func rw_opcodes() (result [256]int) {
	for i := 0; i < 256; i++ {
		b := byte(i)
		if b == BALANCE || b == EXTCODESIZE || b == EXTCODECOPY || b == EXTCODEHASH || b == SELFBALANCE || b == SLOAD {
			result[i] = READ
		}

		if b == SSTORE {
			result[i] = WRITE
		}
	}
	return
}

// Returns TRUE, if REVERT, RETURN, STOP or SELFDESTRUCT instructions come
// before interested_opcodes or JUMP/JUMPI. Also returns TRUE, if pc reached
// the end of the bytecode without encountering interested_opcodes, JUMP/JUMPI.
// Returns FALSE if interested_opcodes or JUMP/JUMPI come before instructions
// that stop execution
func is_skippable(pc, size uint64, bytecode *[]byte) (byte, bool) {

	for i := pc; i < size; {

		opcode := (*bytecode)[i]

		if opcode >= 0x60 && opcode <= 0x7F { // PUSH instractions
			takes := opcode - 0x5F // how many bytes PUSH takes?
			i += uint64(takes) + 1
			continue
		}

		if interested_opcodes[opcode] {
			return opcode, false
		}

		if halt_opcodes[opcode] {
			return opcode, true
		}

		i += 1 // opcode itself
	}

	return NON, true
}

func check_instructions(pc, size uint64, bytecode *[]byte) (bool, bool) {
	stack_dependant, new_exec_frame := false, false

	for i := pc; i < size; {
		opcode := (*bytecode)[i]

		if opcode >= 0x60 && opcode <= 0x7F { // PUSH instractions
			takes := opcode - 0x5F // how many bytes PUSH takes?
			i += uint64(takes) + 1
			continue
		}

		if opcode == JUMP || opcode == JUMPI {
			return stack_dependant, new_exec_frame
		}

		if opcode == BALANCE || opcode == EXTCODESIZE ||
			opcode == EXTCODECOPY || opcode == EXTCODEHASH {
			stack_dependant = true
		}

		if opcode == CREATE || opcode == CREATE2 || opcode == CALL ||
			opcode == CALLCODE || opcode == DELEGATECALL ||
			opcode == STATICCALL {
			new_exec_frame = true
		}
		i += 1
	}

	return stack_dependant, new_exec_frame
}

// checks if bytecode reads and/or writes to contract address
func reads_writes(bytecode *[]byte, size uint64) (bool, bool) {
	reads, writes := false, false
	for i := uint64(0); i < size; {
		opcode := (*bytecode)[i]

		if reads && writes {
			return true, true
		}

		if opcode >= 0x60 && opcode <= 0x7F { // PUSH instractions
			takes := opcode - 0x5F // how many bytes PUSH takes?
			i += uint64(takes) + 1
			continue
		}

		if opcode == SELFBALANCE || opcode == SLOAD {
			reads = true
		}

		if opcode == SSTORE {
			writes = true
		}

		i += 1
	}

	return reads, writes
}

// Returns slice of booleans that marks every valid jump.
func make_valid_jumpdests(bytecode *[]byte) []bool {
	l := len(*bytecode)
	s := make([]bool, l) // slice of valid jumpdests

	for i := 0; i < l; {
		opcode := (*bytecode)[i]

		if opcode >= 0x60 && opcode <= 0x7F { // PUSH instractions
			takes := opcode - 0x5F // how many bytes PUSH takes?
			i += int(takes)
		} else if opcode == 0x5B { // JUMPDEST
			s[i] = true
		}

		i += 1 // opcode itself
	}

	return s
}

type tree struct {
	root *node
}

type node struct {
	start       uint64 // starting point of the node in bytecode
	stop        uint64 // ending point
	jump_dest   uint64 // 0 means no jumpdest
	parent      uint64 // starting pc of the previous code that led to this
	left_child  *node
	right_child *node
}

type code_state struct {
	end_byte byte
}

func new_tree(evm *evm, contract *Contract, input []byte) {

	evm.return_data.renew(evm.level)

	bytecode := contract.Code
	code_size := uint64(len(bytecode))

	if code_size == 0 {
		return
	}

	valid_jumpdests := make_valid_jumpdests(&bytecode)

	contract.Input = input

	ctx := &callCtx{
		memory:   NewMemory(),
		stack:    NewStack(),
		contract: contract,
	}

	seen := make(map[uint64]bool)

	new_node(evm, ctx, ROOT_PARENT, 0, &valid_jumpdests, &bytecode, &code_size, &seen)

}

func new_node(evm *evm, ctx *callCtx, parent, pc uint64, valid_jumpdests *[]bool, bytecode *[]byte, code_size *uint64, seen *map[uint64]bool) {

	if evm.level > 4 { // 4 recursions, so abort
		evm.abort = true
		evm.result = false
		return
	}

	start := pc
	if _, ok := (*seen)[start]; !ok {
		// we have never executed code starting at 'start' before
		(*seen)[start] = true // now we have seen this

		jump_dest, is_jump := evm.interpreter.run(
			&pc, ctx, bytecode, code_size)
		// fmt.Println(jump_dest, is_jump)
		// next instruction in bytecode.
		// if bytecode[pc] == JUMPI then 'stop' is starting point of the false
		// condition
		stop := pc + 1

		if !is_jump {
			if jump_dest == GAS_CONST_ERR || jump_dest == TOO_LARGE_MEM_ERR ||
				jump_dest == GAS_UNIT_OVERFLOW ||
				jump_dest == STACK_OVERFLOW || jump_dest == STACK_UNDERFLOW {

				evm.abort = true
				evm.result = false
				return
			}
		}

		if is_jump {
			is_valid_jump := jump_dest < *code_size &&
				(*valid_jumpdests)[jump_dest]

			is_loop := false

			if is_valid_jump {
				if jump_dest == start { // loop, jumps back to itself
					is_loop = true

					// we have at least 2 instructions between JUMPDEST and JUMP
					// JUMPDEST
					// COMPERSION OPERATION
					// PUSH2
					// JUMPI
					have_code := pc-4 > jump_dest

					_, new_frame := check_instructions(start, *code_size, bytecode)

					if new_frame { // creates new exec frame in loop
						evm.abort = true
						evm.result = false
						return
					}

					if (*bytecode)[pc] == JUMPI && have_code {
						stop, success := handle_loop(evm, ctx, start, *code_size, bytecode)
						if success {
							ctx_copy := ctx.copy()
							new_node(evm, ctx_copy, start, stop, valid_jumpdests, bytecode, code_size, seen)
						} else { // loop made more then 1000 cycles
							evm.abort = true
							evm.result = false
							return
						}
					}

				}
			}

			if stop < *code_size && (*bytecode)[pc] == JUMPI && !is_loop {
				// JUMPI false condition
				ctx_copy := ctx.copy()

				new_node(evm, ctx_copy, start, stop, valid_jumpdests, bytecode, code_size, seen)
			}

			if is_valid_jump && !is_loop {
				// JUMPI/JUMP true condition or JUMP
				new_node(evm, ctx, start, jump_dest, valid_jumpdests, bytecode, code_size, seen)
			}
		}

	} else {
		// we have executed code starting at 'start' before

		opcode, skip := is_skippable(start, *code_size, bytecode)

		// posible scenarios:
		// 1. at the end code block may halt execution
		// 2. at the end it may jump to the block we have never seen before
		// 3. at the end it may jump to the block that ends up jupming
		// back to this pc:
		//   a -> b -> a
		// 	 a -> b -> c -> a
		// 	 a -> b -> c -> d -> a
		// 	 etc

		if skip { // scenario 1

			if opcode == SELFDESTRUCT { // possible account deletion
				// repost possible suicide
				evm.suicide = true
			}

		} else { // scenarios 2, 3
			// we can't skip, it may jump to the block we have never
			// been before, we need to run interpreter to see that
			_pc := start

			jump_dest, is_jump := evm.interpreter.run(
				&_pc, ctx, bytecode, code_size)

			stop := _pc + 1

			if !is_jump {
				if jump_dest == GAS_CONST_ERR ||
					jump_dest == TOO_LARGE_MEM_ERR ||
					jump_dest == GAS_UNIT_OVERFLOW ||
					jump_dest == STACK_OVERFLOW ||
					jump_dest == STACK_UNDERFLOW {
					evm.abort = true
					evm.result = false
					return
				}
			}

			if is_jump {
				if stop < *code_size && (*bytecode)[_pc] == JUMPI {
					// if code block ends with JUMPI, did we execute
					// it for the false condition before?
					if _, ok := (*seen)[stop]; !ok {
						// we did not
						left_ctx := ctx.copy()
						new_node(evm, left_ctx, start, stop, valid_jumpdests, bytecode, code_size, seen)
					} else {
						// we did it before
						// do we need to execute it again?
					}
				}

				if jump_dest < *code_size && (*valid_jumpdests)[jump_dest] {
					// if we have valid jump, did we execute this before?
					if _, ok := (*seen)[jump_dest]; !ok {
						// we did not
						// fmt.Println("WE DID NOT EXECUTE ----> TRUE CONDITION")
						new_node(evm, ctx, start, jump_dest, valid_jumpdests, bytecode, code_size, seen)
					} else {
						// we did it before
						// do we need to execute it again?
					}
				}
			}

		}
	}
}

// runs loop and returns the pc for false condition of the stack and true.
// if more then 100 loop cycles are performed returns 0 and false
func handle_loop(evm *evm, ctx *callCtx, start, size uint64, bytecode *[]byte) (uint64, bool) {

	var stop uint64
	for cycle := 0; cycle < 1000; cycle++ {

		pc := start
		// execute the code, get the jump destination
		jump_dest := evm.interpreter.lp_run(&pc, ctx, bytecode, &size)
		stop = pc + 1 // staring point of the false condition

		if jump_dest == GAS_CONST_ERR || jump_dest == TOO_LARGE_MEM_ERR ||
			jump_dest == GAS_UNIT_OVERFLOW ||
			jump_dest == STACK_OVERFLOW || jump_dest == STACK_UNDERFLOW {
			return 0, false
		}

		if stop == jump_dest { // loop ended, with condition is 0
			return stop, true
		}
	}

	return 0, false
}
