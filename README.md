# bitvec

`bitvec` is a lightweight Go library providing dynamic bit vectors and rank/select indices. It supports efficient **Set**, **Delete**, **Get**, and **Flip** operations, and allows building **rank/select** indices for compact data structures.

## Features

* Dynamic BitVector with automatic resizing
* O(1) Set/Delete/Get/Flip
* Rank/Select index support for compact data structures
* Lightweight and zero dependencies

## Installation

```bash
go get github.com/yourusername/bitvec
```

## Example

```go
package main

import (
	"fmt"
	"github.com/yourusername/bitvec"
)

func main() {
	// Create a BitVector of size 16
	bv := bitvec.NewVector(16)

	// Set some bits
	bv.Set(1)
	bv.Set(3)
	bv.Set(7)

	fmt.Println("BitVector state:")
	for i := 0; i < 8; i++ {
		fmt.Printf("%d: %v\n", i, bv.Get(i))
	}

	// Build a BitIndex for rank/select
	idx := bitvec.NewIndex(bv)

	fmt.Println("\nRank operations:")
	for i := 0; i < bv.Size(); i++ {
		fmt.Printf("Rank(%d) = %d\n", i, idx.Rank(i))
	}

	fmt.Println("\nSelect operations:")
	for k := uint64(1); k <= 3; k++ {
		pos := idx.Select(k)
		fmt.Printf("Select(%d) = %d\n", k, pos)
	}

	// Flip a bit and rebuild index
	bv.Flip(2)
	idx = bitvec.NewIndex(bv)
	fmt.Println("\nAfter flipping bit 2:")
	fmt.Printf("Rank(3) = %d\n", idx.Rank(3))
	fmt.Printf("Select(2) = %d\n", idx.Select(2))
}
```

## License

MIT
