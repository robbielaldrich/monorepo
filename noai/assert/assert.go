package assert

import "testing"

func Equal[V comparable](t *testing.T, got, expected V) {
	t.Helper()

	if got != expected {
		t.Errorf("assertion error: got %v, expected %v", got, expected)
	}
}
