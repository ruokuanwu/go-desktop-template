package view

import (
	"github.com/egoist/mygo/ui"

	"go-desktop-template/internal/pkg/i18n"
	"go-desktop-template/internal/store"
	"go-desktop-template/internal/view/layouts"
	"go-desktop-template/internal/view/pages"
)

type View struct {
	translator   *i18n.Translator
	settings     *store.SettingsStore
	page         string
	tasks        pages.TodoPage
	dashboard    pages.DashboardPage
	settingsPage pages.SettingsPage
}

func NewView(todos *store.TodoStore, settings *store.SettingsStore, translator *i18n.Translator) *View {
	v := &View{
		settings:     settings,
		translator:   translator,
		page:         "todos",
		tasks:        pages.NewTodoPage(todos),
		settingsPage: pages.NewSettingsPage(settings),
	}
	v.dashboard = pages.NewDashboardPage(todos, func() { v.page = "todos" })
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
	layouts.Shell(c, v.translator, v.page, func(page string) { v.page = page }, func() {
		switch v.page {
		case "dashboard":
			v.dashboard.Render(c, v.translator)
		case "settings":
			v.settingsPage.Render(c, v.translator)
		default:
			v.tasks.Render(c, v.translator)
		}
	})
}
