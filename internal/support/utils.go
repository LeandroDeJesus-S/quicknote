package support

// TernaryIf returns t if the condition is true, f otherwise
func TernaryIf[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
}
