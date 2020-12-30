package gui

import (
	"github.com/gotk3/gotk3/gtk"
	"github.com/rs/zerolog/log"
)

type builder struct {
	*gtk.Builder
}

func BuilderNewFromString(s string) (*builder, error) {
	b, err := gtk.BuilderNewFromString(s)
	if err != nil {
		return nil, err
	}
	return &builder{Builder: b}, nil
}

func (b *builder) getWindow(id string) *gtk.Window {
	obj, err := b.GetObject(id)
	if err != nil {
		log.Fatal().Err(err).Str("object_id", id).Msg("failed to get gtk obj")
	}
	r, ok := obj.(*gtk.Window)
	if !ok {
		log.Fatal().Err(err).Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *builder) getApplicationWindow(id string) *gtk.ApplicationWindow {
	obj, err := b.GetObject(id)
	if err != nil {
		log.Fatal().Err(err).Str("object_id", id).Msg("failed to get gtk obj")
	}
	r, ok := obj.(*gtk.ApplicationWindow)
	if !ok {
		log.Fatal().Err(err).Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *builder) getBox(id string) *gtk.Box {
	obj, err := b.GetObject(id)
	if err != nil {
		log.Fatal().Err(err).Str("object_id", id).Msg("failed to get gtk obj")
	}
	r, ok := obj.(*gtk.Box)
	if !ok {
		log.Fatal().Err(err).Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *builder) getDrawingArea(id string) *gtk.DrawingArea {
	obj, err := b.GetObject(id)
	if err != nil {
		log.Fatal().Err(err).Str("object_id", id).Msg("failed to get gtk obj")
	}
	r, ok := obj.(*gtk.DrawingArea)
	if !ok {
		log.Fatal().Err(err).Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *builder) getButton(id string) *gtk.Button {
	obj, err := b.GetObject(id)
	if err != nil {
		log.Fatal().Err(err).Str("object_id", id).Msg("failed to get gtk obj")
	}
	r, ok := obj.(*gtk.Button)
	if !ok {
		log.Fatal().Err(err).Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *builder) getToggleButton(id string) *gtk.ToggleButton {
	obj, err := b.GetObject(id)
	if err != nil {
		log.Fatal().Err(err).Str("object_id", id).Msg("failed to get gtk obj")
	}
	r, ok := obj.(*gtk.ToggleButton)
	if !ok {
		log.Fatal().Err(err).Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}

func (b *builder) getCheckButton(id string) *gtk.CheckButton {
	obj, err := b.GetObject(id)
	if err != nil {
		log.Fatal().Err(err).Str("object_id", id).Msg("failed to get gtk obj")
	}
	r, ok := obj.(*gtk.CheckButton)
	if !ok {
		log.Fatal().Err(err).Str("object_id", id).Msg("gtk obj type mismatch")
	}
	return r
}
