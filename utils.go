package bitvec

func word(i int) int {
	return i / wordSize
}

func offset(i int) uint {
	return uint(i % wordSize)
}
