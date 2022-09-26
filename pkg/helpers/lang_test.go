package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolValueOrDefault(t *testing.T) {
	v := true
	b := BoolValueOrDefault(&v, false)
	assert.Equal(t, true, b)
}
