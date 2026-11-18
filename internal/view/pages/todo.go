package pages

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/action"
	"github.com/lobehub/lobe-mygo/pkg/field"
	"github.com/lobehub/lobe-mygo/pkg/theme"

	"go-desktop-template/internal/pkg/i18n"
	"go-desktop-template/internal/store"
	"go-desktop-template/internal/view/components"
)

type TodoPage struct {
	todos                 *store.TodoStore
	title, search, filter string
}

func NewTodoPage(todos *store.TodoStore) TodoPage {
	return TodoPage{todos: todos, filter: "all"}
}

func (v *TodoPage) Render(c *ui.Context, tr *i18n.Translator) {
	state := v.todos.State()
	add := func() {
		v.todos.Add(v.title, func() { v.title = "" })
	}
	total, completed := counts(state.Items)
	tk := theme.TokensOf(c)
	ui.Column(c).FillWidth().Gap(8).Children(func() {
		ui.Row(c).FillWidth().Gap(10).Children(func() {
			ui.Text(c, tr.T("todos.title")).FontSize(24).FontWeight(600)
			ui.Text(c, tr.Tasks(total)).FontSize(12).TextColor(tk.TextSecondary).Padding(3, 8).Radius(6).Background(tk.FillTertiary)
		})
		ui.Text(c, tr.T("todos.summary", map[string]any{"Pending": total - completed, "Completed": completed})).FontSize(12).TextColor(tk.TextSecondary)
	})
	if state.Message != "" {
		ui.Text(c, tr.T(state.MessageID)).FillWidth().Wrap().FontSize(12).TextColor(tk.Error)
	}
	addLabel := tr.T("todos.add")
	if state.Busy {
		addLabel = tr.T("common.saving")
	}
	ui.Row(c).FillWidth().Gap(12).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			c.Key("new-todo")
			field.Input(c, &v.title, field.InputProps{Placeholder: tr.T("todos.placeholder")}).Label(tr.T("todos.titleLabel")).Disabled(state.Busy).OnSubmit(add)
		})
		c.Key("add-todo")
		action.Button(c, addLabel, action.ButtonProps{Variant: action.Primary}).MinWidth(112).Disabled(state.Busy).OnClick(add)
	})
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		for _, filter := range []string{"all", "pending", "completed"} {
			variant := action.Text
			if v.filter == filter {
				variant = action.Filled
			}
			c.Key(filter)
			action.Button(c, tr.T(map[string]string{"all": "filter.all", "pending": "status.pending", "completed": "status.completed"}[filter]), action.ButtonProps{Variant: variant, Size: theme.Small}).OnClick(func() { v.filter = filter })
		}
		ui.Column(c).Grow(1)
		ui.Column(c).Width(200).Children(func() {
			c.Key("search")
			field.Input(c, &v.search, field.InputProps{Placeholder: tr.T("todos.search")}).Label(tr.T("todos.searchLabel"))
		})
	})
	ui.Scroll(c).Grow(1).MinHeight(0).FillWidth().Border(1, tk.BorderSecondary).Radius(8).Children(func() {
		ui.Column(c).FillWidth().Children(func() {
			visible := 0
			for _, todo := range state.Items {
				if v.filter == "pending" && todo.Completed || v.filter == "completed" && !todo.Completed {
					continue
				}
				if !strings.Contains(strings.ToLower(todo.Title), strings.ToLower(strings.TrimSpace(v.search))) {
					continue
				}
				visible++
				c.Key(todo.ID)
				components.TodoRow(c, tr, todo, state.Busy, v.todos.SetCompleted, v.todos.Delete)
			}
			if visible == 0 {
				ui.Column(c).FillWidth().Padding(48, 24).Gap(10).AlignItems(ui.Center).Children(func() {
					title, hint := tr.T("todos.empty"), tr.T("todos.emptyHint")
					if total > 0 || strings.TrimSpace(v.search) != "" || v.filter != "all" {
						title, hint = tr.T("todos.noMatch"), tr.T("todos.noMatchHint")
					}
					ui.Text(c, title).FontSize(16).FontWeight(500)
					ui.Text(c, hint).FontSize(13).Wrap().TextColor(tk.TextSecondary)
				})
			}
		})
	})
}
