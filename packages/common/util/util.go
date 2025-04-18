package util

// Ternary operator.
// If 'cond' is true then returns 'a', otherwise returns 'b'
func Ternary[T any](cond bool, a T, b T) T {
	if cond {
		return a
	}

	return b
}

