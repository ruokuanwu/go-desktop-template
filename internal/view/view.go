package view

import (
	"github.com/egoist/mygo/ui"

	"go-desktop-template/internal/store"
	"go-desktop-template/internal/view/layouts"
	"go-desktop-template/internal/view/pages"
)

type View struct {
	settings     *store.SettingsStore
	page         string
	tasks        pages.TodoPage
	dashboard    pages.DashboardPage
	settingsPage pages.SettingsPage
}

func New(todos *store.TodoStore, settings *store.SettingsStore, update func(func())) *View {
	v := &View{
		settings:     settings,
		page:         "任务",
		tasks:        pages.NewTodoPage(todos, update),
		settingsPage: pages.NewSettingsPage(settings),
	}
	v.dashboard = pages.NewDashboardPage(todos, func() { v.page = "任务" })
	return v
}

func (v *View) Render(c *ui.Context) {
	settings := v.settings.State()
	switch settings.Value.Theme {
	case "light":
		c.SetTheme(linearTheme.Theme(false))
	case "dark":
		c.SetTheme(linearTheme.Theme(true))
	default:
		linearTheme.Apply(c)
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
