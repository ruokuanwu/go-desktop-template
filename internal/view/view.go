package view

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"go-desktop-template/internal/pkg/i18n"
	"go-desktop-template/internal/store"
	"go-desktop-template/internal/view/layouts"
	"go-desktop-template/internal/view/pages"
)

type View struct {
	translator   *i18n.Translator
	settings     *store.SettingsStore
	router       *ui.Router
	tasks        pages.TodoPage
	dashboard    pages.DashboardPage
	settingsPage pages.SettingsPage
}

func NewView(todos *store.TodoStore, settings *store.SettingsStore, translator *i18n.Translator) *View {
	v := &View{
		settings:     settings,
		translator:   translator,
		router:       ui.NewRouter("/todos"),
		tasks:        pages.NewTodoPage(todos),
		settingsPage: pages.NewSettingsPage(settings),
	}
	v.dashboard = pages.NewDashboardPage(todos, func() { v.router.Push("/todos") })
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
	page := strings.TrimPrefix(v.router.Path(), "/")
	layouts.Shell(c, v.translator, page, func(page string) { v.router.Push("/" + page) }, func() {
		v.router.View(c, func(r *ui.Route) {
			r.Page().Gap(20)
			switch {
			case r.Match("/dashboard"):
				r.Title(v.translator.T("nav.dashboard"))
				v.dashboard.Render(c, v.translator)
			case r.Match("/settings"):
				r.Title(v.translator.T("nav.settings"))
				v.settingsPage.Render(c, v.translator)
			case r.Match("/todos"):
				r.Title(v.translator.T("nav.todos"))
				v.tasks.Render(c, v.translator)
			}
		})
	})
}
