# go-avl-tree

A generic, self-balancing AVL tree in Go — an ordered map that stays
fast no matter what order keys arrive in.

## The problem

A plain binary search tree degrades to a linked list on sorted input:
insert 1, 2, 3, …, 1000 and lookups become O(n). Hash maps fix the
average case but cannot answer ordered queries ("all keys between 25
and 65") or iterate in sorted order. An AVL tree solves both: it
rebalances itself after every insert and delete, so all operations stay
O(log n), and keys are always available in sorted order.

## Design decisions

1. **Generics over `cmp.Ordered`, zero dependencies.** Any ordered key
   type (`int`, `string`, …) works with any value type, using only the
   standard library — no third-party constraint packages.
2. **Recursive insert/delete with a single `rebalance` step.** All four
   rotation cases (LL, RR, LR, RL) funnel through one function, so the
   balancing logic lives in exactly one place and is easy to audit.
3. **Deletion via the in-order successor.** For a two-child node, the
   smallest key in the right subtree takes its place, preserving BST
   order without rebuilding the subtree.
4. **Pruned range queries.** `Range(lo, hi)` skips whole subtrees
   outside the interval instead of scanning everything — the query a
   hash map cannot answer efficiently.
5. **A `Validate()` invariant checker.** It verifies BST ordering,
   balance factors, and stored heights in O(n). It exists for tests and
   debugging, deliberately kept off the hot path.

## Usage

```go
import "github.com/Subhanmalik138/go-avl-tree/avl"

tr := avl.New[int, string]()
tr.Insert(50, "fifty")
tr.Insert(30, "thirty")

v, ok := tr.Get(30)        // "thirty", true
fmt.Println(tr.Keys())     // [30 50]
fmt.Println(tr.Range(25, 65)) // [{50 fifty}]

for p := range tr.All() {  // ascending key order
    fmt.Println(p.Key, p.Value)
}
tr.Delete(50)
```

## Run

```bash
go run ./cmd/demo
```

## Tests

```bash
go test ./...
go vet ./...
```

The suite covers all four rotation cases, sorted-input balance
(1000 sequential inserts stay ≤ 20 high), leaf/one-child/two-child
deletion, range queries, and a 5000-step randomized insert/delete run
that re-validates the tree after every step.

## Tech stack

Go 1.23, standard library only.
