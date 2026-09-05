package alg

import (
	"cmp"
	"math/rand"
)

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

func QuickSort[T cmp.Ordered](elements []T, opts ...QuickSortOpts) {
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
		median := FindMedian(elements)
		found := false
		for i, val := range elements {
			if val == median {
				found = true
				pivotIdx = i
				break
			}
		}
		assert(found, "median not found in elements")
	default:
		panic("pivot choice not yet implemented")
	}

	_ = pivotIdx
}
