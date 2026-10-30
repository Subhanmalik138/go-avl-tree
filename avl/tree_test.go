package avl

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"
)

func TestEmptyTree(t *testing.T) {
	tr := New[int, string]()
	if tr.Len() != 0 || tr.Height() != -1 {
		t.Fatalf("empty tree: len=%d height=%d", tr.Len(), tr.Height())
	}
	if _, ok := tr.Get(1); ok {
		t.Fatal("Get on empty tree should miss")
	}
	if _, ok := tr.Min(); ok {
		t.Fatal("Min on empty tree should miss")
	}
	if !tr.Validate() {
		t.Fatal("empty tree should validate")
	}
}

func TestInsertGetOverwrite(t *testing.T) {
	tr := New[string, int]()
	tr.Insert("b", 2)
	tr.Insert("a", 1)
	tr.Insert("c", 3)
	if tr.Len() != 3 {
		t.Fatalf("len = %d, want 3", tr.Len())
	}
	if v, ok := tr.Get("b"); !ok || v != 2 {
		t.Fatalf("Get(b) = %v, %v", v, ok)
	}
	tr.Insert("b", 20) // overwrite, not a new entry
	if tr.Len() != 3 {
		t.Fatalf("len after overwrite = %d, want 3", tr.Len())
	}
	if v, _ := tr.Get("b"); v != 20 {
		t.Fatalf("Get(b) after overwrite = %d, want 20", v)
	}
	if !tr.Contains("a") || tr.Contains("z") {
		t.Fatal("Contains mismatch")
	}
}

func TestSortedInputStaysBalanced(t *testing.T) {
	// Sorted input is the worst case for a plain BST: it must NOT
	// degenerate here. 1000 sequential inserts must stay ~log2 deep.
	tr := New[int, int]()
	for i := range 1000 {
		tr.Insert(i, i*i)
	}
	if h := tr.Height(); h > 20 {
		t.Fatalf("height = %d after sorted inserts, want <= 20", h)
	}
	if !tr.Validate() {
		t.Fatal("tree invalid after sorted inserts")
	}
	for i := range 1000 {
		if v, ok := tr.Get(i); !ok || v != i*i {
			t.Fatalf("Get(%d) = %v, %v", i, v, ok)
		}
	}
}

func TestAllFourRotationCases(t *testing.T) {
	// Each case is the textbook minimal trigger for one rotation shape.
	cases := map[string][]int{
		"LL": {30, 20, 10},
		"RR": {10, 20, 30},
		"LR": {30, 10, 20},
		"RL": {10, 30, 20},
	}
	for name, keys := range cases {
		tr := New[int, int]()
		for _, k := range keys {
			tr.Insert(k, k)
		}
		if !tr.Validate() {
			t.Fatalf("%s case: tree invalid", name)
		}
		if got := tr.Keys(); !slices.Equal(got, []int{10, 20, 30}) {
			t.Fatalf("%s case: keys = %v", name, got)
		}
	}
}

func TestDeleteLeafOneChildTwoChildren(t *testing.T) {
	tr := New[int, string]()
	for _, k := range []int{50, 30, 70, 20, 40, 60, 80} {
		tr.Insert(k, fmt.Sprint(k))
	}
	if !tr.Delete(20) || tr.Contains(20) {
		t.Fatal("leaf delete failed")
	}
	if !tr.Delete(30) || tr.Contains(30) { // one child left (40)
		t.Fatal("one-child delete failed")
	}
	if !tr.Delete(50) || tr.Contains(50) { // root, two children
		t.Fatal("two-child delete failed")
	}
	if tr.Delete(999) {
		t.Fatal("deleting a missing key should report false")
	}
	if tr.Len() != 4 {
		t.Fatalf("len = %d, want 4", tr.Len())
	}
	if !tr.Validate() {
		t.Fatal("tree invalid after deletes")
	}
	if got := tr.Keys(); !slices.Equal(got, []int{40, 60, 70, 80}) {
		t.Fatalf("keys = %v", got)
	}
}

func TestMinMaxRange(t *testing.T) {
	tr := New[int, string]()
	for _, k := range []int{5, 1, 9, 3, 7} {
		tr.Insert(k, fmt.Sprint(k))
	}
	if mn, _ := tr.Min(); mn.Key != 1 {
		t.Fatalf("Min = %v, want 1", mn)
	}
	if mx, _ := tr.Max(); mx.Key != 9 {
		t.Fatalf("Max = %v, want 9", mx)
	}
	got := tr.Range(3, 7)
	want := []Pair[int, string]{{3, "3"}, {5, "5"}, {7, "7"}}
	if len(got) != len(want) {
		t.Fatalf("Range(3,7) = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Range(3,7)[%d] = %v, want %v", i, got[i], want[i])
		}
	}
	if n := len(tr.Range(8, 2)); n != 0 {
		t.Fatalf("Range with lo > hi returned %d entries", n)
	}
}

func TestInOrderIteration(t *testing.T) {
	tr := New[int, int]()
	for _, k := range []int{4, 2, 6, 1, 3, 5, 7} {
		tr.Insert(k, k*10)
	}
	var keys []int
	for p := range tr.All() {
		keys = append(keys, p.Key)
	}
	if !slices.IsSorted(keys) || len(keys) != 7 {
		t.Fatalf("iteration order = %v", keys)
	}
	// Early break must stop iteration without hanging.
	count := 0
	for range tr.All() {
		count++
		break
	}
	if count != 1 {
		t.Fatalf("early break iterated %d times", count)
	}
}

func TestRandomInsertDeleteStaysValid(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	tr := New[int, int]()
	present := map[int]bool{}
	for step := range 5000 {
		k := rng.Intn(500)
		if step%3 == 0 && present[k] {
			tr.Delete(k)
			delete(present, k)
		} else {
			tr.Insert(k, k)
			present[k] = true
		}
		if !tr.Validate() {
			t.Fatalf("tree invalid at step %d", step)
		}
		if tr.Len() != len(present) {
			t.Fatalf("step %d: len = %d, want %d", step, tr.Len(), len(present))
		}
	}
	if got := tr.Keys(); !slices.IsSorted(got) {
		t.Fatal("final keys not sorted")
	}
}
