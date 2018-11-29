package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSequentialImplementsBot(t *testing.T) {
	assert.Implements(t, (*Bot)(nil), new(SequentialBot))
}
