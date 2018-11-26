package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenize(t *testing.T) {
	var assertTypeTable = map[string]struct {
		expr   Expr
		result []string
	}{
		"value":            {"value", []string{"value"}},
		"response_object":  {"$response.value1.value2", []string{"value1", "value2"}},
		"response_map":     {"$response[\"value1\"][\"value2\"]", []string{"value1", "value2"}},
		"response_special": {"$response[\"value[1]\"][\"value.2\"].value3", []string{"value[1]", "value.2", "value3"}},
	}

	for name, table := range assertTypeTable {
		t.Run(name, func(t *testing.T) {
			result := table.expr.tokenize()
			assert.Equal(t, table.result, result)
		})
	}
}
