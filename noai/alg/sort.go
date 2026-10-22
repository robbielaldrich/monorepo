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

	o := firstOrNew(opts)
	if o.PivotChoice == PivotChoiceNone {
		o.PivotChoice = PivotChoiceFirstElement
	}

	var pivot T
	switch o.PivotChoice {
	case PivotChoiceNone:
		pivot = elements[0]
	case PivotChoiceFirstElement:
		pivot = elements[0]
	case PivotChoiceRandomElement:
		pivot = elements[rand.Intn(len(elements)-1)]
	case PivotChoiceMedianElement:
		pivot = FindMedian(elements)
	default:
		panic("pivot choice not yet implemented")
	}

	// todo
	left := 0
	for right := len(elements) - 1; right >= 0; right-- {
		if elements[right] < pivot {
			elements[right], elements[left] = elements[left], elements[right]
			left++
		}
	}

	elements[left], elements[len(elements)-1] = elements[left], elements[right]
}
