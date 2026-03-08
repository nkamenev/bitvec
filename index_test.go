package bitvec

import (
	"math/rand"
	"sort"
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
				got, _ := idx.Select(uint64(k))
				if got != uint64(tt.expected[k-1]) {
					t.Fatalf("Select1(%d) = %d, want %d", k, got, tt.expected[k-1])
				}
			}
		})
	}
}

func TestTotalOnes(t *testing.T) {
	tests := []struct {
		name     string
		ones     []uint64
		expected uint64
	}{
		{
			name:     "empty_vector",
			ones:     []uint64{},
			expected: 0,
		},
		{
			name:     "single_bit",
			ones:     []uint64{42},
			expected: 1,
		},
		{
			name:     "multiple_bits",
			ones:     []uint64{1, 3, 5, 7, 9},
			expected: 5,
		},
		{
			name:     "bits_across_word_boundary",
			ones:     []uint64{63, 64, 65, 127, 128},
			expected: 5,
		},
		{
			name:     "bits_across_superblock_boundary",
			ones:     []uint64{500, 511, 512, 513, 1024},
			expected: 5,
		},
		{
			name:     "dense_bits",
			ones:     []uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bv := buildTestVector(tt.ones)
			idx := NewIndex(bv)

			got := idx.TotalOnes()
			if got != tt.expected {
				t.Errorf("TotalOnes() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestSelectNext(t *testing.T) {
	tests := []struct {
		name     string
		ones     []uint64
		pos      uint64
		expected uint64
		found    bool
	}{
		{
			name:     "exact_match_at_position",
			ones:     []uint64{1, 3, 5, 7, 9},
			pos:      3,
			expected: 3,
			found:    true,
		},
		{
			name:     "between_bits",
			ones:     []uint64{1, 3, 5, 7, 9},
			pos:      4,
			expected: 5,
			found:    true,
		},
		{
			name:     "before_first_bit",
			ones:     []uint64{10, 20, 30},
			pos:      5,
			expected: 10,
			found:    true,
		},
		{
			name:     "after_last_bit",
			ones:     []uint64{10, 20, 30},
			pos:      31,
			expected: 0,
			found:    false,
		},
		{
			name:     "exactly_at_last_bit",
			ones:     []uint64{10, 20, 30},
			pos:      30,
			expected: 30,
			found:    true,
		},
		{
			name:     "empty_vector",
			ones:     []uint64{},
			pos:      0,
			expected: 0,
			found:    false,
		},
		{
			name:     "single_bit_before_it",
			ones:     []uint64{100},
			pos:      50,
			expected: 100,
			found:    true,
		},
		{
			name:     "single_bit_after_it",
			ones:     []uint64{100},
			pos:      150,
			expected: 0,
			found:    false,
		},
		{
			name:     "position_at_word_boundary",
			ones:     []uint64{63, 64, 65},
			pos:      63,
			expected: 63,
			found:    true,
		},
		{
			name:     "position_just_after_word_boundary",
			ones:     []uint64{63, 64, 65},
			pos:      64,
			expected: 64,
			found:    true,
		},
		{
			name:     "position_at_superblock_boundary",
			ones:     []uint64{511, 512, 513},
			pos:      511,
			expected: 511,
			found:    true,
		},
		{
			name:     "position_just_after_superblock_boundary",
			ones:     []uint64{511, 512, 513},
			pos:      512,
			expected: 512,
			found:    true,
		},
		{
			name:     "position_beyond_vector_size",
			ones:     []uint64{1, 2, 3},
			pos:      1000,
			expected: 0,
			found:    false,
		},
		{
			name:     "dense_bits_exact_match",
			ones:     []uint64{0, 1, 2, 3, 4, 5},
			pos:      3,
			expected: 3,
			found:    true,
		},
		{
			name:     "dense_bits_between",
			ones:     []uint64{0, 1, 2, 3, 4, 5},
			pos:      2,
			expected: 2,
			found:    true,
		},
		{
			name:     "sparse_bits",
			ones:     []uint64{1000, 2000, 3000, 4000},
			pos:      2500,
			expected: 3000,
			found:    true,
		},
		{
			name:     "position_exactly_at_end_of_vector",
			ones:     []uint64{10, 20, 30},
			pos:      31, // vec size 31 (max bit 30 + 1)
			expected: 0,
			found:    false,
		},
		{
			name:     "position_at_last_bit_of_last_word",
			ones:     []uint64{63},
			pos:      63,
			expected: 63,
			found:    true,
		},
		{
			name:     "position_after_last_bit_of_last_word",
			ones:     []uint64{63},
			pos:      64,
			expected: 0,
			found:    false,
		},
		{
			name:     "multiple_bits_in_same_word",
			ones:     []uint64{10, 11, 12, 13, 14, 15},
			pos:      12,
			expected: 12,
			found:    true,
		},
		{
			name:     "between_bits_in_same_word",
			ones:     []uint64{10, 12, 14},
			pos:      11,
			expected: 12,
			found:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bv := buildTestVector(tt.ones)
			idx := NewIndex(bv)

			got, ok := idx.SelectNext(tt.pos)
			if ok != tt.found {
				t.Errorf("SelectNext(%d) found = %v, want %v", tt.pos, ok, tt.found)
			}
			if got != tt.expected {
				t.Errorf("SelectNext(%d) = %d, want %d", tt.pos, got, tt.expected)
			}
		})
	}
}

func TestSelectNextConsistency(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	ones := make([]uint64, 100)
	for i := range ones {
		ones[i] = uint64(rand.Intn(10000))
	}
	unique := make(map[uint64]bool)
	for _, pos := range ones {
		unique[pos] = true
	}
	sorted := make([]uint64, 0, len(unique))
	for pos := range unique {
		sorted = append(sorted, pos)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	bv := buildTestVector(sorted)
	idx := NewIndex(bv)

	t.Run("consistency with Rank/Select", func(t *testing.T) {
		for _, pos := range sorted {
			next, ok := idx.SelectNext(pos)
			if !ok || next != pos {
				t.Errorf("SelectNext(%d) = %d, want %d", pos, next, pos)
			}

			r := idx.Rank(pos)
			if r < idx.TotalOnes() {
				nextSelect, _ := idx.Select(r + 1)
				if nextSelect != pos {
					t.Errorf("Rank/Select mismatch: Rank(%d)=%d, Select(%d)=%d",
						pos, r, r+1, nextSelect)
				}
			}
		}
	})

	t.Run("consistency between positions", func(t *testing.T) {
		for i := 0; i < len(sorted)-1; i++ {
			current := sorted[i]
			next := sorted[i+1]

			if next > current+1 {
				mid := current + 1
				got, ok := idx.SelectNext(mid)
				if !ok || got != next {
					t.Errorf("SelectNext(%d) = %d, want %d", mid, got, next)
				}
			}
		}
	})
}

func TestRankSelectConsistency(t *testing.T) {
	ones := []uint64{2, 5, 9, 15, 33, 65, 130}
	bv := buildTestVector(ones)
	idx := NewIndex(bv)

	for k := 1; k <= len(ones); k++ {
		pos, _ := idx.Select(uint64(k))
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
		if got, _ := idx.Select(uint64(k + 1)); uint64(got) != expected {
			t.Fatalf("Select boundary failed: got %d want %d", got, expected)
		}
	}
}

func TestSuperBlockBoundary(t *testing.T) {
	ones := []uint64{510, 511, 512, 513}
	bv := buildTestVector(ones)
	idx := NewIndex(bv)

	for k, expected := range ones {
		if got, _ := idx.Select(uint64(k + 1)); uint64(got) != expected {
			t.Fatalf("Select superblock failed: got %d want %d", got, expected)
		}
	}
}

func TestSelectTooLarge(t *testing.T) {
	ones := []uint64{1, 2, 3}
	bv := buildTestVector(ones)
	idx := NewIndex(bv)

	if got, ok := idx.Select(4); ok {
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
		_, _ = idx.Select(k)
	}
}

func BenchmarkSelectDense(b *testing.B) {
	bv := buildDenseVector(benchSize)
	idx := NewIndex(bv)

	total := idx.Rank(benchSize - 1)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for b.Loop() {
		k := uint64(rng.Intn(int(total)) + 1)
		_, _ = idx.Select(k)
	}
}
