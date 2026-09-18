// The MIT License (MIT)
//
// Copyright (c) 2015 xtaci
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package heap

import (
	"math/rand"
	"sort"
	"testing"
)

// intHeap is a min-heap of ints implementing Interface[int]. It exercises the
// generic routines exactly the way a real caller (e.g. timedFuncHeap) would,
// and stores ints by value with no any-boxing.
type intHeap []int

func (h intHeap) Len() int           { return len(h) }
func (h intHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h intHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *intHeap) Push(x int)        { *h = append(*h, x) }
func (h *intHeap) Pop() int {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// TestOrdering checks that Pop drains the heap in fully sorted order.
func TestOrdering(t *testing.T) {
	h := &intHeap{}

	const n = 2000
	want := make([]int, 0, n)
	for i := 0; i < n; i++ {
		v := rand.Intn(1 << 20)
		Push[int](h, v)
		want = append(want, v)
	}
	sort.Ints(want)

	if h.Len() != len(want) {
		t.Fatalf("Len = %d, want %d", h.Len(), len(want))
	}
	for i := 0; i < len(want); i++ {
		if got := Pop[int](h); got != want[i] {
			t.Fatalf("Pop %d: got %d, want %d", i, got, want[i])
		}
	}
	if h.Len() != 0 {
		t.Fatalf("Len = %d after draining, want 0", h.Len())
	}
}

// TestInit checks Init heapifies an arbitrary slice so the root is the minimum.
func TestInit(t *testing.T) {
	h := &intHeap{9, 4, 7, 1, 8, 2, 6, 3, 5, 0}
	Init[int](h)

	prev := -1
	for h.Len() > 0 {
		got := Pop[int](h)
		if got < prev {
			t.Fatalf("Pop returned %d after %d: not sorted", got, prev)
		}
		prev = got
	}
}

// TestInterleaved checks the heap invariant holds when Push and Pop are
// interleaved: every Pop must return the current minimum of what's present.
func TestInterleaved(t *testing.T) {
	h := &intHeap{}

	present := make(map[int]int) // reference multiset of what's in the heap
	curMin := func() int {
		m := int(^uint(0) >> 1) // maxInt
		for v, c := range present {
			if c > 0 && v < m {
				m = v
			}
		}
		return m
	}

	for i := 0; i < 5000; i++ {
		if h.Len() == 0 || rand.Intn(2) == 0 {
			v := rand.Intn(1000)
			Push[int](h, v)
			present[v]++
			continue
		}
		want := curMin()
		if got := (*h)[0]; got != want {
			t.Fatalf("iter %d: root = %d, want min %d", i, got, want)
		}
		got := Pop[int](h)
		if got != want {
			t.Fatalf("iter %d: Pop = %d, want %d", i, got, want)
		}
		present[got]--
	}
}

// TestRemoveAndFix checks Remove(i) and Fix(i) keep the heap consistent.
func TestRemoveAndFix(t *testing.T) {
	h := &intHeap{}
	for _, v := range []int{5, 3, 8, 1, 9, 2, 7} {
		Push[int](h, v)
	}

	// remove the current minimum via Remove(0); should equal Pop semantics.
	if got := Remove[int](h, 0); got != 1 {
		t.Fatalf("Remove(0) = %d, want 1", got)
	}

	// mutate an arbitrary element and Fix it, then verify sorted drain.
	(*h)[0] = -100
	Fix[int](h, 0)
	if got := (*h)[0]; got != -100 {
		t.Fatalf("root after Fix = %d, want -100", got)
	}

	prev := int(^uint(0)>>1) * -1 // minInt
	for h.Len() > 0 {
		got := Pop[int](h)
		if got < prev {
			t.Fatalf("Pop returned %d after %d: not sorted", got, prev)
		}
		prev = got
	}
}

// BenchmarkPushPop exercises a full fill/drain cycle. Steady-state allocations
// should be ~0 (only amortized slice growth), demonstrating that no element is
// boxed into an interface{}.
func BenchmarkPushPop(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h := &intHeap{}
		for j := 0; j < 64; j++ {
			Push[int](h, j^0x5a)
		}
		for h.Len() > 0 {
			_ = Pop[int](h)
		}
	}
}
