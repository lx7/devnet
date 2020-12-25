package gui

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/gst"
)

type videoWindow struct {
	*gtk.Window
	overlay *gtk.DrawingArea
	source  *gst.Pipeline
}

func (w *videoWindow) Populate(b *gtk.Builder) error {
	obj, err := b.GetObject("video_window")
	if err != nil {
		return err
	}
	win, ok := obj.(*gtk.Window)
	if !ok {
		return ErrInvalidGTKObj{have: obj, want: w.Window}
	}
	w.Window = win
	w.SetTitle("devnet Video")

	obj, err = b.GetObject("screencast_overlay")
	if err != nil {
		return err
	}
	overlay, ok := obj.(*gtk.DrawingArea)
	if !ok {
		return ErrInvalidGTKObj{have: obj, want: w.overlay}
	}
	w.overlay = overlay
	w.overlay.AddEvents(4)
	w.overlay.Connect("event", w.onDaEvent)

	return nil
}

func (w *videoWindow) onDaEvent(da *gtk.DrawingArea, ev *gdk.Event) bool {
	//evMotion := gdk.EventMotionNewFromEvent(ev)
	//x, y := evMotion.MotionVal()

	// labelX.SetLabel(fmt.Sprintf("%d", int(x)))
	// labelY.SetLabel(fmt.Sprintf("%d", int(y)))
	return false
}
