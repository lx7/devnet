package gst

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/gotk3/gotk3/gtk"
	"github.com/stretchr/testify/assert"
)

func init() {
	runtime.LockOSThread()
	os.Setenv("GST_DEBUG", "*:2")
}

func TestGStreamer_GTKWindow(t *testing.T) {
	gtk.Init(nil)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	assert.NoError(t, err)

	da, err := gtk.DrawingAreaNew()
	assert.NoError(t, err)

	win.Add(da)
	win.SetTitle("Test")
	win.Connect("destroy", gtk.MainQuit)
	win.ShowAll()

	p := NewPipeline("videotestsrc ! videoconvert ! autovideosink")
	p.SetOverlayHandle(da)
	p.Start()

	go func() {
		time.Sleep(5 * time.Second)
		gtk.MainQuit()
	}()

	gtk.Main()
}
