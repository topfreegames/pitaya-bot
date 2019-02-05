package bot

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/topfreegames/pitaya-bot/constants"
	"github.com/topfreegames/pitaya-bot/metrics"
	"github.com/topfreegames/pitaya-bot/models"
	"github.com/topfreegames/pitaya-bot/storage"
)

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
	allowedTypes := map[string]bool{"string": true, "bool": true, "int": true, "<nil>": true}
	if _, ok := allowedTypes[typ]; !ok {
		return nil, fmt.Errorf("Unknown type %s", typ)
	}

	switch v := value.(type) {
	case string, bool, int, nil:
		return assertCastedType(v, fmt.Sprintf("%T", v), typ)
	case float64:
		return assertCastedType(int(v), "int", typ)
	default:
		return nil, fmt.Errorf("Unknown value type %T", v)
	}
}

func assertCastedType(ret interface{}, givenType, expectedType string) (interface{}, error) {
	if givenType == expectedType {
		return ret, nil
	}
	return nil, fmt.Errorf("%s type assertion failed for field: %v", expectedType, ret)
}

func parseArg(params interface{}, store storage.Storage) (interface{}, error) {
	p := params.(map[string]interface{})

	valueFromStorage, err := tryGetValue(p["value"], store)
	if err != nil {
		return nil, err
	}

	pVal, ok := p["type"]
	if !ok {
		return nil, fmt.Errorf("type is not available in arg")
	}
	paramType, ok := pVal.(string)
	if !ok {
		return nil, fmt.Errorf("type is not a string")
	}

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
	switch arg := value.(type) {
	case map[string]interface{}:
		return parseObject(arg, valueType, store)
	case []interface{}:
		return parseArray(arg, valueType, store)
	default:
		return assertType(value, valueType)
	}
}

func parseObject(arg map[string]interface{}, argType string, store storage.Storage) (interface{}, error) {
	if argType != "object" {
		return nil, constants.ErrMalformedObject
	}

	preparedArgs := make(map[string]interface{}, len(arg))
	for key, params := range arg {
		builtParam, err := parseArg(params, store)
		if err != nil {
			return nil, err
		}
		preparedArgs[key] = builtParam
	}

	return preparedArgs, nil
}

func parseArray(arg []interface{}, argType string, store storage.Storage) (interface{}, error) {
	if argType != "array" {
		return nil, constants.ErrMalformedObject
	}

	preparedArgs := make([]interface{}, len(arg))
	for key, params := range arg {
		builtParam, err := parseArg(params, store)
		if err != nil {
			return nil, err
		}
		preparedArgs[key] = builtParam
	}

	return preparedArgs, nil
}

func sendRequest(args interface{}, route string, pclient *PClient, metricsReporter []metrics.Reporter, logger logrus.FieldLogger) (Response, []byte, error) {
	encodedData, err := json.Marshal(args)
	if err != nil {
		return nil, nil, err
	}

	startTime := time.Now()
	response, b, err := pclient.Request(route, encodedData)
	if err != nil {
		metricsReporterTags := map[string]string{"route": route}
		for _, mr := range metricsReporter {
			reportErr := mr.ReportCount(constants.ErrorCount, metricsReporterTags, 1)
			if reportErr != nil {
				logger.WithError(reportErr).Error("Failed to Report Count")
			}
		}
	}

	elapsed := time.Since(startTime)

	metricsReporterTags := map[string]string{"route": route}
	for _, mr := range metricsReporter {
		reportErr := mr.ReportSummary(constants.ResponseTime, metricsReporterTags, float64(elapsed.Nanoseconds()/1e6))
		if reportErr != nil {
			logger.WithError(reportErr).Error("Failed to Report Summary")
		}
	}

	return response, b, err
}

func sendNotify(args interface{}, route string, pclient *PClient) error {
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

func validateExpectations(expectations models.ExpectSpec, response Response, store storage.Storage) error {
	for propertyExpr, spec := range expectations {
		expectedValue, err := getValueFromSpec(spec, store)
		if err != nil {
			return err
		}

		gotValue, err := tryExtractValue(&response, Expr(propertyExpr), spec.Type)
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

func storeData(storeSpec models.StoreSpec, store storage.Storage, response Response) error {
	for name, spec := range storeSpec {
		valueFromResponse, err := tryExtractValue(&response, Expr(spec.Value), spec.Type)
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
