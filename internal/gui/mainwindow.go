package gui

import (
	"github.com/gotk3/gotk3/gtk"
)

type mainWindow struct {
	*gtk.ApplicationWindow
	controlCheckbox *gtk.CheckButton
	shareButton     *gtk.ToggleButton
	shareControls   *gtk.Box
}

func (w *mainWindow) Populate(b *gtk.Builder) error {
	obj, err := b.GetObject("main_window")
	if err != nil {
		return err
	}

	win, ok := obj.(*gtk.ApplicationWindow)
	if !ok {
		return ErrInvalidGTKObj{have: obj, want: w.ApplicationWindow}
	}
	w.ApplicationWindow = win
	w.SetKeepAbove(true)

	obj, err = b.GetObject("control_checkbox")
	if err != nil {
		return err
	}
	chk, ok := obj.(*gtk.CheckButton)
	if !ok {
		return ErrInvalidGTKObj{have: obj, want: w.controlCheckbox}
	}
	w.controlCheckbox = chk

	obj, err = b.GetObject("share_controls")
	if err != nil {
		return err
	}
	sharectl, ok := obj.(*gtk.Box)
	if !ok {
		return ErrInvalidGTKObj{have: obj, want: w.shareControls}
	}
	w.shareControls = sharectl

	obj, err = b.GetObject("share_button")
	if err != nil {
		return err
	}
	btn, ok := obj.(*gtk.ToggleButton)
	if !ok {
		return ErrInvalidGTKObj{have: obj, want: w.shareButton}
	}
	w.shareButton = btn

	return nil
}
