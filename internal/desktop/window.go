package desktop

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"go-desktop-template/internal/view"
)

func Open(screen *view.View) {
	window := mygo.NewWindow(mygo.WindowOptions{
		Title: "Todo Desktop", Width: 1000, Height: 720,
		MinWidth: 720, MinHeight: 520,
		StateKey: "main", Content: ui.View(screen.Render),
	})
	screen.Attach(window)
	window.OnClosed(screen.Cancel)
}
