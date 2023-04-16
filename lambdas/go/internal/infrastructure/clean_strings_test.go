package infrastructure

import (
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestCleanStrings(t *testing.T) {
	assert.Equal(t, CleanStrings("this_is_a-test"), "thisisatest")
}
