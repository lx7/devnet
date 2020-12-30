package gui

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/rs/zerolog/log"
)

type builder struct {
	*gtk.Builder
}

// builderNewFromString returns a new instance of gui.builder. It extends
// gtk.Builder and provides specific getters for GTK objects to reduce clutter
// in the gui logic.
//
// The builder.get[GTKObj]() methods fail with a fatal error if the requested
// object does not exist in the layout. Use gtk.Builder if this behaviour is
// not desired.
func builderNewFromString(s string) (*builder, error) {
	b, err := gtk.BuilderNewFromString(s)
	if err != nil {
		return nil, err
	}
	return &builder{Builder: b}, nil
}

func (b *builder) getObject(id string) glib.IObject {
	obj, err := b.GetObject(id)
	if err != nil {
		log.Fatal().Err(err).Str("object_id", id).Msg("failed to get gtk obj")
	}
	return obj
}

func (b *builder) getWindow(id string) *gtk.Window {
	r, ok := b.getObject(id).(*gtk.Window)
	if !ok {
		log.Fatal().Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *builder) getApplicationWindow(id string) *gtk.ApplicationWindow {
	r, ok := b.getObject(id).(*gtk.ApplicationWindow)
	if !ok {
		log.Fatal().Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *builder) getBox(id string) *gtk.Box {
	r, ok := b.getObject(id).(*gtk.Box)
	if !ok {
		log.Fatal().Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *builder) getDrawingArea(id string) *gtk.DrawingArea {
	r, ok := b.getObject(id).(*gtk.DrawingArea)
	if !ok {
		log.Fatal().Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *builder) getButton(id string) *gtk.Button {
	r, ok := b.getObject(id).(*gtk.Button)
	if !ok {
		log.Fatal().Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *builder) getToggleButton(id string) *gtk.ToggleButton {
	r, ok := b.getObject(id).(*gtk.ToggleButton)
	if !ok {
		log.Fatal().Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *builder) getCheckButton(id string) *gtk.CheckButton {
	r, ok := b.getObject(id).(*gtk.CheckButton)
	if !ok {
		log.Fatal().Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}
