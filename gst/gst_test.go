package gst

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/lx7/devnet/internal/testutil"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/pion/webrtc/v2/pkg/media"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	runtime.LockOSThread()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	os.Setenv("GST_DEBUG", "*:2")
}

func TestGStreamer_Processing(t *testing.T) {
	hook := &testutil.LogHook{}
	log.Logger = log.Hook(hook)

	gtk.Init(nil)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	assert.NoError(t, err)

	da, err := gtk.DrawingAreaNew()
	assert.NoError(t, err)

	win.Add(da)
	win.SetTitle("Test")
	win.Connect("destroy", gtk.MainQuit)
	win.ShowAll()

	tests := []struct {
		desc string
		give *Pipeline
		run  func(*testing.T)
	}{
		{
			desc: "simple pipeline",
			run: func(t *testing.T) {
				p, err := NewPipeline(`
					videotestsrc 
					! videoconvert 
					! autovideosink
					`, 90000)
				require.NoError(t, err)
				p.Start()
				time.Sleep(1 * time.Second)
				p.Stop()
				time.Sleep(1 * time.Second)
				p.Start()
				time.Sleep(1 * time.Second)
				p.Destroy()
			},
		},
		{
			desc: "gtk overlay",
			run: func(t *testing.T) {
				p, err := NewPipeline(`
					videotestsrc 
					! videoconvert 
					! autovideosink
					`, 90000)
				require.NoError(t, err)
				err = p.SetOverlayHandle(da)
				assert.NoError(t, err)
				p.Start()
				time.Sleep(1 * time.Second)
				p.Destroy()
			},
		},
		{
			desc: "nil overlay",
			run: func(t *testing.T) {
				p, err := NewPipeline(`
					videotestsrc 
					! videoconvert 
					! autovideosink
					`, 90000)
				require.NoError(t, err)
				err = p.SetOverlayHandle(nil)
				assert.Error(t, err)
				p.Destroy()
			},
		},
		{
			desc: "appsrc push",
			run: func(t *testing.T) {
				caps := `
					audio/x-raw, 
					format=(string)S16LE,
					channels=(int)1, 
					rate=(int)44100, 
					layout=(string)interleaved
					`
				src, err := NewPipeline(`
					audiotestsrc 
					! queue 
					! appsink name=sink caps="`+caps+`"
					`, 44100)
				require.NoError(t, err)

				sink, err := NewPipeline(`
					appsrc name=src caps="`+caps+`" is-live=true format=3 
					! queue 
					! autoaudiosink
					`, 44100)
				require.NoError(t, err)

				src.HandleSample(func(s media.Sample) {
					sink.Push(s.Data)
				})
				sink.Start()
				src.Start()

				time.Sleep(1 * time.Second)

				src.Destroy()
				sink.Destroy()
			},
		},
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		for _, tt := range tests {
			t.Run(tt.desc, tt.run)
		}
		glib.IdleAdd(gtk.MainQuit)
	}()

	gtk.Main()

	entry := hook.Entry(zerolog.ErrorLevel)
	if entry != nil {
		log.Fatal().Str("err", entry.Msg).Msg("runtime error")
	}
}

func TestGStreamer_NewHWCodec(t *testing.T) {
	gtk.Init(nil)

	// define cases
	tests := []struct {
		give string
		want hwCodec
	}{
		{
			give: "",
			want: NoHardware,
		},
		{
			give: "vaapi",
			want: VAAPI,
		},
		{
			give: "nvcodec",
			want: NVCODEC,
		},
		{
			give: "vdpau",
			want: VDPAU,
		},
		{
			give: "osxvt",
			want: OSXVT,
		},
		{
			give: "OSXvt",
			want: OSXVT,
		},
	}

	// run tests
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			have := NewHardwareCodec(tt.give)
			assert.Equal(t, tt.want, have)
		})
	}
}
