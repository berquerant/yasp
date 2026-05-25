package internal

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

type IDGenerator struct {
	e       *vm.Program
	rawTmpl string
}

var (
	ErrInvalidIDGenerator  = errors.New("InvalidIDGenerator")
	ErrIDGenerationFailure = errors.New("IDGenerationFailure")
)

// Create a new IDGenerator from expr.
func NewIDGenerator(s string) (*IDGenerator, error) {
	e, err := expr.Compile(s, expr.AsKind(reflect.String))
	if err != nil {
		return nil, errors.Join(ErrInvalidIDGenerator, err)
	}
	return &IDGenerator{
		e:       e,
		rawTmpl: s,
	}, nil
}

func (g *IDGenerator) GenerateID(m map[string]any) (string, error) {
	x, err := expr.Run(g.e, m)
	if err != nil {
		return "", fmt.Errorf("%w: m=%#v", errors.Join(ErrIDGenerationFailure, err), m)
	}
	if s, ok := x.(string); ok && s != "" {
		return s, nil
	}
	return "", fmt.Errorf("%w: template=%v", ErrIDGenerationFailure, g.rawTmpl)
}
