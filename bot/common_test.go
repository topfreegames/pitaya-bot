package bot

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/topfreegames/pitaya-bot/models"
	"github.com/topfreegames/pitaya-bot/storage"
)

func TestAssertType(t *testing.T) {
	var assertTypeTable = map[string]struct {
		value  interface{}
		typ    string
		result interface{}
		err    error
	}{
		"success_int":       {1, "int", 1, nil},
		"success_string":    {"2", "string", "2", nil},
		"success_bool":      {true, "bool", true, nil},
		"err_int":           {"$", "int", nil, errors.New("int type assertion failed for field: $")},
		"err_bool":          {"$", "bool", nil, errors.New("bool type assertion failed for field: $")},
		"err_string":        {1, "string", nil, errors.New("string type assertion failed for field: 1")},
		"err_unknown_type":  {"5", "rand", nil, errors.New("Unknown type rand")},
		"err_string_to_int": {"6", "int", nil, errors.New("int type assertion failed for field: 6")},
	}

	for name, table := range assertTypeTable {
		t.Run(name, func(t *testing.T) {
			val, err := assertType(table.value, table.typ)
			assert.Equal(t, table.result, val)
			assert.Equal(t, table.err, err)
		})
	}
}

func TestBuildArgsByType(t *testing.T) {
	var buildArgsWithStorageTable = map[string]struct {
		value     interface{}
		valueType string
		store     storage.Storage
		result    interface{}
		err       error
	}{
		"success_one": {map[string]interface{}{"playerId": map[string]interface{}{"type": "string", "value": "$store.playerId"}}, "object", &storage.MemoryStorage{"playerId": "123456"}, map[string]interface{}{"playerId": "123456"}, nil},
		"success_multiple": {map[string]interface{}{
			"playerId": map[string]interface{}{"type": "string", "value": "$store.playerId"},
			"gold":     map[string]interface{}{"type": "int", "value": 10},
		}, "object", &storage.MemoryStorage{"playerId": "123456"}, map[string]interface{}{"playerId": "123456", "gold": 10}, nil},
		"error_one":            {map[string]interface{}{"playerId": map[string]interface{}{"type": "string", "value": "$store.playerId2"}}, "object", &storage.MemoryStorage{"playerId": "123456"}, nil, errors.New("storage key not found")},
		"error_undefined_util": {map[string]interface{}{"playerId": map[string]interface{}{"type": "string", "value": "$util.unknown"}}, "object", nil, nil, errors.New("util.unknown undefined")},
	}

	for name, table := range buildArgsWithStorageTable {
		t.Run(name, func(t *testing.T) {
			val, err := buildArgByType(table.value, table.valueType, table.store)
			assert.Equal(t, table.result, val)
			assert.Equal(t, table.err, err)
		})
	}
}

func TestUUIDValueFromUtil(t *testing.T) {
	regex := "^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"
	rawArgs := map[string]interface{}{"playerId": map[string]interface{}{"type": "string", "value": "$util.uuid"}}
	val, err := buildArgByType(rawArgs, "object", nil)
	assert.NoError(t, err)
	resultVal := val.(map[string]interface{})["playerId"].(string)
	assert.True(t, regexp.MustCompile(regex).MatchString(resultVal))
}

func TestEquals(t *testing.T) {
	var equalsTable = map[string]struct {
		value1 interface{}
		value2 interface{}
		result bool
	}{
		"string_true":    {"string", "string", true},
		"string_false1":  {"string1", "string2", false},
		"string_false2":  {"string", 1, false},
		"int_true":       {1, 1, true},
		"int_false1":     {1, 2, false},
		"int_false2":     {1, "string", false},
		"bool_true":      {true, true, true},
		"bool_false1":    {true, false, false},
		"bool_false2":    {true, 1, false},
		"unknown_false1": {map[string]bool{}, "int", false},
		"unknown_false2": {"int", map[string]bool{}, false},
	}

	for name, table := range equalsTable {
		t.Run(name, func(t *testing.T) {
			val := equals(table.value1, table.value2)
			assert.Equal(t, table.result, val)
		})
	}
}

func TestStoreData(t *testing.T) {
	var equalsTable = map[string]struct {
		storeSpec models.StoreSpec
		store     storage.Storage
		response  interface{}
		err       error
	}{
		"store_value_success": {models.StoreSpec{"storeVal": models.StoreSpecEntry{Type: "string", Value: "val"}}, &storage.MemoryStorage{}, map[string]interface{}{"val": "value"}, nil},
		"store_map_success":   {models.StoreSpec{"storeVal": models.StoreSpecEntry{Type: "string", Value: "$response.val[\"valMap\"]"}}, &storage.MemoryStorage{}, map[string]interface{}{"val": map[string]interface{}{"valMap": "value"}}, nil},
		"store_slice_success": {models.StoreSpec{"storeVal": models.StoreSpecEntry{Type: "string", Value: "$response.val[0]"}}, &storage.MemoryStorage{}, map[string]interface{}{"val": []interface{}{"value"}}, nil},
		"store_error_token":   {models.StoreSpec{"storeVal": models.StoreSpecEntry{Type: "string", Value: "val"}}, &storage.MemoryStorage{}, nil, errors.New("string type assertion failed for field: <nil>")},
		"store_error_type":    {models.StoreSpec{"storeVal": models.StoreSpecEntry{Type: "string", Value: "val"}}, &storage.MemoryStorage{}, map[string]interface{}{"val": 1}, errors.New("string type assertion failed for field: 1")},
	}

	for name, table := range equalsTable {
		t.Run(name, func(t *testing.T) {
			err := storeData(table.storeSpec, table.store, table.response)
			assert.Equal(t, table.err, err)
			if err == nil {
				val, err := table.store.Get("storeVal")
				assert.NoError(t, err)
				assert.Equal(t, val, "value")
			}
		})
	}
}
