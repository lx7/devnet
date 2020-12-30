package gui

import (
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
		desc string
		give string
		want *gtk.Box
	}{
		{
			desc: "existing object",
			give: "wait_screen",
			want: &gtk.Box{},
		},
	}

	ui, err := FSString(false, layoutPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load layout")
	}

	b, err := BuilderNewFromString(ui)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read layout")
	}

	go func() {
		// run tests
		for _, tt := range tests {
			t.Run(tt.desc, func(t *testing.T) {
				have := b.getBox(tt.give)
				assert.IsType(t, tt.want, have)
			})
		}

		// stop main loop
		glib.IdleAdd(gtk.MainQuit)
	}()

	gtk.Main()
}
