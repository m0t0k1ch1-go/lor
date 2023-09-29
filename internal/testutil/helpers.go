package testutil

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Equal(t *testing.T, expected any, actual any, opts ...cmp.Option) {
	t.Helper()

	if diff := cmp.Diff(expected, actual, opts...); len(diff) > 0 {
		t.Errorf("diff: %s", diff)
	}
}
