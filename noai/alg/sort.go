package alg

import (
	"cmp"
	"math/rand"
	"slices"
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

	o := firstOrNew(opts)
	if o.PivotChoice == PivotChoiceNone {
		o.PivotChoice = PivotChoiceFirstElement
	}

	var pivotIdx int
	var pivot T
	switch o.PivotChoice {
	case PivotChoiceNone:
		pivotIdx = 0
		pivot = elements[0]
	case PivotChoiceFirstElement:
		pivotIdx = 0
		pivot = elements[0]
	case PivotChoiceRandomElement:
		pivotIdx = rand.Intn(len(elements) - 1)
		pivot = elements[pivotIdx]
	case PivotChoiceMedianElement:
		pivot = FindMedian(elements)
		pivotIdx = slices.Index(elements, pivot)
	default:
		panic("pivot choice not yet implemented")
	}

	left, right := 0, len(elements)-1
	for left < right {
		// Advance left to an element greater than pivot, or equal.
		for elements[left] <= pivot {
			left++
		}

		// Decr right to an element less than pivot, or equal.
		for elements[right] >= pivot {
			right--
		}

		elements[left], elements[right] = elements[right], elements[left]
	}

	QuickSort(elements[left:pivotIdx], o)
	QuickSort(elements[pivotIdx+1:], o)
}
