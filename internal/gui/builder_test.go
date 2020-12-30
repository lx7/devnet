package gui

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	runtime.LockOSThread()
}

func TestBuilder_GetObj(t *testing.T) {
	gtk.Init(nil)

	tests := []struct {
		give string
		want interface{}
		f    func(*builder, string) interface{}
	}{
		{
			give: "main_window",
			want: &gtk.ApplicationWindow{},
			f: func(b *builder, id string) interface{} {
				return b.getApplicationWindow(id)
			},
		},
		{
			give: "video_window",
			want: &gtk.Window{},
			f: func(b *builder, id string) interface{} {
				return b.getWindow(id)
			},
		},
		{
			give: "wait_screen",
			want: &gtk.Box{},
			f: func(b *builder, id string) interface{} {
				return b.getBox(id)
			},
		},
		{
			give: "screencast_overlay",
			want: &gtk.DrawingArea{},
			f: func(b *builder, id string) interface{} {
				return b.getDrawingArea(id)
			},
		},
		{
			give: "share_button",
			want: &gtk.ToggleButton{},
			f: func(b *builder, id string) interface{} {
				return b.getToggleButton(id)
			},
		},
	}

	ui, err := FSString(false, layoutPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load layout")
	}

	b, err := builderNewFromString(ui)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read layout")
	}

	go func() {
		// run tests
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%v", tt.want), func(t *testing.T) {
				have := tt.f(b, tt.give)
				assert.IsType(t, tt.want, have)
			})
		}

		// stop main loop
		glib.IdleAdd(gtk.MainQuit)
	}()

	gtk.Main()
}
