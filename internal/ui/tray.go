package ui

import (
	"github.com/egoist/mygo"

	"go-desktop-template/internal/pkg/i18n"
	"go-desktop-template/resources"
)

type trayOptions struct {
	Show func()
	Quit func()
}

func newTray(opts trayOptions, tr *i18n.Translator) (*mygo.Tray, *mygo.Menu, error) {
	menu := mygo.NewMenu([]*mygo.MenuItem{
		{ID: "show", Label: tr.T("tray.show"), Click: func(*mygo.MenuItem, *mygo.Window) { opts.Show() }},
		{ID: "quit", Label: tr.T("tray.quit"), Click: func(*mygo.MenuItem, *mygo.Window) { opts.Quit() }},
	})
	tray, err := mygo.NewTray(mygo.TrayOptions{
		ID: "com.example.tododesktop.main", Icon: resources.Icon,
		ToolTip: "Todo Desktop", Menu: menu,
	})
	return tray, menu, err
}

func (d *Desktop) updateTrayLabels() {
	d.trayMenu.ItemByID("show").SetLabel(d.translator.T("tray.show"))
	d.trayMenu.ItemByID("quit").SetLabel(d.translator.T("tray.quit"))
}
