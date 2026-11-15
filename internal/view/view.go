package view

import (
	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/theme"

	"go-desktop-template/internal/store"
	"go-desktop-template/internal/view/layouts"
	"go-desktop-template/internal/view/pages"
)

type View struct {
	settings     *store.Settings
	page         string
	tasks        pages.Todo
	dashboard    pages.Dashboard
	settingsPage pages.Settings
}

func New(todos *store.Todo, settings *store.Settings, dataPath string) *View {
	v := &View{
		settings:     settings,
		page:         "任务",
		tasks:        pages.NewTodo(todos),
		settingsPage: pages.NewSettings(settings, dataPath),
	}
	v.dashboard = pages.NewDashboard(todos, func() { v.page = "任务" })
	return v
}

func (v *View) Render(c *ui.Context) {
	settings := v.settings.State()
	switch settings.Value.Theme {
	case "light":
		c.SetTheme(theme.Default().Theme(false))
	case "dark":
		c.SetTheme(theme.Default().Theme(true))
	default:
		theme.Apply(c)
	}
	layouts.Shell(c, v.page, func(page string) { v.page = page }, func() {
		switch v.page {
		case "概览":
			v.dashboard.Render(c)
		case "设置":
			v.settingsPage.Render(c)
		default:
			v.tasks.Render(c)
		}
	})
}
