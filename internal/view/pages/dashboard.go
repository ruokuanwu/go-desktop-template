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

type Dashboard struct {
	todos  *store.Todos
	manage func()
}

func NewDashboard(todos *store.Todos, manage func()) Dashboard {
	return Dashboard{todos: todos, manage: manage}
}

func (v *Dashboard) Render(c *ui.Context) {
	state := v.todos.State()
	Status(c, state.Message, state.Busy)
	total, completed := counts(state.Items)
	tk := theme.TokensOf(c)
	ui.Text(c, "每一件小事，都值得完成。").TextColor(tk.TextSecondary)
	for _, stat := range []struct {
		label string
		count int
	}{{"全部任务", total}, {"待完成", total - completed}, {"已完成", completed}} {
		ui.Row(c).FillWidth().Padding(18).Radius(12).Background(tk.BgContainer).Gap(20).Children(func() {
			ui.Text(c, stat.label).Grow(1)
			ui.Text(c, fmt.Sprint(stat.count)).FontSize(26).Bold()
		})
	}
	action.Button(c, "管理任务", action.ButtonProps{Variant: action.Primary}).OnClick(v.manage)
}
