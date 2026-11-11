package layouts

import (
	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/action"
	"github.com/lobehub/lobe-mygo/pkg/theme"
)

func Shell(c *ui.Context, page string, navigate func(string), content func()) {
	tk := theme.TokensOf(c)
	ui.Row(c).Fill().Background(tk.BgLayout).Children(func() {
		ui.Column(c).Width(176).FillHeight().Padding(20).Gap(12).Background(tk.BgContainer).Children(func() {
			ui.Text(c, "TODO DESKTOP").FontSize(16).Bold()
			ui.Text(c, "本地 · 原生 · 离线").FontSize(12).TextColor(tk.TextSecondary)
			for _, target := range []string{"概览", "任务", "设置"} {
				variant := action.Text
				if page == target {
					variant = action.Filled
				}
				c.Key(target)
				action.Button(c, target, action.ButtonProps{Variant: variant}).FillWidth().OnClick(func() { navigate(target) })
			}
			ui.Column(c).Grow(1)
			ui.Text(c, "MyGo + Lobe MyGo").FontSize(11).TextColor(tk.TextSecondary)
		})
		ui.Column(c).Grow(1).MinWidth(0).FillHeight().Padding(28).Gap(18).Children(func() {
			ui.Text(c, page).FontSize(28).Bold()
			content()
		})
	})
}
