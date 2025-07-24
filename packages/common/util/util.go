package util

/*
	Ternary operator.
	If 'cond' is true then returns 'a', otherwise returns 'b'

	Usage-rules:

	1 - Do not nest this function: util.Ternary(cond1, util.Ternary(cond2, a, b), c) [ WRONG ]

	2 - Use it only with primitives: util.Ternary(newState == user.DeletedState, "IS NOT", "IS") [CORRECT]

	3 - Do not place function calls inside of it: util.Ternary(cond, calcA(), calcB()) [WRONG]
		Exceptions:
			- private fields getters: util.Ternary(cond, someStruct.GetA(), anotherStruct.GetB()) [CORRECT]
*/
func Ternary[T any](cond bool, a T, b T) T {
	if cond {
		return a
	}

	return b
}

/*
	This function will return zero-value of T if *v is nil, otherwise it will just dereference *v.

	Why does it needed?
	Cuz usually when you try to dereference *v you will get panic if it's <nil>.
*/
func SafeDereference[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}

