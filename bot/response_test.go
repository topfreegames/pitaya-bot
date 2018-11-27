package bot

import (
	"fmt"
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

func TestTryExtractValue(t *testing.T) {
	var assertTypeTable = map[string]struct {
		resp     interface{}
		expr     Expr
		exprType string
		result   interface{}
		err      error
	}{
		"success_value":  {"value", "", "string", "value", nil},
		"success_map":    {map[string]interface{}{"attr": "value"}, "$response.attr", "string", "value", nil},
		"success_object": {map[string]interface{}{"attr": "value"}, "$response[\"attr\"]", "string", "value", nil},
		"err_value":      {"value", "$response.attr1.attr2", "string", nil, fmt.Errorf("malformed spec file. expr $response.attr1.attr2 doesn't match the object received")},
		"err_map":        {map[string]interface{}{"attr": "value"}, "$response[\"err\"]", "string", nil, fmt.Errorf("token 'err' not found within expr $response[\"err\"]")},
		"err_object":     {map[string]interface{}{"attr": "value"}, "$response.err", "string", nil, fmt.Errorf("token 'err' not found within expr $response.err")},
		"err_slice1":     {[]interface{}{"value"}, "$response.err", "string", nil, fmt.Errorf("malformed spec file. expr $response.err, 'err' token must be an index")},
		"err_slice2":     {[]interface{}{"value"}, "$response[9]", "string", nil, fmt.Errorf("token index 9 not available within expr $response[9]")},
	}

	for name, table := range assertTypeTable {
		t.Run(name, func(t *testing.T) {
			result, err := tryExtractValue(table.resp, table.expr, table.exprType)
			assert.Equal(t, table.result, result)
			assert.Equal(t, table.err, err)
		})
	}
}
