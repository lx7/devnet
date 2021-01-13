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

func (w *mainWindow) Populate(b *Builder) error {
	w.ApplicationWindow = b.GetApplicationWindow("main_window")
	w.SetKeepAbove(true)

	w.waitScreen = b.GetBox("wait_screen")
	w.channelList = b.GetBox("channel_list")
	w.controlCheckbox = b.GetCheckButton("control_checkbox")
	w.detailsBox = b.GetBox("details_box")
	w.shareButton = b.GetToggleButton("share_button")

	return nil
}
