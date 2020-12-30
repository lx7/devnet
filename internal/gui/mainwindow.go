package gui

import (
	"github.com/gotk3/gotk3/gtk"
)

type mainWindow struct {
	*gtk.ApplicationWindow
	controlCheckbox *gtk.CheckButton
	shareButton     *gtk.ToggleButton
	waitScreen      *gtk.Box
	channelList     *gtk.Box
	detailsBox      *gtk.Box
}

func (w *mainWindow) Populate(b *builder) error {
	w.ApplicationWindow = b.getApplicationWindow("main_window")
	w.SetKeepAbove(true)

	w.waitScreen = b.getBox("wait_screen")
	w.channelList = b.getBox("channel_list")
	w.controlCheckbox = b.getCheckButton("control_checkbox")
	w.detailsBox = b.getBox("details_box")
	w.shareButton = b.getToggleButton("share_button")

	return nil
}
