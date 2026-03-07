package bitvec

import (
	"math/rand"
	"testing"
	"time"
)

func buildTestVector(bits []uint64) *BitVector {
	var max uint64
	for _, b := range bits {
		if b > max {
			max = b
		}
	}
	bv := NewVector(max+1, max+1)
	for _, b := range bits {
		bv.Set(b)
	}
	return bv
}

func TestRank(t *testing.T) {
	tests := []struct {
		name     string
		ones     []uint64
		checkPos []uint64
		expected []uint64
	}{
		{
			name:     "simple",
			ones:     []uint64{1, 3, 4, 7},
			checkPos: []uint64{0, 1, 2, 4, 7},
			expected: []uint64{0, 0, 1, 2, 3},
		},
		{
			name:     "dense",
			ones:     []uint64{0, 1, 2, 3, 4},
			checkPos: []uint64{0, 1, 4},
			expected: []uint64{0, 1, 4},
		},
		{
			name:     "single bit",
			ones:     []uint64{100},
			checkPos: []uint64{0, 99, 100},
			expected: []uint64{0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bv := buildTestVector(tt.ones)
			idx := NewIndex(bv)

			for i, pos := range tt.checkPos {
				got := idx.Rank(pos)
				if got != tt.expected[i] {
					t.Fatalf("Rank1(%d) = %d, want %d", pos, got, tt.expected[i])
				}
			}
		})
	}
}

func TestSelect(t *testing.T) {
	tests := []struct {
		name     string
		ones     []uint64
		expected []int
	}{
		{
			name:     "simple",
			ones:     []uint64{1, 3, 4, 7},
			expected: []int{1, 3, 4, 7},
		},
		{
			name:     "dense",
			ones:     []uint64{0, 1, 2, 3, 4},
			expected: []int{0, 1, 2, 3, 4},
		},
		{
			name:     "sparse",
			ones:     []uint64{10, 100, 1000},
			expected: []int{10, 100, 1000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bv := buildTestVector(tt.ones)
			idx := NewIndex(bv)

			for k := 1; k <= len(tt.expected); k++ {
				got := idx.Select(uint64(k))
				if got != tt.expected[k-1] {
					t.Fatalf("Select1(%d) = %d, want %d", k, got, tt.expected[k-1])
				}
			}
		})
	}
}

func TestRankSelectConsistency(t *testing.T) {
	ones := []uint64{2, 5, 9, 15, 33, 65, 130}
	bv := buildTestVector(ones)
	idx := NewIndex(bv)

	for k := 1; k <= len(ones); k++ {
		pos := idx.Select(uint64(k))
		r := idx.Rank(uint64(pos))
		if r != uint64(k-1) {
			t.Fatalf("Rank/Select mismatch: Select1(%d)=%d but Rank1=%d", k, pos, r)
		}
	}
}

func TestSelectOutOfRange(t *testing.T) {
	bv := buildTestVector([]uint64{1, 2, 3})
	idx := NewIndex(bv)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for Select1(0)")
		}
	}()
	idx.Select(0)
}

func TestWordBoundary(t *testing.T) {
	ones := []uint64{63, 64, 65}
	bv := buildTestVector(ones)
	idx := NewIndex(bv)

	for k, expected := range ones {
		if got := idx.Select(uint64(k + 1)); uint64(got) != expected {
			t.Fatalf("Select boundary failed: got %d want %d", got, expected)
		}
	}
}

func TestSuperBlockBoundary(t *testing.T) {
	ones := []uint64{510, 511, 512, 513}
	bv := buildTestVector(ones)
	idx := NewIndex(bv)

	for k, expected := range ones {
		if got := idx.Select(uint64(k + 1)); uint64(got) != expected {
			t.Fatalf("Select superblock failed: got %d want %d", got, expected)
		}
	}
}

func TestSelectTooLarge(t *testing.T) {
	ones := []uint64{1, 2, 3}
	bv := buildTestVector(ones)
	idx := NewIndex(bv)

	if got := idx.Select(4); got != -1 {
		t.Fatalf("expected -1 for too large k, got %d", got)
	}
}

const benchSize = 1_000_000

func buildSparseVector(n uint64, step uint64) *BitVector {
	bv := NewVector(n, n)
	var i uint64
	for ; i < n; i += step {
		bv.Set(i)
	}
	return bv
}

func buildDenseVector(n uint64) *BitVector {
	bv := NewVector(n, n)
	for i := range n {
		bv.Set(i)
	}
	return bv
}

func BenchmarkBuildIndexSparse(b *testing.B) {
	bv := buildSparseVector(benchSize, 7)

	for b.Loop() {
		_ = NewIndex(bv)
	}
}

func BenchmarkBuildIndexDense(b *testing.B) {
	bv := buildDenseVector(benchSize)

	for b.Loop() {
		_ = NewIndex(bv)
	}
}

func BenchmarkRankSparse(b *testing.B) {
	bv := buildSparseVector(benchSize, 7)
	idx := NewIndex(bv)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for b.Loop() {
		pos := rng.Intn(benchSize - 1)
		_ = idx.Rank(uint64(pos))
	}
}

func BenchmarkRankDense(b *testing.B) {
	bv := buildDenseVector(benchSize)
	idx := NewIndex(bv)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for b.Loop() {
		pos := rng.Intn(benchSize - 1)
		_ = idx.Rank(uint64(pos))
	}
}

func BenchmarkSelectSparse(b *testing.B) {
	bv := buildSparseVector(benchSize, 7)
	idx := NewIndex(bv)

	total := idx.Rank(benchSize - 1)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for b.Loop() {
		k := uint64(rng.Intn(int(total)) + 1)
		_ = idx.Select(k)
	}
}

func BenchmarkSelectDense(b *testing.B) {
	bv := buildDenseVector(benchSize)
	idx := NewIndex(bv)

	total := idx.Rank(benchSize - 1)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for b.Loop() {
		k := uint64(rng.Intn(int(total)) + 1)
		_ = idx.Select(k)
	}
}
