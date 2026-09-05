package alg_test

import (
	"testing"

	"github.com/robbielaldrich/monorepo/noai/alg"
	"github.com/robbielaldrich/monorepo/noai/assert"
)

func TestFindMedian(t *testing.T) {
	for name, test := range map[string]struct {
		algorithm      alg.FindMedianAlgorithm
		elements       []int
		expectedMedian int
	}{
		"QuickSelectOneElement": {
			algorithm:      alg.FindMedianAlgorithmQuickSelect,
			elements:       []int{8},
			expectedMedian: 8,
		},
	} {
		t.Run(name, func(t *testing.T) {
			gotMedian := alg.FindMedian(test.elements, alg.FindMedianOpts{
				Algorithm: test.algorithm,
			})
			assert.Equal(t, gotMedian, test.expectedMedian)
		})
	}
}
