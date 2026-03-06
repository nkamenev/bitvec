package bitvec

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
func NewVector(n uint64) *BitVector {
	nWords := (n + wordSize - 1) / wordSize
	return &BitVector{
		words: make([]uint64, nWords),
		size:  n,
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
			newWords := make([]uint64, nWords)
			copy(newWords, bv.words)
			bv.words = newWords
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
