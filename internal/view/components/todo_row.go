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
	ui.Row(c).FillWidth().Padding(14).Gap(12).Radius(10).Background(tk.BgContainer).Children(func() {
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
		c.Key("delete")
		action.Button(c, "删除", action.ButtonProps{Variant: action.Text}).Disabled(busy).OnClick(func() { remove(todo.ID) })
	})
}
