package ui

import (
	"sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"go.uber.org/zap"

	"go-desktop-template/internal/pkg/gtkfix"
	"go-desktop-template/internal/pkg/i18n"
	"go-desktop-template/internal/store"
)

// Desktop owns the native window and tray and borrows the stores for rendering.
type Desktop struct {
	todos      *store.TodoStore
	settings   *store.SettingsStore
	translator *i18n.Translator
	window     *mygo.Window
	tray       *mygo.Tray
	trayMenu   *mygo.Menu
	quitting   bool
}

func NewDesktop(todos *store.TodoStore, settings *store.SettingsStore) *Desktop {
	return &Desktop{todos: todos, settings: settings}
}

// Run starts the UI loop and notifies onExit on the UI thread when closing for real.
func (d *Desktop) Run(onExit func()) error {
	tr, err := i18n.New(d.settings.State().Value.Language)
	if err != nil {
		return err
	}
	d.translator = tr
	onExit = sync.OnceFunc(onExit)
	mygo.App.WhenReady(func() { d.ready(onExit) })
	defer mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) { d.quitting = true })()
	defer mygo.App.OnWillQuit(func(*mygo.QuitEvent) {
		onExit()
		if d.tray != nil {
			d.tray.Destroy()
		}
	})()
	return mygo.App.Run()
}

func (d *Desktop) ready(onExit func()) {
	if err := gtkfix.TextColor(); err != nil {
		zap.L().Warn("apply text color workaround", zap.Error(err))
	}
	screen := NewView(d.todos, d.settings, d.translator)
	d.window = mygo.NewWindow(mygo.WindowOptions{
		Title: "Todo Desktop", Width: 1000, Height: 720,
		MinWidth: 720, MinHeight: 520,
		StateKey: "main", Content: ui.View(screen.Render),
	})
	d.window.OnClosed(onExit)
	var err error
	d.tray, d.trayMenu, err = newTray(trayOptions{Show: d.show, Quit: mygo.App.Quit}, d.translator)
	if err != nil {
		zap.L().Warn("create system tray", zap.Error(err))
	}
	d.window.OnClose(d.onClose)
}

func (d *Desktop) show() {
	d.window.Show()
	if d.window.IsMinimized() {
		d.window.Restore()
	}
	d.window.Focus()
}

func (d *Desktop) onClose(e *mygo.CloseEvent) {
	if !d.quitting && d.tray != nil && d.settings.State().Value.CloseToTray {
		e.PreventDefault()
		d.window.Hide()
	}
}

// Post publishes store results on the UI thread after the window has opened.
func (d *Desktop) Post(fn func()) {
	d.window.Update(func() {
		previous := d.settings.State().Value.Language
		fn()
		selected := d.settings.State().Value.Language
		if previous != selected {
			d.translator.SetLanguage(selected)
			if d.tray != nil {
				d.updateTrayLabels()
			}
		}
	})
}
