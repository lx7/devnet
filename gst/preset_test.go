package gst

import (
	"strings"
	"testing"
	"time"

	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/internal/testutil"
	"github.com/pion/webrtc/v2/pkg/media"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreset_PresetBySource(t *testing.T) {
	p, err := PresetBySource(SourceTypeScreen, H264, AccelTypeNone)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	p, err = PresetBySource(SourceTypeScreen, H264, codecAccel(""))
	assert.Error(t, err)
	assert.Nil(t, p)

	p, err = PresetBySource(SourceTypeScreen, codecName(""), AccelTypeNone)
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestPreset_Permutations(t *testing.T) {
	hook := testutil.NewLogHook()
	gtk.Init(nil)

	type perm struct{ p1, p2 Preset }
	var perms []perm

	// set up permutations of all screen presets
	ps := PresetsBySource(SourceTypeScreen)
	for _, p1 := range ps {
		// encode into rtp packets (normally done by the webrtc layer)
		p1.Local = strings.ReplaceAll(
			p1.Local, "! appsink", "! rtph264pay ! appsink")
		for _, p2 := range ps {
			perms = append(perms, perm{p1, p2})
		}
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		for _, p := range perms {
			local, err := NewPipeline(p.p1.Local, p.p1.Clock)
			require.NoError(t, err)

			remote, err := NewPipeline(p.p2.Remote, p.p2.Clock)
			require.NoError(t, err)

			log.Infof("%s -> %s", p.p1.String(), p.p2.String())

			local.HandleSample(func(s media.Sample) {
				remote.Push(s.Data)
			})
			local.Start()
			remote.Start()

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

	errorlog := hook.Entry(log.ErrorLevel)
	if errorlog != nil {
		t.Errorf("runtime error: '%v'", errorlog.Message)
	}
}
