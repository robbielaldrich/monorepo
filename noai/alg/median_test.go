package alg_test

import (
	"fmt"
	"testing"

	"github.com/robbielaldrich/monorepo/noai/alg"
	"github.com/robbielaldrich/monorepo/noai/assert"
)

func TestFindMedian(t *testing.T) {
	for algorithmName, algorithm := range map[string]alg.FindMedianAlgorithm{
		"SortElements": alg.FindMedianAlgorithmSortElements,
		"QuickSelect":  alg.FindMedianAlgorithmQuickSelect,
	} {
		testFindMedian(t, algorithmName, algorithm)
	}
}

func testFindMedian(t *testing.T, algorithmName string, algorithm alg.FindMedianAlgorithm) {
	for testName, test := range map[string]struct {
		elements       []int
		expectedMedian int
	}{
		"OneElement": {
			elements:       []int{8},
			expectedMedian: 8,
		},
		"TwoElements": {
			elements:       []int{8, 9},
			expectedMedian: 9,
		},
		"ThreeElements": {
			elements:       []int{8, 9, 11},
			expectedMedian: 9,
		},
		"ThreeElementsJumbled": {
			elements:       []int{11, 8, 9},
			expectedMedian: 9,
		},
	} {
		t.Run(fmt.Sprintf("%s_%s", algorithmName, testName), func(t *testing.T) {
			gotMedian := alg.FindMedian(test.elements, alg.FindMedianOpts{
				Algorithm: algorithm,
			})
			assert.Equal(t, gotMedian, test.expectedMedian)
		})
	}
}
