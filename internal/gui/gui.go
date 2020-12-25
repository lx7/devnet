package gui

//go:generate esc -o static.gen.go -pkg gui -include .*\.(ui|css)$ -prefix ../../cmd/devnet ../../cmd/devnet/static
import (
	"fmt"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/internal/client"
	log "github.com/sirupsen/logrus"
)

const (
	layoutPath = "/static/main.ui"
	stylePath  = "/static/gtk.css"
)

type GUI struct {
	app         *gtk.Application
	mainWindow  *mainWindow
	videoWindow *videoWindow
	session     client.SessionI
	ready       chan bool
}

func New(id string, s client.SessionI) (*GUI, error) {
	app, err := gtk.ApplicationNew(id, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		return nil, fmt.Errorf("failed to create gtk application: %v", err)
	}
	g := &GUI{
		app:         app,
		mainWindow:  &mainWindow{},
		videoWindow: &videoWindow{},
		session:     s,
		ready:       make(chan bool),
	}

	app.Connect("startup", g.onStartup)
	app.Connect("activate", g.onActivate)
	app.Connect("shutdown", g.onShutdown)

	return g, nil
}

func (g *GUI) Run() int {
	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			select {
			case e := <-g.session.Events():
				g.onEvent(e)
			case <-g.ready:
				g.session.SetOverlay(client.RemoteScreen, g.videoWindow.overlay)
			case <-done:
				break
			}
		}
		log.Infof("gui client event loop done")
	}()

	return g.app.Run(nil)
}

func (g *GUI) Quit() {
	execOnMain(func() { g.app.Quit() })
}

func (g *GUI) onEvent(e client.Event) {
	switch e.(type) {
	case client.EventSessionStart:
		execOnMain(func() { g.mainWindow.shareControls.Show() })
	case client.EventSessionEnd:
		execOnMain(func() { g.mainWindow.shareControls.Hide() })
	case client.EventSCInboundStart:
		execOnMain(func() { g.videoWindow.Show() })
	case client.EventSCInboundEnd:
		execOnMain(func() { g.videoWindow.Hide() })
	}
}

func (g *GUI) onStartup() {
	log.Info("application startup")

	ui, err := FSString(true, layoutPath)
	if err != nil {
		log.Fatal("failed to load layout: ", err)
	}

	builder, err := gtk.BuilderNewFromString(ui)
	if err != nil {
		log.Fatal("failed to read layout: ", err)
	}

	css, err := FSString(true, stylePath)
	if err != nil {
		log.Fatal("failed to load styles: ", err)
	}

	cssProvider, _ := gtk.CssProviderNew()
	err = cssProvider.LoadFromData(css)
	if err != nil {
		log.Fatal("failed to read styles: ", err)
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
		log.Fatal(err)
	}
	g.app.AddWindow(g.mainWindow)

	if err := g.videoWindow.Populate(builder); err != nil {
		log.Fatal(err)
	}
	g.app.AddWindow(g.videoWindow)

}

func (g *GUI) onActivate() {
	log.Info("application activated")
	g.mainWindow.Show()
	g.videoWindow.Show()
	g.videoWindow.Hide()
	g.ready <- true
}

func (g *GUI) onShutdown() {
	log.Info("application shutdown")
}

func (g *GUI) onDestroy() {
	g.app.Quit()
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
		log.Errorf("failed to run func on main: %v with args: %v", f, args)
	}
}
