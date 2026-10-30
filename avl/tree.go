// Package avl implements a generic, self-balancing AVL tree.
//
// An AVL tree is a binary search tree that keeps the heights of the two
// child subtrees of every node within one of each other. After every
// insert or delete the tree rebalances itself with at most O(log n)
// rotations, so lookup, insert and delete all stay O(log n) even for
// adversarial (e.g. already-sorted) input — the failure mode that turns
// a plain BST into a linked list.
//
// The tree is generic over any ordered key type (cmp.Ordered) and any
// value type. It is not safe for concurrent use; wrap it in a mutex if
// goroutines share one instance.
package avl

import (
	"cmp"
	"iter"
)

// Pair is one key/value entry returned by iteration and range queries.
type Pair[K cmp.Ordered, V any] struct {
	Key   K
	Value V
}

type node[K cmp.Ordered, V any] struct {
	key    K
	value  V
	height int
	left   *node[K, V]
	right  *node[K, V]
}

// Tree is a self-balancing binary search tree.
type Tree[K cmp.Ordered, V any] struct {
	root *node[K, V]
	size int
}

// New returns an empty tree.
func New[K cmp.Ordered, V any]() *Tree[K, V] { return &Tree[K, V]{} }

// Len returns the number of entries in the tree.
func (t *Tree[K, V]) Len() int { return t.size }

// Height returns the height of the tree: -1 for empty, 0 for a single
// node. For a balanced tree of n nodes this is O(log n).
func (t *Tree[K, V]) Height() int { return height(t.root) }

func height[K cmp.Ordered, V any](n *node[K, V]) int {
	if n == nil {
		return -1
	}
	return n.height
}

func updateHeight[K cmp.Ordered, V any](n *node[K, V]) {
	n.height = 1 + max(height(n.left), height(n.right))
}

// balanceFactor is left height minus right height; an AVL tree keeps it
// in {-1, 0, 1} for every node.
func balanceFactor[K cmp.Ordered, V any](n *node[K, V]) int {
	return height(n.left) - height(n.right)
}

func rotateRight[K cmp.Ordered, V any](y *node[K, V]) *node[K, V] {
	x := y.left
	t2 := x.right
	x.right = y
	y.left = t2
	updateHeight(y)
	updateHeight(x)
	return x
}

func rotateLeft[K cmp.Ordered, V any](x *node[K, V]) *node[K, V] {
	y := x.right
	t2 := y.left
	y.left = x
	x.right = t2
	updateHeight(x)
	updateHeight(y)
	return y
}

// rebalance restores the AVL invariant at n, assuming its children are
// already balanced. It handles all four cases: LL, RR, LR and RL.
func rebalance[K cmp.Ordered, V any](n *node[K, V]) *node[K, V] {
	updateHeight(n)
	switch bf := balanceFactor(n); {
	case bf > 1: // left-heavy
		if balanceFactor(n.left) < 0 {
			n.left = rotateLeft(n.left) // LR case
		}
		return rotateRight(n)
	case bf < -1: // right-heavy
		if balanceFactor(n.right) > 0 {
			n.right = rotateRight(n.right) // RL case
		}
		return rotateLeft(n)
	default:
		return n
	}
}

// Insert adds or replaces the value for key. Inserting an existing key
// overwrites its value and does not change Len.
func (t *Tree[K, V]) Insert(key K, value V) {
	var inserted bool
	t.root = insert(t.root, key, value, &inserted)
	if inserted {
		t.size++
	}
}

func insert[K cmp.Ordered, V any](n *node[K, V], key K, value V, inserted *bool) *node[K, V] {
	if n == nil {
		*inserted = true
		return &node[K, V]{key: key, value: value}
	}
	switch cmp.Compare(key, n.key) {
	case -1:
		n.left = insert(n.left, key, value, inserted)
	case 1:
		n.right = insert(n.right, key, value, inserted)
	default:
		n.value = value // overwrite, not a new entry
	}
	return rebalance(n)
}

// Get returns the value for key and whether it was found.
func (t *Tree[K, V]) Get(key K) (V, bool) {
	n := t.root
	for n != nil {
		switch cmp.Compare(key, n.key) {
		case -1:
			n = n.left
		case 1:
			n = n.right
		default:
			return n.value, true
		}
	}
	var zero V
	return zero, false
}

// Contains reports whether key is present.
func (t *Tree[K, V]) Contains(key K) bool {
	_, ok := t.Get(key)
	return ok
}

