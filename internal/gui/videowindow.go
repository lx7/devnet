package gui

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type videoWindow struct {
	*gtk.Window
	overlay *gtk.GLArea
}

func (w *videoWindow) Populate(b *Builder) error {
	w.Window = b.GetWindow("video_window")
	w.SetTitle("devnet Video")

	w.overlay = b.GetGLArea("screencast_overlay")
	w.overlay.AddEvents(4)
	w.overlay.Connect("event", w.onDaEvent)

	return nil
}

func (w *videoWindow) onDaEvent(da *gtk.GLArea, ev *gdk.Event) bool {
	// evMotion := gdk.EventMotionNewFromEvent(ev)
	// x, y := evMotion.MotionVal()

	// log.Printf("x: %v y: %v", int(x), int(y))

	// labelX.SetLabel(fmt.Sprintf("%d", int(x)))
	// labelY.SetLabel(fmt.Sprintf("%d", int(y)))
	return false
}
