package desktop

import (
	"github.com/egoist/mygo"

	"go-desktop-template/resources"
)

type trayOptions struct {
	Show func()
	Quit func()
}

func newTray(opts trayOptions) (*mygo.Tray, error) {
	menu := mygo.NewMenu([]*mygo.MenuItem{
		{Label: "显示", Click: func(*mygo.MenuItem, *mygo.Window) { opts.Show() }},
		{Label: "退出", Click: func(*mygo.MenuItem, *mygo.Window) { opts.Quit() }},
	})
	return mygo.NewTray(mygo.TrayOptions{
		ID: "com.example.tododesktop.main", Icon: resources.Icon,
		ToolTip: "Todo Desktop", Menu: menu,
	})
}
