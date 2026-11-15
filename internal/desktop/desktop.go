package desktop

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"go.uber.org/zap"

	"go-desktop-template/internal/config"
	"go-desktop-template/internal/model"
	"go-desktop-template/internal/pkg/gtkfix"
	"go-desktop-template/internal/store"
	"go-desktop-template/internal/view"
)

type desktop struct {
	window   *mygo.Window
	tray     *mygo.Tray
	todos    *store.Todo
	config   *store.Config
	quitting bool
}

type Options struct {
	Todos      store.Service
	Items      []model.Todo
	Config     config.Config
	SaveConfig func(config.Config) error
}

func Run(opts Options) error {
	d := &desktop{}
	defer func() {
		if d.todos != nil {
			d.todos.Close()
			d.todos.Wait()
		}
		if d.config != nil {
			d.config.Close()
			d.config.Wait()
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
	d.todos = store.NewTodo(opts.Todos, d.window.Update, opts.Items)
	d.config = store.NewConfig(d.window.Update, opts.Config, opts.SaveConfig)
	// On the UI thread, NewWindow defers the first render until the next loop turn.
	screen = view.New(d.todos, d.config)
	d.window.OnClosed(func() {
		d.todos.Close()
		d.config.Close()
	})
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
