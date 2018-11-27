package bot

import (
	"testing"
)

func TestSequentialImplementsBot(t *testing.T) {
	var _ Bot = (*SequentialBot)(nil)
}
