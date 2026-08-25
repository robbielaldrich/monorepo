package ds_test

import (
	"testing"

	"github.com/robbielaldrich/monorepo/noai/ds"
)

func TestSet(t *testing.T) {
	t.Run("StringInitEmpty", func(t *testing.T) {
		s := ds.NewSet[string]()
		s.Add("asdf")
		if !s.Has("asdf") {
			t.Fatal("has after add is false")
		}
		if s.Has("qwerty") {
			t.Fatal("has on non-added is true")
		}
	})

	t.Run("IntInitOneVal", func(t *testing.T) {
		s := ds.NewSet(11)
		s.Add(14)
		if !s.Has(11) {
			t.Fatal("has on init val fails")
		}
		if !s.Has(14) {
			t.Fatal("has after add fails")
		}
		if s.Has(15) {
			t.Fatal("has on non-added is true")
		}
	})

	t.Run("InitMultipleVals", func(t *testing.T) {
		s := ds.NewSet(0.23, -1.31)
		s.Add(1.48)
		if !s.Has(0.23) {
			t.Fatal("has on init val fails")
		}
		if !s.Has(-1.31) {
			t.Fatal("has on init val fails")
		}
		if !s.Has(1.48) {
			t.Fatal("has after add fails")
		}
		if s.Has(15) {
			t.Fatal("has on non-added is true")
		}
	})
}
