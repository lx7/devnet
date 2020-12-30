package gui

import (
	"fmt"
	"reflect"
)

type ErrInvalidGTKObj struct {
	have, want interface{}
}

func (e ErrInvalidGTKObj) Error() string {
	return fmt.Sprintf(
		"gtk object type (%v) does not match expected type (%v)",
		reflect.TypeOf(e.have),
		reflect.TypeOf(e.want),
	)
}
