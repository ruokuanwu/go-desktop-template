package desktop

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"go.uber.org/zap"

	"go-desktop-template/internal/pkg/gtkfix"
	"go-desktop-template/internal/store"
	"go-desktop-template/internal/view"
)

type desktop struct {
	window   *mygo.Window
	tray     *mygo.Tray
	config   *store.Config
	quitting bool
}

// Lifecycle connects background result publication and cancellation to the window.
// Ready runs on the UI thread before the window can receive input.
type Lifecycle struct {
	Ready  func(update func(func()))
	Closed func()
}

type Options struct {
	Todos     *store.Todo
	Config    *store.Config
	Lifecycle Lifecycle
}

func Run(opts Options) error {
	d := &desktop{config: opts.Config}
	mygo.App.WhenReady(func() { d.ready(opts) })
	return mygo.App.Run()
}

func (d *desktop) ready(opts Options) {
	if err := gtkfix.TextColor(); err != nil {
		zap.L().Warn("apply text color workaround", zap.Error(err))
	}
	screen := view.New(opts.Todos, d.config)
	d.window = mygo.NewWindow(mygo.WindowOptions{
		Title: "Todo Desktop", Width: 1000, Height: 720,
		MinWidth: 720, MinHeight: 520,
		StateKey: "main", Content: ui.View(screen.Render),
	})
	opts.Lifecycle.Ready(d.window.Update)
	d.window.OnClosed(opts.Lifecycle.Closed)
	mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) { d.quitting = true })
	var err error
	d.tray, err = newTray(trayOptions{Show: d.show, Quit: mygo.App.Quit})
	if err != nil {
		zap.L().Warn("create system tray", zap.Error(err))
		d.config.SetTrayError(err)
		d.window.Invalidate()
	} else {
		mygo.App.OnWillQuit(func(*mygo.QuitEvent) { d.tray.Destroy() })
	}
	d.window.OnClose(d.onClose)
}

func (d *desktop) show() {
	d.window.Show()
	if d.window.IsMinimized() {
		d.window.Restore()
	}
	d.window.Focus()
}

func (d *desktop) onClose(e *mygo.CloseEvent) {
	if !d.quitting && d.tray != nil && d.config.State().Value.CloseToTray {
		e.PreventDefault()
		d.window.Hide()
	}
}
