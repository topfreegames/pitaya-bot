package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/topfreegames/pitaya-bot/models"
)

func initializeDb(store *storage) error {
	return nil
}

func valueFromUtil(fName string) (interface{}, error) {
	switch fName {
	case "uuid":
		return uuid.New().String(), nil
	default:
		return nil, fmt.Errorf("util.%s undefined", fName)
	}
}

func tryGetValue(expr interface{}, store *storage) (interface{}, error) {
	if val, ok := expr.(string); ok {
		if strings.HasPrefix(val, "$store") {
			variable := val[7:]
			if val, ok := store.Get(variable); ok {
				return val, nil
			}

			return nil, fmt.Errorf("Variable %s not found", variable)
		}

		if strings.HasPrefix(val, "$util") {
			f := val[6:]
			return valueFromUtil(f)
		}
	}

	return nil, nil
}

func assertType(value interface{}, typ string) (interface{}, error) {
	ret := value
	switch typ {
	case "string":
		if val, ok := ret.(string); ok {
			ret = val
		} else {
			return nil, fmt.Errorf("String type assetion failed for filed: %v", ret)
		}
	case "int":
		t := reflect.TypeOf(ret)
		switch t.Kind() {
		case reflect.Int:
			if val, ok := ret.(int); ok {
				ret = val
			} else {
				return nil, fmt.Errorf("Int type assetion failed for filed: %v", ret)
			}

		case reflect.Float64:
			if val, ok := ret.(float64); ok {
				ret = int(val)
			} else {
				return nil, fmt.Errorf("Int type assetion failed for filed: %v", ret)
			}
		default:
			return nil, fmt.Errorf("Int type assertion failed for field: %v", ret)
		}
	default:
		return nil, fmt.Errorf("Unknown type %s", typ)
	}

	return ret, nil
}

func buildArgByType(value interface{}, valueType string, store *storage) (interface{}, error) {
	var err error
	switch valueType {
	case "object":
		arg, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.New("Malformed object type argument")
		}

		preparedArgs := map[string]interface{}{}
		for key, params := range arg {
			p := params.(map[string]interface{})

			valueFromStorage, err := tryGetValue(p["value"], store)
			if err != nil {
				return nil, err
			}

			paramType := p["type"].(string)
			var paramValue interface{}

			if valueFromStorage != nil {
				paramValue = valueFromStorage
			} else {
				paramValue = p["value"]
			}

			builtParam, err := buildArgByType(paramValue, paramType, store)
			if err != nil {
				return nil, err
			}
			preparedArgs[key] = builtParam
		}

		return preparedArgs, nil
	case "array":
		fmt.Println("buildArgByType for array Not implemented")
	default:
		value, err = assertType(value, valueType)
		if err != nil {
			return nil, err
		}
	}

	return value, nil
}

func buildArgs(rawArgs map[string]interface{}, store *storage) (map[string]interface{}, error) {
	args, err := buildArgByType(rawArgs, "object", store)
	if err != nil {
		return nil, err
	}

	r := args.(map[string]interface{})
	return r, nil
}

func sendRequest(args map[string]interface{}, route string, pclient *PClient) (Response, []byte, error) {
	encodedData, err := json.Marshal(args)
	if err != nil {
		return nil, nil, err
	}

	return pclient.Request(route, encodedData)
}

func getValueFromSpec(spec models.ExpectSpecEntry, store *storage) (interface{}, error) {
	value, err := tryGetValue(spec.Value, store)
	if err != nil {
		return nil, err
	}

	if value == nil {
		value, err = assertType(spec.Value, spec.Type)
		if err != nil {
			return nil, err
		}
	}

	return value, nil
}

func validateExpectations(expectations models.ExpectSpec, resp Response, store *storage) error {
	for propertyExpr, spec := range expectations {
		expectedValue, err := getValueFromSpec(spec, store)
		if err != nil {
			return err
		}

		gotValue, err := Response(resp).extractValue(Expr(propertyExpr), spec.Type)
		if err != nil {
			return err
		}

		if !equals(expectedValue, gotValue) {
			return fmt.Errorf("%v != %v", expectedValue, gotValue)
		}
	}

	return nil
}

func equals(lhs interface{}, rhs interface{}) bool {
	t := reflect.TypeOf(lhs)

	switch t.Kind() {
	case reflect.String:
		lhsVal := lhs.(string)
		rhsVal, err := assertType(rhs, "string")
		if err != nil {
			return false
		}

		return lhsVal == rhsVal
	case reflect.Int:
		lhsVal := lhs.(int)
		rhsVal, err := assertType(rhs, "int")
		if err != nil {
			return false
		}

		return lhsVal == rhsVal

	default:
		fmt.Printf("Unknown type %s\n", t.Kind().String())
		return false
	}

	return false
}

func storeData(storeSpec models.StoreSpec, store *storage, resp Response) error {
	for name, spec := range storeSpec {
		valueFromResponse, err := resp.tryExtractValue(Expr(spec.Value), spec.Type)
		if err != nil {
			return err
		}
		if valueFromResponse != nil {
			store.Set(name, valueFromResponse)
			continue
		}

		store.Set(name, spec.Value)
	}

	return nil
}
