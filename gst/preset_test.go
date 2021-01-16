package gst

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/internal/testutil"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreset_GetPreset(t *testing.T) {
	p, err := GetPreset(Screen, H264, NoHardware)
	require.NoError(t, err)
	assert.Equal(t, H264, p.Codec)

	p, err = GetPreset(Voice, Opus, NoHardware)
	require.NoError(t, err)
	assert.Equal(t, Opus, p.Codec)

	p, err = GetPreset(Screen, H264, hwCodec("---"))
	require.Error(t, err)
	assert.Nil(t, p)

	p, err = GetPreset(Screen, codecName(""), NoHardware)
	require.Error(t, err)
	assert.Nil(t, p)
}

func TestPreset_Permutations(t *testing.T) {
	hook := &testutil.LogHook{}
	log.Logger = log.Hook(hook)
	gtk.Init(nil)

	type perm struct{ p1, p2 Preset }
	var perms []perm

	// set up permutations of all screen presets
	ps := PresetsBySource(Screen)
	for _, p1 := range ps {
		// encode into rtp packets (normally done by the webrtc layer)
		p1.Local = strings.ReplaceAll(
			p1.Local, "! appsink", "! rtph264pay ! appsink")
		for _, p2 := range ps {
			perms = append(perms, perm{p1, p2})
		}
	}

	// set up permutations of all voice presets
	ps = PresetsBySource(Voice)
	for _, p1 := range ps {
		// encode into rtp packets (normally done by the webrtc layer)
		p1.Local = strings.ReplaceAll(
			p1.Local, "! appsink", "! rtpopuspay ! appsink")
		p1.Local = strings.ReplaceAll(
			p1.Local, "autoaudiosrc", "audiotestsrc")
		for _, p2 := range ps {
			perms = append(perms, perm{p1, p2})
		}
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		for _, p := range perms {
			local, err := NewPipeline(p.p1.Local)
			require.NoError(t, err)

			remote, err := NewPipeline(p.p2.Remote)
			require.NoError(t, err)

			fmt.Printf("... test stream: %s -> %s\n", p.p1.String(), p.p2.String())

			local.HandleSample(func(s media.Sample) {
				remote.Push(s.Data)
			})
			remote.Start()
			local.Start()

			time.Sleep(1 * time.Second)

			local.Stop()
			local.Destroy()
			remote.Stop()
			remote.Destroy()

			time.Sleep(100 * time.Millisecond)
		}
		gtk.MainQuit()
	}()

	gtk.Main()

	entry := hook.Entry(zerolog.ErrorLevel)
	if entry != nil {
		log.Fatal().Str("err", entry.Msg).Msg("runtime error")
	}
}

func TestPreset_Webcam(t *testing.T) {
	hook := &testutil.LogHook{}
	log.Logger = log.Hook(hook)
	gtk.Init(nil)

	preset, err := GetPreset(Camera, H264, NoHardware)
	require.NoError(t, err)

	tests := []struct {
		desc   string
		local  string
		remote string
	}{
		{
			desc:   "default video source",
			local:  preset.Local,
			remote: preset.Remote,
		},
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		for _, tt := range tests {
			lp := strings.ReplaceAll(
				tt.local,
				"! appsink",
				"! rtph264pay ! appsink",
			)

			local, err := NewPipeline(lp)
			require.NoError(t, err)

			remote, err := NewPipeline(tt.remote)
			require.NoError(t, err)

			local.HandleSample(func(s media.Sample) {
				remote.Push(s.Data)
			})
			remote.Start()
			local.Start()

			time.Sleep(5 * time.Second)

			//local.Stop()
			//local.Destroy()
			//remote.Stop()
			//remote.Destroy()

			time.Sleep(100 * time.Millisecond)
		}
		gtk.MainQuit()
	}()

	gtk.Main()

	entry := hook.Entry(zerolog.ErrorLevel)
	if entry != nil {
		log.Fatal().Str("err", entry.Msg).Msg("runtime error")
	}
}
