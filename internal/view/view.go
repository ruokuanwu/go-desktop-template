package view

import (
	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/theme"

	"go-desktop-template/internal/store"
	"go-desktop-template/internal/view/layouts"
	"go-desktop-template/internal/view/pages"
)

type View struct {
	preferences *store.Preferences
	page        string
	tasks       pages.Todo
	dashboard   pages.Dashboard
	settings    pages.Settings
}

func New(todos *store.Todo, preferences *store.Preferences, dataDir string) *View {
	v := &View{
		preferences: preferences,
		page:        "任务",
		tasks:       pages.NewTodo(todos),
		settings:    pages.NewSettings(preferences, dataDir),
	}
	v.dashboard = pages.NewDashboard(todos, func() { v.page = "任务" })
	return v
}

func (v *View) Render(c *ui.Context) {
	preferences := v.preferences.State()
	switch preferences.Value.Appearance {
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
			v.settings.Render(c)
		default:
			v.tasks.Render(c)
		}
	})
}
