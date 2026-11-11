package view

import (
	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/theme"

	"go-desktop-template/internal/store"
	"go-desktop-template/internal/view/layouts"
	"go-desktop-template/internal/view/pages"
)

type View struct {
	config     *store.Config
	page       string
	tasks      pages.Todo
	dashboard  pages.Dashboard
	configPage pages.Config
}

func New(todos *store.Todo, config *store.Config, dataDir string) *View {
	v := &View{
		config:     config,
		page:       "任务",
		tasks:      pages.NewTodo(todos),
		configPage: pages.NewConfig(config, dataDir),
	}
	v.dashboard = pages.NewDashboard(todos, func() { v.page = "任务" })
	return v
}

func (v *View) Render(c *ui.Context) {
	config := v.config.State()
	switch config.Value.Theme {
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
			v.configPage.Render(c)
		default:
			v.tasks.Render(c)
		}
	})
}
