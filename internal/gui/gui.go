package gui

//go:generate esc -o static.gen.go -pkg gui -include .*\.(ui|css)$ -prefix ../../cmd/devnet ../../cmd/devnet/static
import (
	"fmt"
	"os"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/internal/client"
	"github.com/rs/zerolog/log"
)

const (
	layoutPath = "/static/main.ui"
	stylePath  = "/static/gtk.css"
)

var localFiles = os.Getenv("DEVNET_LOCAL") != ""

// GUI represents the DevNet GTK interface.
type GUI struct {
	*gtk.Application

	mainWindow  *mainWindow
	videoWindow *videoWindow
	session     client.SessionI
	ready       chan bool
}

// New returns a new instance of GUI. It requires a uniqe GTK application id and
// interfaces with the business logic through client.Session.
func New(id string, s client.SessionI) (*GUI, error) {
	app, err := gtk.ApplicationNew(id, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		return nil, fmt.Errorf("failed to create gtk application: %v", err)
	}

	g := &GUI{
		Application: app,
		mainWindow:  &mainWindow{},
		videoWindow: &videoWindow{},
		session:     s,
		ready:       make(chan bool),
	}

	g.Connect("startup", g.onStartup)
	g.Connect("activate", g.onActivate)
	g.Connect("shutdown", g.onShutdown)

	return g, nil
}

// Run represents the gdk main loop. It executes the GUI logic and must be
// locked to the main thread (see runtime.LockOSThread()). Processing intensive
// non-GUI work should be done on a separate thread to minimize interference.
// Run() wil return when the application exits.
func (g *GUI) Run() int {
	done := make(chan struct{})
	defer close(done)

	// run network and video in a separate thread
	go func() {
		for {
			select {
			case e := <-g.session.Events():
				g.onSessionEvent(e)
			case <-g.ready:
				g.session.SetOverlay(client.RemoteScreen, g.videoWindow.overlay)
			case <-done:
				log.Info().Msg("gui client event loop done")
				return
			}
		}
	}()

	return g.Application.Run(nil)
}

// Quit immediately quits the application and lets the main loop return.
func (g *GUI) Quit() {
	execOnMain(func() { g.Application.Quit() })
}

func (g *GUI) onSessionEvent(e client.Event) {
	switch e.(type) {
	case client.EventConnected:
		execOnMain(func() {
			g.mainWindow.waitScreen.Hide()
			g.mainWindow.channelList.Show()
		})
	case client.EventDisconnected:
		execOnMain(func() {
			g.mainWindow.waitScreen.Show()
			g.mainWindow.channelList.Hide()
		})
	case client.EventSessionStart:
		execOnMain(func() { g.mainWindow.detailsBox.Show() })
	case client.EventSessionEnd:
		execOnMain(func() { g.mainWindow.detailsBox.Hide() })
	case client.EventSCInboundStart:
		execOnMain(func() { g.videoWindow.Show() })
	case client.EventSCInboundEnd:
		execOnMain(func() { g.videoWindow.Hide() })
	}
}

func (g *GUI) onStartup() {
	log.Info().Msg("application startup")

	ui, err := FSString(localFiles, layoutPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load layout")
	}

	builder, err := builderNewFromString(ui)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read layout")
	}

	css, err := FSString(localFiles, stylePath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load styles")
	}

	cssProvider, _ := gtk.CssProviderNew()
	err = cssProvider.LoadFromData(css)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read styles")
	}

	screen, _ := gdk.ScreenGetDefault()
	gtk.AddProviderForScreen(
		screen,
		cssProvider,
		gtk.STYLE_PROVIDER_PRIORITY_USER)

	signals := map[string]interface{}{
		"main_window_destroy": g.onDestroy,
		"share_button_toggle": g.onShareButtonToggle,
		"test_call_user1":     g.onCallUser1,
		"test_call_user2":     g.onCallUser2,
	}
	builder.ConnectSignals(signals)

	if err := g.mainWindow.Populate(builder); err != nil {
		log.Fatal().Err(err).Msg("populate main window")
	}
	g.AddWindow(g.mainWindow)

	if err := g.videoWindow.Populate(builder); err != nil {
		log.Fatal().Err(err).Msg("populate video window")
	}
	g.AddWindow(g.videoWindow)

}

func (g *GUI) onActivate() {
	log.Info().Msg("application activated")
	g.mainWindow.Show()
	g.videoWindow.Show()
	g.videoWindow.Hide()
	g.ready <- true
}

func (g *GUI) onShutdown() {
	log.Info().Msg("application shutdown")
}

func (g *GUI) onDestroy() {
	g.Application.Quit()
}

func (g *GUI) onCallUser1() {
	g.session.Connect("user1")
}

func (g *GUI) onCallUser2() {
	g.session.Connect("user2")
}

func (g *GUI) onShareButtonToggle(b *gtk.ToggleButton) {
	if b.GetActive() {
		g.session.StartStream(client.LocalScreen)
		g.mainWindow.controlCheckbox.SetSensitive(true)
	} else {
		//g.session.Screen().Stop()
		g.mainWindow.controlCheckbox.SetSensitive(false)
	}
}

func execOnMain(f interface{}, args ...interface{}) {
	_, err := glib.IdleAdd(f, args)
	if err != nil {
		log.Error().Interface("func", f).Msg("failed to run func on main loop")
	}
}
