package domain

import (
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestStringToBounds(t *testing.T) {
	assert.Equal(t, StringToBounds("UkraineAndAround"), UkraineAndAround)
	assert.Equal(t, StringToBounds("ukraine_and_around"), UkraineAndAround)
	assert.Equal(t, StringToBounds("fake"), BoundsUnspecified)
}
