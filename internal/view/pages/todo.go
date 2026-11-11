package pages

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/action"
	"github.com/lobehub/lobe-mygo/pkg/field"
	"github.com/lobehub/lobe-mygo/pkg/theme"

	"go-desktop-template/internal/store"
	"go-desktop-template/internal/view/components"
)

type Todo struct {
	todos                 *store.Todo
	title, search, filter string
}

func NewTodo(todos *store.Todo) Todo {
	return Todo{todos: todos, filter: "全部"}
}

func (v *Todo) Render(c *ui.Context) {
	state := v.todos.State()
	add := func() { v.todos.Add(v.title, func() { v.title = "" }) }
	total, completed := counts(state.Items)
	tk := theme.TokensOf(c)
	ui.Row(c).FillWidth().Height(32).Gap(12).Children(func() {
		ui.Text(c, fmt.Sprintf("共 %d 项 · 待完成 %d 项 · 已完成 %d 项", total, total-completed, completed)).TextColor(tk.TextSecondary)
		ui.Text(c, state.Message).Grow(1).MinWidth(0).SingleLine().TextColor(tk.Error)
	})
	addLabel := "添加"
	if state.Busy {
		addLabel = "保存中…"
	}
	ui.Row(c).FillWidth().Gap(12).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			c.Key("new-todo")
			field.Input(c, &v.title, field.InputProps{Placeholder: "添加任务，按 Enter 保存（1–200 字）"}).Label("任务标题").Disabled(state.Busy).OnSubmit(add)
		})
		c.Key("add-todo")
		action.Button(c, addLabel, action.ButtonProps{Variant: action.Primary}).Width(96).Disabled(state.Busy).OnClick(add)
	})
	c.Key("search")
	field.Input(c, &v.search, field.InputProps{Placeholder: "搜索任务标题"})
	ui.Row(c).Gap(8).Children(func() {
		for _, filter := range []string{"全部", "待完成", "已完成"} {
			variant := action.Text
			if v.filter == filter {
				variant = action.Filled
			}
			c.Key(filter)
			action.Button(c, filter, action.ButtonProps{Variant: variant}).OnClick(func() { v.filter = filter })
		}
	})
	ui.Scroll(c).Grow(1).FillWidth().Children(func() {
		ui.Column(c).FillWidth().Gap(10).Children(func() {
			visible := 0
			for _, todo := range state.Items {
				if v.filter == "待完成" && todo.Completed || v.filter == "已完成" && !todo.Completed {
					continue
				}
				if !strings.Contains(strings.ToLower(todo.Title), strings.ToLower(strings.TrimSpace(v.search))) {
					continue
				}
				visible++
				c.Key(todo.ID)
				components.TodoRow(c, todo, state.Busy, v.todos.SetCompleted, v.todos.Delete)
			}
			if visible == 0 {
				ui.Text(c, "暂无任务，添加一件想完成的事吧。").Padding(24).TextColor(tk.TextSecondary)
			}
		})
	})
}
