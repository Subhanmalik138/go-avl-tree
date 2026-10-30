// Command demo inserts a handful of keys into an AVL tree and shows
// in-order traversal, a range query, and a deletion — the three things
// a sorted map is for.
package main

import (
	"fmt"

	"github.com/Subhanmalik138/go-avl-tree/avl"
)

func main() {
	tr := avl.New[int, string]()
	for _, k := range []int{50, 30, 70, 20, 40, 60, 80, 10} {
		tr.Insert(k, fmt.Sprintf("value-%d", k))
	}

	fmt.Printf("entries: %d, height: %d\n", tr.Len(), tr.Height())
	fmt.Printf("in-order keys: %v\n", tr.Keys())

	fmt.Print("range [25, 65]: ")
	for _, p := range tr.Range(25, 65) {
		fmt.Printf("%d ", p.Key)
	}
	fmt.Println()

	tr.Delete(50)
	fmt.Printf("after deleting 50: %v (valid: %v)\n", tr.Keys(), tr.Validate())
}
