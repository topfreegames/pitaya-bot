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

func (e Expr) extractAtoms() []string {
	atoms := strings.Split(string(e), ".")

	if atoms[0] == "$response" {
		return atoms[1:]
	}
	return atoms
}

// TODO - handle slice atom
func (r Response) fromAtom(atom string) (interface{}, error) {
	value, ok := r[atom]
	if !ok {
		return nil, fmt.Errorf("atom '%s' not found", atom)
	}

	return value, nil
}

func extractValue(src map[string]interface{}, expr Expr) (interface{}, error) {
	atoms := expr.extractAtoms()
	var container interface{} = src
	for i, atom := range atoms {
		if isLiteral(container) {
			if i == len(atoms)-1 {
				return container, nil
			}

			return nil, fmt.Errorf("malformed spec file. expr %s doesn't match the object received", expr)
		}

		parsedContainer, ok := container.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Unable to parse container to Response")
		}

		var err error
		container, err = Response(parsedContainer).fromAtom(atom)
		if err != nil {
			return nil, err
		}
	}

	return container, nil
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
