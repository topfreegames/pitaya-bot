package bot

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/topfreegames/pitaya-bot/models"
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
		"err_int":           {"$", "int", nil, errors.New("Int type assertion failed for field: $")},
		"err_bool":          {"$", "bool", nil, errors.New("Boolean type assertion failed for field: $")},
		"err_string":        {1, "string", nil, errors.New("String type assertion failed for field: 1")},
		"err_unknown_type":  {"5", "rand", nil, errors.New("Unknown type rand")},
		"err_string_to_int": {"6", "int", nil, errors.New("Int type assertion failed for field: 6")},
	}

	for name, table := range assertTypeTable {
		t.Run(name, func(t *testing.T) {
			val, err := assertType(table.value, table.typ)
			assert.Equal(t, table.result, val)
			assert.Equal(t, table.err, err)
		})
	}
}

func TestBuildArgsWithStorage(t *testing.T) {
	var buildArgsWithStorageTable = map[string]struct {
		rawArgs map[string]interface{}
		store   *storage
		result  map[string]interface{}
		err     error
	}{
		"success_one": {map[string]interface{}{"playerId": map[string]interface{}{"type": "string", "value": "$store.playerId"}}, &storage{"playerId": "123456"}, map[string]interface{}{"playerId": "123456"}, nil},
		"success_multiple": {map[string]interface{}{
			"playerId": map[string]interface{}{"type": "string", "value": "$store.playerId"},
			"gold":     map[string]interface{}{"type": "int", "value": 10},
		}, &storage{"playerId": "123456"}, map[string]interface{}{"playerId": "123456", "gold": 10}, nil},
		"error_one": {map[string]interface{}{"playerId": map[string]interface{}{"type": "string", "value": "$store.playerId2"}}, &storage{"playerId": "123456"}, nil, errors.New("Variable playerId2 not found")},
	}

	for name, table := range buildArgsWithStorageTable {
		t.Run(name, func(t *testing.T) {
			val, err := buildArgs(table.rawArgs, table.store)
			assert.Equal(t, table.result, val)
			assert.Equal(t, table.err, err)
		})
	}
}

func TestBuildArgsUtils(t *testing.T) {
	var buildArgsUtilTable = map[string]struct {
		rawArgs   map[string]interface{}
		resultKey string
		regex     string
		err       error
	}{
		"success_util_uuid":    {map[string]interface{}{"playerId": map[string]interface{}{"type": "string", "value": "$util.uuid"}}, "playerId", "^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$", nil},
		"error_undefined_util": {map[string]interface{}{"playerId": map[string]interface{}{"type": "string", "value": "$util.unknown"}}, "", "", errors.New("util.unknown undefined")},
	}

	for name, table := range buildArgsUtilTable {
		t.Run(name, func(t *testing.T) {
			val, err := buildArgs(table.rawArgs, nil)
			assert.Equal(t, table.err, err)
			if table.err == nil {
				resultVal := val[table.resultKey].(string)
				valid := regexp.MustCompile(table.regex).MatchString(resultVal)
				assert.True(t, valid)
			}
		})
	}
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
		store     *storage
		resp      Response
		err       error
	}{
		"store_success":     {models.StoreSpec{"storeVal": models.StoreSpecEntry{Type: "string", Value: "val"}}, &storage{}, map[string]interface{}{"val": "value"}, nil},
		"store_error_token": {models.StoreSpec{"storeVal": models.StoreSpecEntry{Type: "string", Value: "val"}}, &storage{}, nil, errors.New("token 'val' not found")},
		"store_error_type":  {models.StoreSpec{"storeVal": models.StoreSpecEntry{Type: "string", Value: "val"}}, &storage{}, map[string]interface{}{"val": 1}, errors.New("String type assertion failed for field: 1")},
	}

	for name, table := range equalsTable {
		t.Run(name, func(t *testing.T) {
			fmt.Printf("table resp's type: %T\n", table.resp)

			var container interface{} = table.resp

			fmt.Printf("container's type: %T\n", container)
			err := storeData(table.storeSpec, table.store, table.resp)
			assert.Equal(t, table.err, err)
			if err == nil {
				val, ok := table.store.Get("storeVal")
				assert.True(t, ok)
				assert.Equal(t, val, "value")
			}
		})
	}
}
