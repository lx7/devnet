package gui

import (
	"testing"

	"github.com/gotk3/gotk3/gtk"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	err := ErrInvalidGTKObj{
		have: &gtk.Window{},
		want: &gtk.Button{},
	}
	assert.Contains(t, err.Error(), "does not match expected type")
}
