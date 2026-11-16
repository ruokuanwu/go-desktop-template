package pages

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/action"
	"github.com/lobehub/lobe-mygo/pkg/theme"

	"go-desktop-template/internal/model"
	"go-desktop-template/internal/store"
)

func counts(items []model.Todo) (total, completed int) {
	for _, todo := range items {
		if todo.Completed {
			completed++
		}
	}
	return len(items), completed
}

type DashboardPage struct {
	todos  *store.TodoStore
	manage func()
}

func NewDashboardPage(todos *store.TodoStore, manage func()) DashboardPage {
	return DashboardPage{todos: todos, manage: manage}
}

func (v *DashboardPage) Render(c *ui.Context) {
	state := v.todos.State()
	total, completed := counts(state.Items)
	tk := theme.TokensOf(c)
	ui.Column(c).Gap(8).Children(func() {
		ui.Text(c, "工作概览").FontSize(24).FontWeight(600)
		ui.Text(c, "专注当下，让每一件事有条不紊。").FontSize(13).TextColor(tk.TextSecondary)
	})
	ui.Row(c).FillWidth().Gap(12).Children(func() {
		for _, stat := range []struct {
			label string
			count int
		}{{"全部任务", total}, {"待完成", total - completed}, {"已完成", completed}} {
			ui.Column(c).Grow(1).MinWidth(0).Padding(20).Radius(12).Background(tk.BgContainer).
				Border(1, tk.BorderSecondary).Gap(16).Children(func() {
				ui.Text(c, stat.label).FontSize(12).TextColor(tk.TextSecondary)
				ui.Text(c, fmt.Sprint(stat.count)).FontSize(32).FontWeight(600)
			})
		}
	})
	ui.Column(c).FillWidth().Padding(24).Radius(12).Border(1, tk.BorderSecondary).Gap(12).Children(func() {
		ui.Text(c, "每一件小事，都值得完成。").FontSize(18).FontWeight(500)
		ui.Text(c, "整理待办、跟进进度，留出空间专注重要的事。").FontSize(13).Wrap().TextColor(tk.TextSecondary)
		action.Button(c, "查看我的任务", action.ButtonProps{Variant: action.Primary}).OnClick(v.manage)
	})
}
