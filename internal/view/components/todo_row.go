package components

import (
	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/action"
	"github.com/lobehub/lobe-mygo/pkg/choice"
	"github.com/lobehub/lobe-mygo/pkg/theme"

	"go-desktop-template/internal/model"
	"go-desktop-template/internal/pkg/i18n"
)

func TodoRow(c *ui.Context, tr *i18n.Translator, todo model.Todo, busy bool, setCompleted func(int64, bool), remove func(int64)) {
	tk := theme.TokensOf(c)
	row := ui.Row(c).FillWidth().MinHeight(52).Padding(10, 16).Gap(12).
		Border(1, tk.BorderSecondary).BorderWidth(0, 0, 1, 0)
	if row.Hovered() {
		row.Background(tk.BgContainer)
	}
	row.Children(func() {
		checked := todo.Completed
		c.Key("completed")
		choice.Checkbox(c, &checked, "", choice.CheckboxProps{}).Label(tr.T("todos.toggle")).Disabled(busy).OnClick(func() {
			setCompleted(todo.ID, !todo.Completed)
		})
		color := tk.Text
		if todo.Completed {
			color = tk.TextSecondary
		}
		ui.Text(c, todo.Title).Grow(1).MinWidth(0).Wrap().TextColor(color)
		status := tr.T("status.pending")
		if todo.Completed {
			status = tr.T("status.completed")
		}
		ui.Text(c, status).FontSize(11).TextColor(tk.TextSecondary)
		c.Key("delete")
		action.Button(c, tr.T("common.delete"), action.ButtonProps{Variant: action.Text, Size: theme.Small}).Disabled(busy).OnClick(func() { remove(todo.ID) })
	})
}
