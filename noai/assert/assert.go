// Package assert provides test helpers for making assertions.
package assert

import "testing"

func Equal[V comparable](t *testing.T, got, expected V) {
	t.Helper()

	if got != expected {
		t.Errorf("assertion error: got %v, expected %v", got, expected)
	}
}

func True(b bool, falseDesc string) {
	if !b {
		panic(falseDesc)
	}
}

func NoError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
