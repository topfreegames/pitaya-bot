package bot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

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

func mapAccess(term string) (string, string) {
	r := regexp.MustCompile(`\[(.*?)\]`)
	ssubmatch := r.FindStringSubmatch(term)

	if len(ssubmatch) == 2 {
		key := ssubmatch[1][1 : len(ssubmatch[1])-1]
		return key, term[0:strings.Index(term, "[")]
	}

	return "", ""
}

func tryExtractValue(resp interface{}, expr Expr, exprType string) (interface{}, error) {
	tokens := expr.tokenize()
	var container interface{} = resp
	fmt.Printf("container's type: %T\n", container)
	fmt.Printf("tokens: %v\n", tokens)
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

		// Is map
		key, exprWithoutBracket := mapAccess(string(token))
		if key != "" {
			fmt.Printf("key: %v\n", key)
			container = (container.(map[string]interface{})[exprWithoutBracket]).(map[string]interface{})[key]
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

func isLiteral(i interface{}) bool {
	fmt.Printf("container's type: %T\n", i)
	switch i.(type) {
	case []interface{}, map[string]interface{}:
		return false
	default:
		return true
	}
}
