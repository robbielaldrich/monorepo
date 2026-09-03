package alg

type PivotChoice int

const (
	PivotChoiceNone PivotChoice = iota
	PivotChoiceFirstElement
	PivotChoiceRandomElement
	PivotChoiceMedianElement
)

type QuickSortOpts struct {
	PivotChoice PivotChoice
}

func QuickSort[T comparable](l []T, opts ...QuickSortOpts) {
	if len(l) < 2 {
		return
	}

	o := getFirstOrDefault(opts)
	if o.PivotChoice == PivotChoiceNone {
		o.PivotChoice = PivotChoiceFirstElement
	}

	var pivotIdx int
	switch o.PivotChoice {
	case PivotChoiceFirstElement:
		pivotIdx = 0
	case PivotChoiceNone:
		pivotIdx = 0
	}

	_ = pivotIdx
}

func getFirstOrDefault[T any](l []T) T {
	if len(l) == 0 {
		return *new(T)
	}

	return l[0]
}
