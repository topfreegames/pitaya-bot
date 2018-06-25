package bot

import (
	"fmt"
	"reflect"
	"strings"
)

// Response ...
type Response map[string]interface{}

// Expr ...
type Expr string

func (e Expr) tokenize() []string {
	tokens := strings.Split(string(e), ".")

	if tokens[0] == "$response" {
		return tokens[1:]
	}
	return tokens
}

// TODO - handle slice token
func visitToken(container map[string]interface{}, token string) (interface{}, error) {
	value, ok := container[token]
	if !ok {
		return nil, fmt.Errorf("atom '%s' not found", token)
	}

	return value, nil
}

func extractValue(src map[string]interface{}, expr Expr) (interface{}, error) {
	tokens := expr.tokenize()
	var container interface{} = src
	for i, token := range tokens {
		if isLiteral(container) {
			if i == len(tokens)-1 {
				return container, nil
			}

			return nil, fmt.Errorf("malformed spec file. expr %s doesn't match the object received", expr)
		}

		parsedContainer, ok := container.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Unable to parse container to Response")
		}

		var err error
		container, err = visitToken(parsedContainer, token)
		if err != nil {
			return nil, err
		}
	}

	return container, nil
}

func (r Response) tryExtractValue(expr Expr) interface{} {
	v, _ := r.extractValue(expr)
	return v
}

func (r Response) extractValue(expr Expr) (interface{}, error) {
	return extractValue(r, expr)
}

func isLiteral(i interface{}) bool {
	t := reflect.TypeOf(i)

	switch t.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Ptr, reflect.Slice, reflect.Struct:
		return false

	default:
		return true
	}
}
