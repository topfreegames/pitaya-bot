package bot

import (
	"encoding/json"
	"fmt"

	"github.com/topfreegames/pitaya-bot/models"
)

// ExpectError ...
type ExpectError struct {
	Err     error
	RawData []byte
	Expect  string
}

func (b *ExpectError) Error() string {
	return fmt.Sprintf("\nErr: %s \nRawData: %s \nExpected: %s\n", b.Err.Error(), string(b.RawData), b.Expect)
}

// NewExpectError ...
func NewExpectError(err error, rawData []byte, expect models.ExpectSpec) *ExpectError {
	bexpect, _ := json.Marshal(expect)
	return &ExpectError{
		Err:     err,
		RawData: rawData,
		Expect:  string(bexpect),
	}
}
