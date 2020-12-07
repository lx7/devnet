package gst

/*
#cgo pkg-config: gtk+-3.0 gstreamer-1.0 gstreamer-app-1.0 gstreamer-video-1.0
#include "gst.h"
#include <gtk/gtk.h>
*/
import "C"
import (
	"errors"
	"unsafe"

	"github.com/gotk3/gotk3/gtk"
)

type Pipeline struct {
	gstElement *C.GstElement
}

func NewPipeline(descr string) *Pipeline {
	descrUnsafe := C.CString(descr)
	defer C.free(unsafe.Pointer(descrUnsafe))
	return &Pipeline{
		gstElement: C.gs_new_pipeline(descrUnsafe),
	}
}

func (p *Pipeline) SetOverlayHandle(w gtk.IWidget) error {
	if w == nil {
		return errors.New("overlay handle: nil")
	}
	widget := w.ToWidget()
	gdkWindow, err := widget.GetParentWindow()
	if err != nil {
		return err
	}

	nativeWindow := C.toGdkWindow(C.ulong(gdkWindow.Native()))
	C.gs_pipeline_set_overlay_handle(p.gstElement, nativeWindow)
	return nil
}

func (p *Pipeline) Start() {
	C.gs_pipeline_start(p.gstElement)
}

func (p *Pipeline) Stop() {
	C.gs_pipeline_stop(p.gstElement)
}
