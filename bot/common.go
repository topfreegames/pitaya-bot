package bot

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/topfreegames/pitaya-bot/models"
	"github.com/topfreegames/pitaya/client"
)

func initializeDb(store *storage) error {
	return nil
}

func tryGetFromStorage(expr interface{}, store *storage) (interface{}, error) {
	if val, ok := expr.(string); ok && strings.HasPrefix(val, "$store") {
		variable := val[7:]
		if val, ok := store.Get(variable); ok {
			return val, nil
		}

		return nil, fmt.Errorf("Variable %s not found", variable)
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
			return nil, fmt.Errorf("Failed to cast to string")
		}
	case "int":
		if val, ok := ret.(int); ok {
			ret = val
		} else {
			return nil, fmt.Errorf("Failed to cast to int")
		}
	default:
		return nil, fmt.Errorf("Unknown type %s", typ)
	}

	return ret, nil
}

func buildArgs(rawArgs map[string]interface{}, store *storage) (map[string]interface{}, error) {
	preparedArgs := map[string]interface{}{}

	for key, params := range rawArgs {
		p := params.(map[string]interface{})
		valueFromStorage, err := tryGetFromStorage(p["value"], store)
		if err != nil {
			return nil, err
		}

		var value interface{}
		var valueType = p["type"].(string)
		if valueFromStorage != nil {
			value = valueFromStorage
		}

		value, err = assertType(value, valueType)
		if err != nil {
			return nil, err
		}
		preparedArgs[key] = value
	}

	return preparedArgs, nil
}

func sendRequest(args map[string]interface{}, route string, pclient *client.Client) (Response, error) {
	encodedData, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	err = pclient.SendRequest(route, encodedData)
	if err != nil {
		return nil, err
	}

	// TODO: Define type
	ret := make(Response)
	select {
	case val := <-pclient.IncomingMsgChan:
		// TODO: Treat Route?
		if val.Err {
			return nil, fmt.Errorf("Request error: %s", string(val.Data))
		}
		err = json.Unmarshal(val.Data, &ret)
		if err != nil {
			err = fmt.Errorf("Error unmarshaling response: %s", err)
			return nil, err
		}
	case <-time.After(time.Second):
		return nil, fmt.Errorf("Timeout waiting for response on route %s", route)
	}

	return ret, nil
}

func getValueFromSpec(spec models.ExpectSpecEntry, store *storage) (interface{}, error) {
	value, err := tryGetFromStorage(spec.Value, store)
	if err != nil {
		return nil, err
	}

	if value == nil {
		value, err = assertType(spec.Value, spec.Type)
	}

	return value, nil
}

func validateExpectations(expectations models.ExpectSpec, resp Response, store *storage) error {
	fmt.Println("validateExpectations")
	for propertyExpr, spec := range expectations {
		expectedValue, err := getValueFromSpec(spec, store)
		if err != nil {
			return err
		}

		gotValue, err := Response(resp).extractValue(Expr(propertyExpr))
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
		rhsVal, ok := rhs.(string)
		if !ok {
			return false
		}

		return lhsVal == rhsVal
	case reflect.Int:
		lhsVal := lhs.(int)
		rhsVal, ok := rhs.(int)
		if !ok {
			return false
		}

		return lhsVal == rhsVal

	default:
		fmt.Println("Unknown type %s", t.Kind().String())
		return false
	}

	return false
}

func storeData(storeSpec models.StoreSpec, store *storage, resp Response) error {
	for name, spec := range storeSpec {
		valueFromResponse := resp.tryExtractValue(Expr(spec.Value))
		if valueFromResponse != nil {
			store.Set(name, valueFromResponse)
			return nil
		}

		store.Set(name, spec.Value)
		return nil
	}

	return nil
}
