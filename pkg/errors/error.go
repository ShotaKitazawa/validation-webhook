package errors

import "fmt"

type Immutable struct {
	Field string
}

func (e *Immutable) Error() string {
	return fmt.Sprintf("%s: this field is immutable", e.Field)
}