// Delete removes key from the tree. It reports whether anything was removed.
func (t *Tree[K, V]) Delete(key K) bool {
	var deleted bool
	t.root = deleteNode(t.root, key, &deleted)
	if deleted {
		t.size--
	}
	return deleted
}

func deleteNode[K cmp.Ordered, V any](n *node[K, V], key K, deleted *bool) *node[K, V] {
	if n == nil {
		return nil
	}
	switch cmp.Compare(key, n.key) {
	case -1:
		n.left = deleteNode(n.left, key, deleted)
	case 1:
		n.right = deleteNode(n.right, key, deleted)
	default:
		*deleted = true
		switch {
		case n.left == nil:
			return n.right
		case n.right == nil:
			return n.left
		default:
			// Two children: replace with the in-order successor
			// (smallest key in the right subtree), then delete it.
			succ := n.right
			for succ.left != nil {
				succ = succ.left
			}
			n.key, n.value = succ.key, succ.value
			n.right = deleteNode(n.right, succ.key, new(bool))
		}
	}
	return rebalance(n)
}

// Min returns the smallest key/value pair, or false if the tree is empty.
func (t *Tree[K, V]) Min() (Pair[K, V], bool) {
	if t.root == nil {
		return Pair[K, V]{}, false
	}
	n := t.root
	for n.left != nil {
		n = n.left
	}
	return Pair[K, V]{n.key, n.value}, true
}

// Max returns the largest key/value pair, or false if the tree is empty.
func (t *Tree[K, V]) Max() (Pair[K, V], bool) {
	if t.root == nil {
		return Pair[K, V]{}, false
	}
	n := t.root
	for n.right != nil {
		n = n.right
	}
	return Pair[K, V]{n.key, n.value}, true
}

// All yields every entry in ascending key order. Iteration stops early
// if the caller breaks out of the range loop.
func (t *Tree[K, V]) All() iter.Seq[Pair[K, V]] {
	return func(yield func(Pair[K, V]) bool) {
		var walk func(n *node[K, V]) bool
		walk = func(n *node[K, V]) bool {
			if n == nil {
				return true
			}
			return walk(n.left) &&
				yield(Pair[K, V]{n.key, n.value}) &&
				walk(n.right)
		}
		walk(t.root)
	}
}

// Keys returns all keys in ascending order.
func (t *Tree[K, V]) Keys() []K {
	keys := make([]K, 0, t.size)
	for p := range t.All() {
		keys = append(keys, p.Key)
	}
	return keys
}

// Range returns all entries with lo <= key <= hi, in ascending order.
// This is the query a plain hash map cannot answer efficiently, and the
// reason ordered trees still earn their place.
func (t *Tree[K, V]) Range(lo, hi K) []Pair[K, V] {
	var out []Pair[K, V]
	var walk func(n *node[K, V])
	walk = func(n *node[K, V]) {
		if n == nil {
			return
		}
		if cmp.Compare(n.key, lo) > 0 {
			walk(n.left)
		}
		if cmp.Compare(n.key, lo) >= 0 && cmp.Compare(n.key, hi) <= 0 {
			out = append(out, Pair[K, V]{n.key, n.value})
		}
		if cmp.Compare(n.key, hi) < 0 {
			walk(n.right)
		}
	}
	walk(t.root)
	return out
}

// Validate checks the two invariants that define a correct AVL tree:
// keys are in BST order, and every node's balance factor is in {-1, 0, 1}
// with stored heights consistent with its children. It is O(n) and meant
// for tests and debugging, not the hot path.
func (t *Tree[K, V]) Validate() bool {
	_, _, ok := validate(t.root, nil, nil)
	return ok
}

func validate[K cmp.Ordered, V any](n *node[K, V], lo, hi *K) (int, bool, bool) {
	if n == nil {
		return -1, true, true
	}
	if lo != nil && cmp.Compare(n.key, *lo) <= 0 {
		return 0, false, false
	}
	if hi != nil && cmp.Compare(n.key, *hi) >= 0 {
		return 0, false, false
	}
	lh, lok, _ := validate(n.left, lo, &n.key)
	rh, rok, _ := validate(n.right, &n.key, hi)
	if !lok || !rok {
		return 0, false, false
	}
	h := 1 + max(lh, rh)
	if n.height != h || lh-rh > 1 || rh-lh > 1 {
		return 0, false, false
	}
	return h, true, true
}
