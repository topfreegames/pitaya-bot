package bot

import (
	"fmt"
	"strconv"
	"strings"
)

// Expr is the expression which contains the tokens to access the object value
type Expr string

// Response is the response received from the pitaya server beign tested
type Response interface{}

func (e Expr) tokenize() []string {
	if strings.HasPrefix(string(e), "$response") {
		return tokenSplit(string(e)[9:])
	}
	return []string{string(e)}
}

func tokenSplit(str string) []string {
	isQuote := false
	f := func(c rune) bool {
		if c == '"' {
			isQuote = !isQuote
			return true
		}
		if !isQuote {
			return c == '[' || c == ']' || c == '.' || c == ' '
		}
		return false
	}

	result := strings.FieldsFunc(str, f)
	return result
}

func tryExtractValue(resp *Response, expr Expr, exprType string) (interface{}, error) {
	tokens := expr.tokenize()
	var container interface{} = (interface{})(*resp)
	var ok bool
	for i, token := range tokens {
		switch container.(type) {
		case []interface{}:
			if index, err := strconv.Atoi(token); err == nil {
				if len(container.([]interface{})) <= index {
					return nil, fmt.Errorf("token index %v not available within expr %s", index, expr)
				}
				container = container.([]interface{})[index]
			} else {
				return nil, fmt.Errorf("malformed spec file. expr %s, '%s' token must be an index", expr, token)
			}
		case map[string]interface{}:
			if container, ok = container.(map[string]interface{})[token]; !ok {
				return nil, fmt.Errorf("token '%s' not found within expr %s", token, expr)
			}
		default:
			if i == len(tokens)-1 {
				break // Found value. Exit loop
			}

			return nil, fmt.Errorf("malformed spec file. expr %s doesn't match the object received", expr)
		}
	}

	finalValue, err := assertType(container, exprType)
	if err != nil {
		return nil, err
	}

	return finalValue, nil
}
