package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/topfreegames/pitaya-bot/metrics"
	"github.com/topfreegames/pitaya-bot/models"
	"github.com/topfreegames/pitaya-bot/storage"
)

func initializeDb(store storage.Storage) error {
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

func tryGetValue(expr interface{}, store storage.Storage) (interface{}, error) {
	if val, ok := expr.(string); ok {
		if strings.HasPrefix(val, "$store") {
			variable := val[7:]
			return store.Get(variable)
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
			return nil, fmt.Errorf("String type assertion failed for field: %v", ret)
		}
	case "bool":
		if val, ok := ret.(bool); ok {
			ret = val
		} else {
			return nil, fmt.Errorf("Boolean type assertion failed for field: %v", ret)
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

func parseArg(params interface{}, store storage.Storage) (interface{}, error) {
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

	return builtParam, nil
}

func buildArgByType(value interface{}, valueType string, store storage.Storage) (interface{}, error) {
	var err error
	switch valueType {
	case "object":
		arg, ok := value.(map[string]interface{})
		if !ok {
			return nil, errors.New("Malformed object type argument")
		}

		preparedArgs := map[string]interface{}{}
		for key, params := range arg {
			builtParam, err := parseArg(params, store)
			if err != nil {
				return nil, err
			}
			preparedArgs[key] = builtParam
		}

		return preparedArgs, nil
	case "array":
		arg, ok := value.([]interface{})
		if !ok {
			return nil, errors.New("Malformed object type argument")
		}

		preparedArgs := make([]interface{}, len(arg))
		for idx, params := range arg {
			builtParam, err := parseArg(params, store)
			if err != nil {
				return nil, err
			}
			preparedArgs[idx] = builtParam
		}

		return preparedArgs, nil
	default:
		value, err = assertType(value, valueType)
		if err != nil {
			return nil, err
		}
	}

	return value, nil
}

func buildArgs(rawArgs map[string]interface{}, store storage.Storage) (map[string]interface{}, error) {
	args, err := buildArgByType(rawArgs, "object", store)
	if err != nil {
		return nil, err
	}

	r := args.(map[string]interface{})
	return r, nil
}

func sendRequest(args map[string]interface{}, route string, pclient *PClient, metricsReporter []metrics.Reporter) (interface{}, []byte, error) {
	encodedData, err := json.Marshal(args)
	if err != nil {
		return nil, nil, err
	}

	startTime := time.Now()
	response, b, err := pclient.Request(route, encodedData)
	if err != nil {
		metricsReporterTags := map[string]string{"route": route}
		for _, mr := range metricsReporter {
			mr.ReportCount(metrics.ErrorCount, metricsReporterTags, 1)
		}
	}

	elapsed := time.Since(startTime)

	metricsReporterTags := map[string]string{"route": route}
	for _, mr := range metricsReporter {
		mr.ReportSummary(metrics.ResponseTime, metricsReporterTags, float64(elapsed.Nanoseconds()/1e6))
	}

	return response, b, err
}

func sendNotify(args map[string]interface{}, route string, pclient *PClient) error {
	encodedData, err := json.Marshal(args)
	if err != nil {
		return err
	}

	return pclient.Notify(route, encodedData)
}

func getValueFromSpec(spec models.ExpectSpecEntry, store storage.Storage) (interface{}, error) {
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

func validateExpectations(expectations models.ExpectSpec, response interface{}, store storage.Storage) error {
	for propertyExpr, spec := range expectations {
		expectedValue, err := getValueFromSpec(spec, store)
		if err != nil {
			return err
		}

		gotValue, err := tryExtractValue(response, Expr(propertyExpr), spec.Type)
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

	case reflect.Bool:
		lhsVal := lhs.(bool)
		rhsVal, err := assertType(rhs, "bool")
		if err != nil {
			return false
		}

		return lhsVal == rhsVal

	default:
		fmt.Printf("Unknown type %s\n", t.Kind().String())
		return false
	}
}

func storeData(storeSpec models.StoreSpec, store storage.Storage, response interface{}) error {
	for name, spec := range storeSpec {
		valueFromResponse, err := tryExtractValue(response, Expr(spec.Value), spec.Type)
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
