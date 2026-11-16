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

type TodoPage struct {
	todos                 *store.TodoStore
	update                func(func())
	title, search, filter string
}

func NewTodoPage(todos *store.TodoStore, update func(func())) TodoPage {
	return TodoPage{todos: todos, update: update, filter: "全部"}
}

func (v *TodoPage) Render(c *ui.Context) {
	state := v.todos.State()
	add := func() {
		v.todos.Add(v.title, func() { v.update(func() { v.title = "" }) })
	}
	total, completed := counts(state.Items)
	tk := theme.TokensOf(c)
	ui.Column(c).FillWidth().Gap(8).Children(func() {
		ui.Row(c).FillWidth().Gap(10).Children(func() {
			ui.Text(c, "我的任务").FontSize(24).FontWeight(600)
			ui.Text(c, fmt.Sprint(total)).FontSize(12).TextColor(tk.TextSecondary).Padding(3, 8).Radius(6).Background(tk.FillTertiary)
		})
		ui.Text(c, fmt.Sprintf("待完成 %d 项  /  已完成 %d 项", total-completed, completed)).FontSize(12).TextColor(tk.TextSecondary)
	})
	if state.Message != "" {
		ui.Text(c, state.Message).FillWidth().Wrap().FontSize(12).TextColor(tk.Error)
	}
	addLabel := "+ 新建任务"
	if state.Busy {
		addLabel = "保存中…"
	}
	ui.Row(c).FillWidth().Gap(12).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			c.Key("new-todo")
			field.Input(c, &v.title, field.InputProps{Placeholder: "接下来要做什么？按 Enter 保存"}).Label("任务标题（1–200 字）").Disabled(state.Busy).OnSubmit(add)
		})
		c.Key("add-todo")
		action.Button(c, addLabel, action.ButtonProps{Variant: action.Primary}).Width(112).Disabled(state.Busy).OnClick(add)
	})
	ui.Row(c).FillWidth().Gap(8).Children(func() {
		for _, filter := range []string{"全部", "待完成", "已完成"} {
			variant := action.Text
			if v.filter == filter {
				variant = action.Filled
			}
			c.Key(filter)
			action.Button(c, filter, action.ButtonProps{Variant: variant, Size: theme.Small}).OnClick(func() { v.filter = filter })
		}
		ui.Column(c).Grow(1)
		ui.Column(c).Width(200).Children(func() {
			c.Key("search")
			field.Input(c, &v.search, field.InputProps{Placeholder: "搜索任务…"}).Label("搜索任务标题")
		})
	})
	ui.Scroll(c).Grow(1).MinHeight(0).FillWidth().Border(1, tk.BorderSecondary).Radius(8).Children(func() {
		ui.Column(c).FillWidth().Children(func() {
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
				ui.Column(c).FillWidth().Padding(48, 24).Gap(10).AlignItems(ui.Center).Children(func() {
					title, hint := "还没有任务", "从一件小事开始，在上方创建你的第一个任务。"
					if total > 0 || strings.TrimSpace(v.search) != "" || v.filter != "全部" {
						title, hint = "没有匹配的任务", "试试其他关键词，或切换任务筛选。"
					}
					ui.Text(c, title).FontSize(16).FontWeight(500)
					ui.Text(c, hint).FontSize(13).Wrap().TextColor(tk.TextSecondary)
				})
			}
		})
	})
}
