package components

import (
	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/action"
	"github.com/lobehub/lobe-mygo/pkg/choice"
	"github.com/lobehub/lobe-mygo/pkg/theme"

	"go-desktop-template/internal/model"
)

func TodoRow(c *ui.Context, todo model.Todo, busy bool, setCompleted func(int64, bool), remove func(int64)) {
	tk := theme.TokensOf(c)
	row := ui.Row(c).FillWidth().MinHeight(52).Padding(10, 16).Gap(12).
		Border(1, tk.BorderSecondary).BorderWidth(0, 0, 1, 0)
	if row.Hovered() {
		row.Background(tk.BgContainer)
	}
	row.Children(func() {
		checked := todo.Completed
		c.Key("completed")
		choice.Checkbox(c, &checked, "", choice.CheckboxProps{}).Label("切换完成状态").Disabled(busy).OnClick(func() {
			setCompleted(todo.ID, !todo.Completed)
		})
		color := tk.Text
		if todo.Completed {
			color = tk.TextSecondary
		}
		ui.Text(c, todo.Title).Grow(1).MinWidth(0).Wrap().TextColor(color)
		status := "待完成"
		if todo.Completed {
			status = "已完成"
		}
		ui.Text(c, status).FontSize(11).TextColor(tk.TextSecondary)
		c.Key("delete")
		action.Button(c, "删除", action.ButtonProps{Variant: action.Text, Size: theme.Small}).Disabled(busy).OnClick(func() { remove(todo.ID) })
	})
}
