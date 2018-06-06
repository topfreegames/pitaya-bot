package bot

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/topfreegames/pitaya/client"
)

func initializeDb(store *storage) error {
	return nil
}

func castType(value interface{}, typ string, store *storage) (interface{}, error) {
	ret := value

	if val, ok := value.(string); ok && val[0] == '$' {
		variable := val[1:]
		if val, ok := store.Get(variable); ok {
			ret = val
		} else {
			return nil, fmt.Errorf("Variable %s not found", variable)
		}
	} else {
		ret = value
	}

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
		value, err := castType(p["value"], p["type"].(string), store)
		if err != nil {
			return nil, err
		}
		preparedArgs[key] = value
	}

	return preparedArgs, nil
}

func sendRequest(args map[string]interface{}, route string, pclient *client.Client) (interface{}, error) {
	encodedData, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	err = pclient.SendRequest(route, encodedData)
	if err != nil {
		return nil, err
	}

	// TODO: Define type
	var ret interface{}
	select {
	case val := <-pclient.IncomingMsgChan:
		// TODO: Treat Route?
		if val.Err {
			return nil, fmt.Errorf("Request error: %s", string(val.Data))
		}
		err = json.Unmarshal(val.Data, ret)
		if err != nil {
			err = fmt.Errorf("Error unmarshaling response: %s", err)
			return nil, err
		}
	case <-time.After(time.Second):
		return nil, fmt.Errorf("Timeout waiting for response on route %s", route)
	}

	return ret, nil
}

func validateExpectations(expectations map[string]interface{}, resp interface{}) error {
	return nil
}

func storeData(storeMap map[string]interface{}, store *storage, resp interface{}) error {
	respMap := resp.(map[string]interface{})
	for name, params := range storeMap {
		// TODO: Add support to nested variables
		p := params.(map[string]interface{})
		key := p["value"].(string)
		// typ := p["type"] // TODO: Use type to assert
		if val, ok := respMap[key]; ok {
			store.Set(name, val)
		} else {
			return fmt.Errorf("Key %s not found in response", key)
		}
	}
	return nil
}
