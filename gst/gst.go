package gst

/*
#cgo pkg-config: gtk+-3.0 gstreamer-1.0 gstreamer-app-1.0 gstreamer-video-1.0
#include "gst.h"
#include <gtk/gtk.h>
*/
import "C"
import (
	"errors"
	"sync"
	"unsafe"

	"github.com/gotk3/gotk3/gtk"
	"github.com/pion/webrtc/v2/pkg/media"
	log "github.com/sirupsen/logrus"
)

type SampleHandlerFunc func(media.Sample)

type Pipeline struct {
	id      int
	el      *C.GstElement
	handler SampleHandlerFunc
	clock   float32
}

func NewPipeline(desc string, clock float32) *Pipeline {
	descU := C.CString(desc)
	defer C.free(unsafe.Pointer(descU))

	p := &Pipeline{clock: clock}
	p.id = pipes.register(p)
	p.el = C.gs_new_pipeline(descU, C.int(p.id))

	return p
}

func (p *Pipeline) SetOverlayHandle(w gtk.IWidget) error {
	if w == nil {
		return errors.New("invalid overlay handle: nil")
	}
	widget := w.ToWidget()
	gdkWin, err := widget.GetParentWindow()
	if err != nil {
		return err
	}

	native := C.to_gdk_window(C.ulong(gdkWin.Native()))
	C.gs_pipeline_set_overlay_handle(p.el, native)
	return nil
}

func (p *Pipeline) HandleSample(f SampleHandlerFunc) {
	p.handler = f
}

func (p *Pipeline) Start() {
	C.gs_pipeline_start(p.el)
}

func (p *Pipeline) Stop() {
	C.gs_pipeline_stop(p.el)
}

func (p *Pipeline) Push(buf []byte) {
	bytes := C.CBytes(buf)
	defer C.free(bytes)
	C.gs_pipeline_appsrc_push(p.el, bytes, C.int(len(buf)))
}

func (p *Pipeline) Destroy() {
	C.gs_pipeline_destroy(p.el)
	pipes.unregister(p.id)
}

//export go_sample_cb
func go_sample_cb(ref C.int, buf unsafe.Pointer, bufl C.int, dur C.int) {
	p := pipes.lookup(int(ref))
	if p == nil {
		log.Errorf("no pipeline with id %v, discarding buffer", int(ref))
		return
	}

	p.handler(media.Sample{
		Data:    C.GoBytes(buf, bufl),
		Samples: uint32(p.clock * float32(dur) / 1000000000),
	})
	C.free(buf)
}

type pipeRegister struct {
	sync.Mutex
	m    map[int]*Pipeline
	last int
}

var pipes = pipeRegister{m: make(map[int]*Pipeline)}

func (r pipeRegister) register(p *Pipeline) int {
	r.Lock()
	defer r.Unlock()
	r.last++
	for r.m[r.last] != nil {
		r.last++
	}
	r.m[r.last] = p
	return r.last
}

func (r pipeRegister) lookup(i int) *Pipeline {
	r.Lock()
	defer r.Unlock()
	return r.m[i]
}

func (r pipeRegister) unregister(i int) {
	r.Lock()
	defer r.Unlock()
	delete(r.m, i)
}
