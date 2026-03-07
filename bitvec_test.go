package bitvec

import "testing"

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
			name: "set within size",
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
			name: "set beyond initial size (dynamic expand)",
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
			name: "get out of bounds panics",
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
			name: "empty vector",
			bv: func() *BitVector {
				return NewVector(0, 0)
			},
			want: "",
		},
		{
			name: "all zeros",
			bv: func() *BitVector {
				return NewVector(5, 5)
			},
			want: "00000",
		},
		{
			name: "basic set",
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
			name: "flip and delete",
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
			name: "dynamic expand",
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
