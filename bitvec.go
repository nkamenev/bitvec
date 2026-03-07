package bitvec

import (
	"fmt"
	"strings"
)

const (
	wordSize    = 64
	wordSizeLog = 6
)

// BitVector represents a simple dynamic bit vector.
// Bits are stored in a slice of uint64 words. Supports Set/Delete/Get/Flip/Clear.
type BitVector struct {
	words []uint64
	size  uint64
}

// NewVector creates a new BitVector with n bits (all zeroed).
func NewVector(size uint64, cp uint64) *BitVector {
	if cp < size {
		cp = size
	}
	nWords := (size + wordSize - 1) / wordSize
	cpWords := (cp + wordSize - 1) / wordSize
	return &BitVector{
		words: make([]uint64, nWords, cpWords),
		size:  size,
	}
}

// NewVectorFromWords creates a BitVector from an existing slice of words.
// The slice is copied, so the original can be modified safely.
func NewVectorFromWords(words []uint64) *BitVector {
	size := len(words) * wordSize
	return &BitVector{
		words: append([]uint64(nil), words...),
		size:  uint64(size),
	}
}

// Set sets the bit at index i to 1.
// If i >= current size, the vector is automatically expanded to fit.
func (bv *BitVector) Set(i uint64) {
	w := word(i)

	if i >= bv.size {
		nWords := w + 1

		if nWords > uint64(len(bv.words)) {
			if nWords <= uint64(cap(bv.words)) {
				// We have enough capacity — just extend the slice.
				bv.words = bv.words[:nWords]
			} else {
				// Capacity is insufficient — allocate a new backing array.
				newWords := make([]uint64, nWords)
				copy(newWords, bv.words)
				bv.words = newWords
			}
		}

		bv.size = i + 1
	}

	bv.words[w] |= 1 << offset(i)
}

// Delete clears the bit at index i (sets to 0).
// Panics if i is out of range.
func (bv *BitVector) Delete(i uint64) {
	bv.checkBounds(i)
	bv.words[word(i)] &^= 1 << offset(i)
}

// Get returns the value of the bit at index i (true = 1, false = 0).
// Panics if i is out of range.
func (bv *BitVector) Get(i uint64) bool {
	bv.checkBounds(i)
	return (bv.words[word(i)]>>offset(i))&1 == 1
}

// Flip inverts the bit at index i (0 → 1, 1 → 0).
func (bv *BitVector) Flip(i uint64) {
	bv.checkBounds(i)
	bv.words[word(i)] ^= 1 << offset(i)
}

// Size returns the number of bits currently represented by the BitVector.
// Note that the underlying storage may be larger due to word alignment.
func (bv *BitVector) Size() uint64 {
	return bv.size
}

// Clear resets all bits to 0.
func (bv *BitVector) Clear() {
	for i := range bv.words {
		bv.words[i] = 0
	}
}

func (bv *BitVector) checkBounds(i uint64) {
	if i >= bv.size {
		panic("index out of range")
	}
}

func (bv *BitVector) String() string {
	if bv.size == 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(int(bv.size))

	for i := uint64(0); i < bv.size; i++ {
		w := i >> wordSizeLog
		o := i & (wordSize - 1)

		if (bv.words[w]>>o)&1 == 1 {
			b.WriteByte('1')
		} else {
			b.WriteByte('0')
		}
	}

	return b.String()
}

func (bv *BitVector) StringWords() string {
	if bv.size == 0 {
		return ""
	}

	ws := uint64(wordSize)

	var sb strings.Builder
	for i, w := range bv.words {
		bitsInWord := ws
		if i == len(bv.words)-1 && bv.size%wordSize != 0 {
			bitsInWord = bv.size % wordSize
		}
		s := fmt.Sprintf("%0*b", bitsInWord, w)
		sb.WriteString(s)
		sb.WriteByte('\n')
	}
	return sb.String()
}

// Merge appends bits from the provided vectors to the current BitVector.
// Existing bits remain unchanged, and new bits are added after them.
//
// The underlying storage is reused if capacity allows; otherwise it is reallocated.
func (bv *BitVector) Merge(vectors ...*BitVector) {
	var totalNewBits uint64
	for _, v := range vectors {
		if v != nil {
			totalNewBits += v.size
		}
	}
	if totalNewBits == 0 {
		return
	}

	newSz := bv.size + totalNewBits
	totalWs := (newSz + wordSize - 1) / wordSize

	// Ensure enough capacity in the underlying slice.
	if uint64(cap(bv.words)) < totalWs {
		newWords := make([]uint64, totalWs)
		copy(newWords, bv.words)
		bv.words = newWords
	} else if totalWs > uint64(len(bv.words)) {
		bv.words = bv.words[:totalWs]
	}

	offset := bv.size
	for _, v := range vectors {
		if v == nil || v.size == 0 {
			continue
		}
		copyWordsWithOffset(bv.words, offset, v)
		offset += v.size
	}

	bv.size = newSz
}

// Merge creates a new BitVector containing bits from all provided vectors
// concatenated in the given order.
//
// Nil vectors are ignored.
func Merge(vectors ...*BitVector) *BitVector {
	var totalBits uint64
	for _, v := range vectors {
		if v != nil {
			totalBits += v.size
		}
	}

	if totalBits == 0 {
		return NewVector(0, 0)
	}

	totalWs := (totalBits + wordSize - 1) / wordSize
	res := &BitVector{
		words: make([]uint64, totalWs),
		size:  totalBits,
	}

	var offset uint64
	for _, v := range vectors {
		copyWordsWithOffset(res.words, offset, v)
		if v != nil {
			offset += v.size
		}
	}

	return res
}

// copyWordsWithOffset copies bits from src into dstWords starting at dstOffset.
// Bits may need to be shifted if dstOffset is not word-aligned.
//
// This function performs word-level merging and handles carry-over
// into the next word when shifts cross the 64-bit boundary.
func copyWordsWithOffset(dstWords []uint64, dstOffset uint64, src *BitVector) {
	if src == nil || src.size == 0 {
		return
	}

	nSrcWords := uint64(len(src.words))
	lastBits := src.size & (wordSize - 1) // number of valid bits in the last word

	for i := range nSrcWords {
		w := src.words[i]

		// Mask out unused bits in the last source word.
		if i == nSrcWords-1 && lastBits != 0 {
			w &= (1 << lastBits) - 1
		}

		dstIdx := (dstOffset >> wordSizeLog) + i
		shift := dstOffset & (wordSize - 1)

		if shift == 0 {
			// Fast path: word-aligned copy.
			dstWords[dstIdx] |= w
		} else {
			// Split the shifted word across two destination words.
			dstWords[dstIdx] |= w << shift
			if dstIdx+1 < uint64(len(dstWords)) {
				dstWords[dstIdx+1] |= w >> (wordSize - shift)
			}
		}
	}
}
