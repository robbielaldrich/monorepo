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
	FindMedianAlgorithmQuickSelect
	FindMedianAlgorithmPick // https://people.csail.mit.edu/rivest/pubs/BFPRT73.pdf
)

type FindMedianOpts struct {
	Algorithm FindMedianAlgorithm
}

func FindMedian[T cmp.Ordered](elements []T, opts ...FindMedianOpts) T {
	if len(elements) == 0 {
		panic("no elements")
	}

	o := getFirstOrDefault(opts)
	if o.Algorithm == FindMedianAlgorithmNone {
		o.Algorithm = FindMedianAlgorithmQuickSelect
	}

	switch o.Algorithm {
	case FindMedianAlgorithmSortElements:
		slices.Sort(elements)
		return elements[len(elements)/2]
	case FindMedianAlgorithmQuickSelect:
		return quickSelectKthSmallestElement(elements, len(elements)/2)
	default:
		panic("not yet implemented")
	}
}

// quickSelectKthSmallestElement references https://rcoh.me/posts/linear-time-median-finding.
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

func assert(b bool, msg string) {
	if !b {
		panic(msg)
	}
}
