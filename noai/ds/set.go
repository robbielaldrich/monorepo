package ds

type Set[T comparable] struct {
	m map[T]struct{}
}

func NewSet[T comparable](values ...T) Set[T] {
	s := Set[T]{m: make(map[T]struct{}, len(values))}
	s.Add(values...)
	return s
}

func (s Set[T]) Add(values ...T) {
	for _, val := range values {
		s.m[val] = struct{}{}
	}
}

func (s Set[T]) Has(value T) bool {
	_, ok := s.m[value]
	return ok
}
