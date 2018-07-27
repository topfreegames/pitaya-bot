package bot

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
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
		return nil, fmt.Errorf("token '%s' not found", token)
	}

	return value, nil
}

func sliceAccess(term string) (int, string) {
	r := regexp.MustCompile(`\[([0-9]+)\]`)
	ssubmatch := r.FindStringSubmatch(term)

	if len(ssubmatch) == 2 {
		idx, _ := strconv.Atoi(ssubmatch[1])
		return idx, term[0:strings.Index(term, "[")]
	}

	return -1, ""
}

func extractValue(src map[string]interface{}, expr Expr, exprType string) (interface{}, error) {
	tokens := expr.tokenize()
	var container interface{} = src
	for i, token := range tokens {
		// Is literal
		if isLiteral(container) {
			if i == len(tokens)-1 {
				break // Found value. Exit loop
			}

			return nil, fmt.Errorf("malformed spec file. expr %s doesn't match the object received", expr)
		}

		// Is slice
		idx, exprWithoutBracket := sliceAccess(string(token))
		if idx != -1 {
			container = (container.(map[string]interface{})[exprWithoutBracket]).([]interface{})[idx]
			continue
		}

		// Is object
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

	finalValue, err := assertType(container, exprType)
	if err != nil {
		return nil, err
	}

	return finalValue, nil
}

func (r Response) tryExtractValue(expr Expr, exprType string) (interface{}, error) {
	return r.extractValue(expr, exprType)
}

func (r Response) extractValue(expr Expr, exprType string) (interface{}, error) {
	return extractValue(r, expr, exprType)
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
