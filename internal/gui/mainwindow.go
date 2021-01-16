package gui

import (
	"github.com/gotk3/gotk3/gtk"
)

type mainWindow struct {
	*gtk.ApplicationWindow
	shareButton  *gtk.ToggleButton
	cameraButton *gtk.ToggleButton
	waitScreen   *gtk.Box
	channelList  *gtk.Box
	detailsBox   *gtk.Box
	remoteCam    *gtk.GLArea
	localCam     *gtk.GLArea
}

func (w *mainWindow) Populate(b *Builder) error {
	w.ApplicationWindow = b.GetApplicationWindow("main_window")
	w.SetKeepAbove(true)

	w.waitScreen = b.GetBox("wait_screen")
	w.channelList = b.GetBox("channel_list")
	w.detailsBox = b.GetBox("details_box")
	w.shareButton = b.GetToggleButton("share_button")
	w.cameraButton = b.GetToggleButton("camera_button")
	w.remoteCam = b.GetGLArea("remote_camera_overlay")
	w.localCam = b.GetGLArea("local_camera_overlay")

	return nil
}
