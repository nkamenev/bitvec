package bitvec

import (
	"testing"
)

func TestBitVectorBasic(t *testing.T) {
	type exp struct {
		i        uint64
		expected bool
	}
	tests := []struct {
		name   string
		ops    func(bv *BitVector)
		checks []exp
	}{
		{
			name: "set_within_size",
			ops: func(bv *BitVector) {
				bv.Set(0)
				bv.Set(3)
				bv.Set(5)
			},
			checks: []exp{
				{0, true},
				{1, false},
				{2, false},
				{3, true},
				{5, true},
			},
		},
		{
			name: "set_beyond_initial_size_(dynamic expand)",
			ops: func(bv *BitVector) {
				bv.Set(100)
				bv.Set(64)
			},
			checks: []exp{
				{64, true},
				{100, true},
				{99, false},
			},
		},
		{
			name: "flip",
			ops: func(bv *BitVector) {
				bv.Set(2)
				bv.Flip(2)
				bv.Flip(1)
			},
			checks: []exp{
				{2, false},
				{1, true},
			},
		},
		{
			name: "delete",
			ops: func(bv *BitVector) {
				bv.Set(2)
				bv.Set(10)
				bv.Delete(2)
			},
			checks: []exp{
				{2, false},
				{10, true},
			},
		},
		{
			name: "get_out_of_bounds_panics",
			ops: func(bv *BitVector) {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic for Get out of bounds")
					}
				}()
				_ = bv.Get(1000)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bv := NewVector(10, 10)
			if tt.ops != nil {
				tt.ops(bv)
			}
			for _, check := range tt.checks {
				got := bv.Get(check.i)
				if got != check.expected {
					t.Errorf("get(%d) = %v, want %v", check.i, got, check.expected)
				}
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name string
		bv   func() *BitVector
		want string
	}{
		{
			name: "empty_vector",
			bv: func() *BitVector {
				return NewVector(0, 0)
			},
			want: "",
		},
		{
			name: "all_zeros",
			bv: func() *BitVector {
				return NewVector(5, 5)
			},
			want: "00000",
		},
		{
			name: "basic_set",
			bv: func() *BitVector {
				bv := NewVector(8, 8)
				bv.Set(0)
				bv.Set(3)
				bv.Set(7)
				return bv
			},
			want: "10010001",
		},
		{
			name: "flip_and_delete",
			bv: func() *BitVector {
				bv := NewVector(5, 5)
				bv.Set(1)
				bv.Set(2)
				bv.Flip(2)
				bv.Flip(3)
				bv.Delete(1)
				return bv
			},
			want: "00010",
		},
		{
			name: "dynamic_expand",
			bv: func() *BitVector {
				bv := NewVector(4, 4)
				bv.Set(6)
				return bv
			},
			want: "0000001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bv := tt.bv()
			got := bv.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
func wantStr(sz int, ones ...int) string {
	s := make([]byte, sz)
	for i := range s {
		s[i] = '0'
	}
	for _, bit := range ones {
		if bit >= sz {
			panic("bit position out of range")
		}
		s[bit] = '1'
	}
	return string(s)
}

func TestMergeFunc(t *testing.T) {
	tests := []struct {
		name    string
		vecs    []*BitVector
		wantStr string
	}{
		{
			name:    "empty",
			vecs:    nil,
			wantStr: "",
		},
		{
			name: "single_empty_vector",
			vecs: []*BitVector{
				NewVector(0, 0),
			},
			wantStr: "",
		},
		{
			name: "single_vector",
			vecs: []*BitVector{
				func() *BitVector {
					v := NewVector(5, 5)
					v.Set(1)
					v.Set(3)
					return v
				}(),
			},
			wantStr: "01010",
		},
		{
			name: "two_vectors_aligned",
			vecs: []*BitVector{
				func() *BitVector {
					v := NewVector(4, 4)
					v.Set(0)
					v.Set(2)
					return v // 1010
				}(),
				func() *BitVector {
					v := NewVector(4, 4)
					v.Set(1)
					v.Set(3)
					return v // 0101
				}(),
			},
			wantStr: "10100101",
		},
		{
			name: "two_vectors_with_different_cap",
			vecs: []*BitVector{
				func() *BitVector {
					v := NewVector(4, 10)
					v.Set(0)
					v.Set(2)
					return v // 1010
				}(),
				func() *BitVector {
					v := NewVector(4, 4)
					v.Set(1)
					v.Set(3)
					return v // 0101
				}(),
			},
			wantStr: "10100101",
		},
		{
			name: "two_vectors_unaligned",
			vecs: []*BitVector{
				func() *BitVector {
					v := NewVector(70, 70)
					v.Set(0)
					v.Set(65)
					return v
				}(),
				func() *BitVector {
					v := NewVector(70, 70)
					v.Set(1)
					v.Set(66)
					return v
				}(),
			},
			wantStr: wantStr(
				140,
				0, 65, 71, 136,
			),
		},
		{
			name: "three_vectors_different_sizes",
			vecs: []*BitVector{
				func() *BitVector {
					v := NewVector(3, 3)
					v.Set(0)
					return v // 100
				}(),
				func() *BitVector {
					v := NewVector(5, 5)
					v.Set(2)
					return v // 0010
				}(),
				func() *BitVector {
					v := NewVector(4, 4)
					v.Set(3)
					return v // 0001
				}(),
			},
			wantStr: "100001000001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Merge(tt.vecs...)
			if got.String() != tt.wantStr {
				t.Errorf("Merge() = \n%swant %s\n", got.String(), tt.wantStr)
			}
		})
	}
}

func TestMergeMethod(t *testing.T) {
	tests := []struct {
		name          string
		init          *BitVector
		toMerge       []*BitVector
		wantStr       string
		expectCapSame bool
	}{
		{
			name: "merge_into_empty_vector",
			init: NewVector(0, 10),
			toMerge: []*BitVector{
				func() *BitVector {
					v := NewVector(5, 5)
					v.Set(1)
					return v
				}(),
			},
			wantStr:       "01000",
			expectCapSame: true,
		},
		{
			name: "merge_with_enough_capacity",
			init: func() *BitVector {
				v := NewVector(3, 10)
				v.Set(0)
				return v
			}(),
			toMerge: []*BitVector{
				func() *BitVector {
					v := NewVector(2, 2)
					v.Set(1)
					return v
				}(),
			},
			wantStr:       "10001",
			expectCapSame: true,
		},
		{
			name: "merge_requires_realloc",
			init: func() *BitVector {
				v := NewVector(5, 5)
				v.Set(0)
				return v
			}(),
			toMerge: []*BitVector{
				func() *BitVector {
					v := NewVector(70, 70)
					v.Set(1)
					return v
				}(),
			},
			wantStr: wantStr(
				75,
				0, 6,
			),
			expectCapSame: false,
		},
		{
			name: "merge_multiple_vectors_with_realloc",
			init: NewVector(1, 1),
			toMerge: []*BitVector{
				func() *BitVector {
					v := NewVector(63, 63)
					v.Set(0)
					return v
				}(),
				func() *BitVector {
					v := NewVector(63, 63)
					v.Set(0)
					return v
				}(),
			},
			wantStr: wantStr(
				127,
				1, 64,
			),
			expectCapSame: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldCap := cap(tt.init.words)
			tt.init.Merge(tt.toMerge...)
			gotStr := tt.init.String()
			if gotStr != tt.wantStr {
				t.Errorf("Merge() = %s, want %s", gotStr, tt.wantStr)
			}
			newCap := cap(tt.init.words)
			if tt.expectCapSame && newCap != oldCap {
				t.Errorf("expected capacity to stay the same, but changed from %d to %d", oldCap, newCap)
			}
			if !tt.expectCapSame && newCap <= oldCap {
				t.Errorf("expected capacity to grow, but it stayed %d", newCap)
			}
		})
	}
}
