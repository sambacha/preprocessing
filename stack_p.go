package main

import (
	"fmt"
	"sync"
)

const (
	KNOWN        = 1
	UNKNOWN      = -1
	MAX_SIZE int = 100
)

var p_stackPool = sync.Pool{
	New: func() interface{} {
		return &p_stack{data: make([]int, 0, MAX_SIZE)}
	},
}

// parallel stack - keep tracks of known and unknown values
type p_stack struct {
	data []int
}

func new_p_stack() *p_stack {
	return p_stackPool.Get().(*p_stack)
}

func (st *p_stack) push(n int) {
	// NOTE push limit (1024) is checked in baseCheck
	// if st.Len() > 100 {
	// 	panic("STACK SIZE > 100")
	// }
	st.data = append(st.data, n)
}

func (st *p_stack) PushN() {

}

func (st *p_stack) pop() (result int) {
	size := len(st.data)
	result = st.data[size-1]
	st.data = st.data[:size-1]
	return
}

func (st *p_stack) cap() int {
	return cap(st.data)
}

func (st *p_stack) swap(n int) {
	size := st.size()
	st.data[size-n], st.data[size-1] = st.data[size-1], st.data[size-n]
}

func (st *p_stack) dup(n int) {
	st.push(st.data[st.size()-n])
}

func (st *p_stack) peek() *int {
	return &st.data[st.size()-1]
}

// Back returns the n'th item in stack
func (st *p_stack) Back(n int) *int {
	return &st.data[st.size()-n-1]
}

func (st *p_stack) reset() {
	st.data = st.data[:0]
}

func (st *p_stack) size() int {
	return len(st.data)
}

// Print dumps the content of the stack
func (st *p_stack) Print() {
	fmt.Println("### parallel stack ###")
	if len(st.data) > 0 {
		for i, val := range st.data {
			fmt.Printf("%3d  %v\n", i, val)
		}
	} else {
		fmt.Println("-- empty --")
	}
	fmt.Println("######################")
}

func (st *p_stack) _copy() *p_stack {
	return &p_stack{data: st.data}
}
