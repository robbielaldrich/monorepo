package alg

func getFirstOrDefault[T any](l []T) T {
	if len(l) == 0 {
		return *new(T)
	}

	return l[0]
}
