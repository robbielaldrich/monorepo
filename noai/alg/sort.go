package alg

import "math/rand"

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

func QuickSort[T comparable](elements []T, opts ...QuickSortOpts) {
	if len(elements) < 2 {
		return
	}

	o := getFirstOrDefault(opts)
	if o.PivotChoice == PivotChoiceNone {
		o.PivotChoice = PivotChoiceFirstElement
	}

	var pivotIdx int
	switch o.PivotChoice {
	case PivotChoiceNone:
		pivotIdx = 0
	case PivotChoiceFirstElement:
		pivotIdx = 0
	case PivotChoiceRandomElement:
		pivotIdx = rand.Intn(len(elements) - 1)
	case PivotChoiceMedianElement:
		pivotIdx, _ = FindMedian(elements)
	default:
		panic("pivot choice not yet implemented")
	}

	_ = pivotIdx
}
