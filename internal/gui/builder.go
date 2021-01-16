package gui

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/rs/zerolog/log"
)

type Builder struct {
	*gtk.Builder
}

// BuilderNewFromString returns a new instance of gui.Builder. It extends
// gtk.Builder and provides specific getters for GTK objects to reduce clutter
// in the gui logic.
//
// The Builder.Get[GTKObj]() methods fail with a fatal error if the requested
// object does not exist in the layout. Use gtk.Builder if this behaviour is
// not desired.
func BuilderNewFromString(s string) (*Builder, error) {
	b, err := gtk.BuilderNewFromString(s)
	if err != nil {
		return nil, err
	}
	return &Builder{Builder: b}, nil
}

func (b *Builder) GetObject(id string) glib.IObject {
	obj, err := b.Builder.GetObject(id)
	if err != nil {
		log.Fatal().Err(err).Str("object_id", id).Msg("failed to get gtk obj")
	}
	return obj
}

func (b *Builder) GetWindow(id string) *gtk.Window {
	r, ok := b.GetObject(id).(*gtk.Window)
	if !ok {
		log.Fatal().Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *Builder) GetApplicationWindow(id string) *gtk.ApplicationWindow {
	r, ok := b.GetObject(id).(*gtk.ApplicationWindow)
	if !ok {
		log.Fatal().Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *Builder) GetBox(id string) *gtk.Box {
	r, ok := b.GetObject(id).(*gtk.Box)
	if !ok {
		log.Fatal().Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *Builder) GetGLArea(id string) *gtk.GLArea {
	r, ok := b.GetObject(id).(*gtk.GLArea)
	if !ok {
		log.Fatal().Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *Builder) GetDrawingArea(id string) *gtk.DrawingArea {
	r, ok := b.GetObject(id).(*gtk.DrawingArea)
	if !ok {
		log.Fatal().Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *Builder) GetButton(id string) *gtk.Button {
	r, ok := b.GetObject(id).(*gtk.Button)
	if !ok {
		log.Fatal().Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *Builder) GetToggleButton(id string) *gtk.ToggleButton {
	r, ok := b.GetObject(id).(*gtk.ToggleButton)
	if !ok {
		log.Fatal().Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *Builder) GetCheckButton(id string) *gtk.CheckButton {
	r, ok := b.GetObject(id).(*gtk.CheckButton)
	if !ok {
		log.Fatal().Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}
