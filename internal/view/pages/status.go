package pages

import (
	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/theme"
)

func Status(c *ui.Context, message string, busy bool) {
	tk := theme.TokensOf(c)
	if message != "" {
		ui.Text(c, message).TextColor(tk.Error).Wrap()
	}
	if busy {
		ui.Text(c, "正在保存…").TextColor(tk.TextSecondary)
	}
}
