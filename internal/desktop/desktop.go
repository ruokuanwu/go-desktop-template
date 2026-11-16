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
	todos    *store.TodoStore
	settings *store.SettingsStore
	quitting bool
}

// Options provides application-owned stores; Run does not close them.
type Options struct {
	Todos    *store.TodoStore
	Settings *store.SettingsStore
}

func Run(opts Options) error {
	d := &desktop{todos: opts.Todos, settings: opts.Settings}
	mygo.App.WhenReady(d.ready)
	return mygo.App.Run()
}

func (d *desktop) ready() {
	if err := gtkfix.TextColor(); err != nil {
		zap.L().Warn("apply text color workaround", zap.Error(err))
	}
	var screen *view.View
	d.window = mygo.NewWindow(mygo.WindowOptions{
		Title: "Todo Desktop", Width: 1000, Height: 720,
		MinWidth: 720, MinHeight: 520,
		StateKey: "main", Content: ui.View(func(c *ui.Context) { screen.Render(c) }),
	})
	unsubscribeTodos := d.todos.Subscribe(func(store.TodoState) { d.window.Invalidate() })
	unsubscribeSettings := d.settings.Subscribe(func(store.SettingsState) { d.window.Invalidate() })
	// On the UI thread, NewWindow defers the first render until the next loop turn.
	screen = view.New(d.todos, d.settings, d.window.Update)
	d.window.OnClosed(func() {
		unsubscribeTodos()
		unsubscribeSettings()
	})
	mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) { d.quitting = true })
	var err error
	d.tray, err = newTray(trayOptions{Show: d.show, Quit: mygo.App.Quit})
	if err != nil {
		zap.L().Warn("create system tray", zap.Error(err))
		d.settings.SetTrayError(err)
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
	if !d.quitting && d.tray != nil && d.settings.State().Value.CloseToTray {
		e.PreventDefault()
		d.window.Hide()
	}
}
