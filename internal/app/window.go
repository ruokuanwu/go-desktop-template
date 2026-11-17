package app

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"go.uber.org/zap"

	"go-desktop-template/internal/pkg/gtkfix"
	"go-desktop-template/internal/view"
)

func (a *App) ready() {
	if err := gtkfix.TextColor(); err != nil {
		zap.L().Warn("apply text color workaround", zap.Error(err))
	}
	// On the UI thread, NewWindow defers the first render until the next loop turn.
	var screen *view.View
	a.window = mygo.NewWindow(mygo.WindowOptions{
		Title: "Todo Desktop", Width: 1000, Height: 720,
		MinWidth: 720, MinHeight: 520,
		StateKey: "main", Content: ui.View(func(c *ui.Context) { screen.Render(c) }),
	})
	screen = view.NewView(a.todos, a.settings)
	a.window.OnClosed(func() {
		a.todos.Close()
		a.settings.Close()
	})
	var err error
	a.tray, err = newTray(trayOptions{Show: a.show, Quit: mygo.App.Quit})
	if err != nil {
		zap.L().Warn("create system tray", zap.Error(err))
	} else {
		mygo.App.OnWillQuit(func(*mygo.QuitEvent) { a.tray.Destroy() })
	}
	a.window.OnClose(a.onClose)
}

func (a *App) show() {
	a.window.Show()
	if a.window.IsMinimized() {
		a.window.Restore()
	}
	a.window.Focus()
}

func (a *App) onClose(e *mygo.CloseEvent) {
	if !a.quitting && a.tray != nil && a.settings.State().Value.CloseToTray {
		e.PreventDefault()
		a.window.Hide()
	}
}

func (a *App) post(fn func()) { a.window.Update(fn) }
