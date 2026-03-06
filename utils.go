package bitvec

func word(i uint64) uint64 {
	return i / wordSize
}

func offset(i uint64) uint64 {
	return uint64(i % wordSize)
}
