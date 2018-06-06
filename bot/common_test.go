package bot

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var castTable = map[string]struct {
	value  interface{}
	typ    string
	store  *storage
	result interface{}
	err    error
}{
	"success_int":             {1, "int", &storage{}, 1, nil},
	"success_string":          {"2", "string", &storage{}, "2", nil},
	"success_variable_int":    {"$var", "int", &storage{"var": 3}, 3, nil},
	"success_variable_string": {"$var2", "string", &storage{"var2": "4"}, "4", nil},
	"err_unknown_type":        {"5", "rand", &storage{}, nil, errors.New("Unknown type rand")},
	"err_string_to_int":       {"6", "int", &storage{}, nil, errors.New("Failed to cast to int")},
	"err_not_in_storage":      {"$var3", "string", &storage{"var4": "5"}, nil, errors.New("Variable var3 not found")},
}

var buildArgsTable = map[string]struct {
	rawArgs map[string]interface{}
	store   *storage
	result  map[string]interface{}
	err     error
}{
	"success_one": {map[string]interface{}{"playerId": map[string]interface{}{"type": "string", "value": "$playerId"}}, &storage{"playerId": "123456"}, map[string]interface{}{"playerId": "123456"}, nil},
	"success_multiple": {map[string]interface{}{
		"playerId": map[string]interface{}{"type": "string", "value": "$playerId"},
		"gold":     map[string]interface{}{"type": "int", "value": 10},
	}, &storage{"playerId": "123456"}, map[string]interface{}{"playerId": "123456", "gold": 10}, nil},
	"error_one": {map[string]interface{}{"playerId": map[string]interface{}{"type": "string", "value": "$playerId2"}}, &storage{"playerId": "123456"}, nil, errors.New("Variable playerId2 not found")},
}

func TestCast(t *testing.T) {
	for name, table := range castTable {
		t.Run(name, func(t *testing.T) {
			val, err := castType(table.value, table.typ, table.store)
			assert.Equal(t, table.result, val)
			assert.Equal(t, table.err, err)
		})
	}
}

func TestBuildArgs(t *testing.T) {
	for name, table := range buildArgsTable {
		t.Run(name, func(t *testing.T) {
			val, err := buildArgs(table.rawArgs, table.store)
			assert.Equal(t, table.result, val)
			assert.Equal(t, table.err, err)
		})
	}
}
