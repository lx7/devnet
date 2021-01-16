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
	"time"
	"unsafe"

	"github.com/gotk3/gotk3/gtk"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/rs/zerolog/log"
)

type SampleHandlerFunc func(media.Sample)

type Pipeline struct {
	id      int
	el      *C.GstElement
	handler SampleHandlerFunc
}

func NewPipeline(desc string) (*Pipeline, error) {
	descU := C.CString(desc)
	defer C.free(unsafe.Pointer(descU))

	p := &Pipeline{}
	p.id = pipes.register(p)
	p.el = C.gs_new_pipeline(descU, C.int(p.id))
	if p.el == nil {
		return nil, errors.New("failed to create pipeline")
	}

	return p, nil
}

func (p *Pipeline) SetOverlayHandle(w gtk.IWidget) error {
	if w == nil {
		return errors.New("invalid overlay handle: nil")
	}

	native := C.to_gtk_widget(C.ulong(w.ToWidget().Native()))
	C.gs_pipeline_set_overlay_handle(C.int(p.id), native)
	return nil
}

func (p *Pipeline) HandleSample(f SampleHandlerFunc) {
	p.handler = f
}

func (p *Pipeline) Push(buf []byte) {
	bytes := C.CBytes(buf)
	defer C.free(bytes)
	C.gs_pipeline_appsrc_push(p.el, bytes, C.int(len(buf)))
}

func (p *Pipeline) Start() {
	C.gs_pipeline_start(p.el)
}

func (p *Pipeline) Stop() {
	C.gs_pipeline_stop(p.el)
}

func (p *Pipeline) Destroy() {
	C.gs_pipeline_destroy(p.el)
	pipes.unregister(p.id)
}

//export go_sample_cb
func go_sample_cb(ref C.int, buf unsafe.Pointer, bufl C.int, dur C.int) {
	p := pipes.lookup(int(ref))
	if p == nil {
		log.Error().
			Int("pipeline", int(ref)).
			Msg("no pipeline with id, discarding buffer")
		return
	}

	p.handler(media.Sample{
		Data:     C.GoBytes(buf, bufl),
		Duration: time.Duration(dur),
	})
	C.free(buf)
}

//export go_error_cb
func go_error_cb(ref C.int, msg *C.char) {
	log.Error().
		Int("pipeline", int(ref)).
		Str("err", C.GoString(msg)).
		Msg("gst pipeline")
}

//export go_warning_cb
func go_warning_cb(ref C.int, msg *C.char) {
	log.Warn().
		Int("pipeline", int(ref)).
		Str("msg", C.GoString(msg)).
		Msg("gst pipeline")
}

//export go_info_cb
func go_info_cb(ref C.int, msg *C.char) {
	log.Info().
		Int("pipeline", int(ref)).
		Str("msg", C.GoString(msg)).
		Msg("gst pipeline")
}

//export go_debug_cb
func go_debug_cb(ref C.int, msg *C.char) {
	log.Debug().
		Int("pipeline", int(ref)).
		Str("msg", C.GoString(msg)).
		Msg("gst pipeline")
}

type pipeRegister struct {
	sync.Mutex
	m    map[int]*Pipeline
	last int
}

var pipes = &pipeRegister{m: make(map[int]*Pipeline)}

func (r *pipeRegister) register(p *Pipeline) int {
	r.Lock()
	defer r.Unlock()
	r.last++
	for r.m[r.last] != nil {
		r.last++
	}
	r.m[r.last] = p
	return r.last
}

func (r *pipeRegister) lookup(i int) *Pipeline {
	r.Lock()
	defer r.Unlock()
	return r.m[i]
}

func (r *pipeRegister) unregister(i int) {
	r.Lock()
	defer r.Unlock()
	delete(r.m, i)
}
