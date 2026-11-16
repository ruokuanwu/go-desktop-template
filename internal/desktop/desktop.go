package desktop

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"go.uber.org/zap"

	"go-desktop-template/internal/model"
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

type Options struct {
	Todos       store.TodoService
	Items       []model.Todo
	Settings    store.SettingsService
	Preferences model.Settings
}

func Run(opts Options) error {
	d := &desktop{}
	defer func() {
		if d.todos != nil {
			d.todos.Close()
			d.todos.Wait()
		}
		if d.settings != nil {
			d.settings.Close()
			d.settings.Wait()
		}
	}()
	mygo.App.WhenReady(func() { d.ready(opts) })
	return mygo.App.Run()
}

func (d *desktop) ready(opts Options) {
	if err := gtkfix.TextColor(); err != nil {
		zap.L().Warn("apply text color workaround", zap.Error(err))
	}
	var screen *view.View
	d.window = mygo.NewWindow(mygo.WindowOptions{
		Title: "Todo Desktop", Width: 1000, Height: 720,
		MinWidth: 720, MinHeight: 520,
		StateKey: "main", Content: ui.View(func(c *ui.Context) { screen.Render(c) }),
	})
	d.todos = store.NewTodoStore(opts.Todos, d.window.Update, opts.Items)
	d.settings = store.NewSettingsStore(opts.Settings, opts.Preferences)
	unsubscribeSettings := d.settings.Subscribe(d.window.Invalidate)
	// On the UI thread, NewWindow defers the first render until the next loop turn.
	screen = view.New(d.todos, d.settings)
	d.window.OnClosed(func() {
		unsubscribeSettings()
		d.todos.Close()
		d.settings.Close()
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
