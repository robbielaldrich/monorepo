package alg

import (
	"cmp"
	"math/rand"
	"slices"
)

type FindMedianAlgorithm int

const (
	FindMedianAlgorithmNone FindMedianAlgorithm = iota
	FindMedianAlgorithmSortElements
	FindMedianAlgorithmQuickSelect // https://rcoh.me/posts/linear-time-median-finding
	FindMedianAlgorithmPick        // https://people.csail.mit.edu/rivest/pubs/BFPRT73.pdf
)

type FindMedianOpts struct {
	Algorithm FindMedianAlgorithm
}

func FindMedian[T cmp.Ordered](elements []T, opts ...FindMedianOpts) T {
	if len(elements) == 0 {
		panic("no elements")
	}

	o := firstOrNew(opts)
	if o.Algorithm == FindMedianAlgorithmNone {
		o.Algorithm = FindMedianAlgorithmQuickSelect
	}

	switch o.Algorithm {
	case FindMedianAlgorithmSortElements:
		slices.Sort(elements)
		return elements[len(elements)/2]
	case FindMedianAlgorithmQuickSelect:
		return quickSelectKthSmallestElement(elements, len(elements)/2)
	case FindMedianAlgorithmPick:
		return pickKthSmallestElement(elements, len(elements)/2)
	default:
		panic("not yet implemented")
	}
}

func quickSelectKthSmallestElement[T cmp.Ordered](elements []T, k int) T {
	if len(elements) == 1 {
		assert(k == 0, "unexpected k in base case")
		return elements[0]
	}

	pivot := elements[rand.Intn(len(elements))]

	elementsLessThanPivot := make([]T, 0, len(elements))
	elementsEqualToPivot := make([]T, 0, len(elements))
	elementsGreaterThanPivot := make([]T, 0, len(elements))

	for _, element := range elements {
		if element < pivot {
			elementsLessThanPivot = append(elementsLessThanPivot, element)
		} else if element == pivot {
			elementsEqualToPivot = append(elementsEqualToPivot, element)
		} else {
			elementsGreaterThanPivot = append(elementsGreaterThanPivot, element)
		}
	}

	nLess := len(elementsLessThanPivot)
	nEq := len(elementsEqualToPivot)
	if nLess > k {
		// kth smallest element is inside elementsLessThanPivot.
		return quickSelectKthSmallestElement(elementsLessThanPivot, k)
	} else if nLess+nEq > k {
		// kth smallest element is found in elementsEqualToPivot (so, it's pivot itself).
		return pivot
	} else {
		// kth smallest element is in elementsGreaterThanPivot.
		// Pass smaller k since we're jumping past elementsLessThanPivot and elementsEqualToPivot.
		return quickSelectKthSmallestElement(elementsGreaterThanPivot, k-(nLess+nEq))
	}
}

func pickKthSmallestElement[T cmp.Ordered](elements []T, k int) T {
	const (
		c = 5 // Smallest odd value that works (keeps O(N) run time).
		d = 3 // Position of median in sorted list of size c (value 5).
	)

	// Split elements into columns of size c.
	nOverC := len(elements) / c
	columns := make([][]T, 0, nOverC)

	columnIdx := -1
	for elementIdx, element := range elements {
		if elementIdx%c == 0 {
			columnIdx += 1
			columns = append(columns, make([]T, 0, c))
		}

		columns[columnIdx] = append(columns[columnIdx], element)
	}

	for _, column := range columns {
		// Sort the column (sorting a list of constant size takes constant time).
		slices.Sort(column)

		// Pick the middle element of the column.
		if len(elements) > 1

	}
}

func assert(b bool, msg string) {
	if !b {
		panic(msg)
	}
}
