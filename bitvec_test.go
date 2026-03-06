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
			bv := NewVector(10)
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
