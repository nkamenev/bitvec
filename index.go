package bitvec

import (
	"math/bits"
	"sort"
)

const (
	superBlock    = 512
	superBlockLog = 9
)

// BitIndex is a two-level succinct index over a BitVector that
// supports O(1) Rank and O(log n) Select operations.
//
// The index uses a superblock/block decomposition:
//
//   - superRank stores the number of set bits (1s) at the beginning
//     of each 512-bit superblock.
//   - blockRank stores the number of set bits from the beginning
//     of the superblock to the beginning of each 64-bit word.
//
// This layout allows fast Rank queries and logarithmic Select
// without scanning the entire bit vector.
//
// The index does not copy the underlying BitVector.
// The BitVector must not be mutated after the index is built.
type BitIndex struct {
	vec       *BitVector
	superRank []uint64
	blockRank []uint16
}

// NewIndex builds a rank/select index over the given BitVector.
//
// The construction runs in O(n) time where n is the number of
// 64-bit words in the vector.
//
// It precomputes:
//   - superRank: cumulative ranks at each 512-bit boundary
//   - blockRank: ranks relative to the start of each superblock
//
// Returns nil if the BitVector is empty.
func NewIndex(vec *BitVector) *BitIndex {
	nWords := len(vec.words)
	if nWords == 0 {
		return nil
	}

	nSuper := (vec.size + superBlock - 1) >> superBlockLog

	superRank := make([]uint64, nSuper)
	blockRank := make([]uint16, nWords)

	var (
		rank       uint64
		superIdx   int
		wordsPerSB = superBlock >> wordSizeLog
	)

	for w := range nWords {
		if w%wordsPerSB == 0 {
			superRank[superIdx] = rank
			superIdx++
		}

		blockRank[w] = uint16(rank - superRank[superIdx-1])
		rank += uint64(bits.OnesCount64(vec.words[w]))
	}

	return &BitIndex{
		vec:       vec,
		superRank: superRank,
		blockRank: blockRank,
	}
}

// Rank returns the number of set bits (1s) in the range [0, i).
//
// The operation runs in O(1) time.
//
// Panics if i is out of bounds.
func (bi *BitIndex) Rank(i int) uint64 {
	bi.vec.checkBounds(i)

	wordIdx := i >> wordSizeLog
	offset := uint(i % wordSize)
	superIdx := i >> superBlockLog

	r := uint64(0)
	if superIdx < len(bi.superRank) {
		r = bi.superRank[superIdx]
	}
	r += uint64(bi.blockRank[wordIdx])
	mask := (uint64(1) << offset) - 1
	r += uint64(bits.OnesCount64(bi.vec.words[wordIdx] & mask))
	return r
}

// Select returns the position of the k-th set bit (1-based).
//
// That is, Select(1) returns the position of the first 1,
// Select(2) returns the position of the second 1, etc.
//
// The operation runs in:
//
//   - O(log (#superblocks)) +
//   - O(log wordsPerSuperblock) +
//   - O(popcount(word))
//
// In practice this is very fast since each superblock contains
// only 8 words (512 bits).
//
// Returns -1 if k exceeds the total number of set bits.
// Panics if k == 0.
func (bi *BitIndex) Select(k uint64) int {
	if k == 0 {
		panic("k must be >= 1")
	}

	var s int
	if len(bi.superRank) == 0 {
		return -1
	}
	// Find first superblock where cumulative rank >= k (lower_bound).
	idx := sort.Search(len(bi.superRank), func(i int) bool {
		return bi.superRank[i] >= k
	})
	// We need the superblock that strictly precedes k.
	if idx == 0 {
		s = 0
	} else {
		s = idx - 1
	}
	// Make k relative to the beginning of the selected superblock.
	k -= bi.superRank[s]

	nWords := len(bi.vec.words)
	wordsPerSB := superBlock >> wordSizeLog

	// Compute word range covered by the superblock.
	wStart := s << (superBlockLog - wordSizeLog)
	wEnd := min(wStart+wordsPerSB, nWords)
	if wStart >= wEnd {
		return -1
	}

	// Bin search inside the superblock over its words.
	off := sort.Search(wEnd-wStart, func(i int) bool {
		w := wStart + i
		// blockRank[w] is rank at word start (relative to superblock).
		return uint64(bi.blockRank[w])+uint64(bits.OnesCount64(bi.vec.words[w])) >= k
	})
	if off == wEnd-wStart {
		return -1
	}
	w := wStart + off

	// Make k relative to the beginning of the selected word.
	kInWord := k - uint64(bi.blockRank[w])
	word := bi.vec.words[w]

	for word != 0 {
		pos := bits.TrailingZeros64(word)
		if kInWord == 1 {
			return w<<wordSizeLog + pos
		}
		kInWord--
		word &= word - 1 // clear lowest set bit
	}

	return -1
}
