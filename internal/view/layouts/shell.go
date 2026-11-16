package layouts

import (
	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/action"
	"github.com/lobehub/lobe-mygo/pkg/theme"
)

func Shell(c *ui.Context, page string, navigate func(string), content func()) {
	tk := theme.TokensOf(c)
	ui.Row(c).Fill().FontSize(14).TextColor(tk.Text).Background(tk.BgLayout).Children(func() {
		ui.Column(c).Width(208).FillHeight().Padding(16, 12).Gap(6).Background(tk.BgContainer).
			Border(1, tk.BorderSecondary).BorderWidth(0, 1, 0, 0).Children(func() {
			ui.Row(c).FillWidth().Padding(8).Gap(10).Children(func() {
				ui.Text(c, "T").FontSize(14).Bold().TextColor(tk.PrimaryForeground).Background(tk.Primary).Padding(5, 9).Radius(6)
				ui.Column(c).Gap(2).Children(func() {
					ui.Text(c, "Todo Desktop").FontSize(14).FontWeight(600)
					ui.Text(c, "个人工作空间").FontSize(11).TextColor(tk.TextSecondary)
				})
			})
			ui.Text(c, "工作空间").FontSize(11).TextColor(tk.TextSecondary).Padding(16, 10, 6)
			for _, target := range []string{"概览", "任务", "设置"} {
				variant := action.Text
				if page == target {
					variant = action.Filled
				}
				c.Key(target)
				action.Button(c, target, action.ButtonProps{Variant: variant}).FillWidth().Height(34).Justify(ui.Start).PaddingX(12).OnClick(func() { navigate(target) })
			}
			ui.Column(c).Grow(1)
			ui.Column(c).FillWidth().Padding(10).Gap(5).Border(1, tk.BorderSecondary).BorderWidth(1, 0, 0, 0).Children(func() {
				ui.Text(c, "本地工作空间").FontSize(12).TextColor(tk.TextSecondary)
				ui.Text(c, "离线可用 · 数据仅保存在本机").FontSize(11).TextColor(tk.TextSecondary)
			})
		})
		ui.Column(c).Grow(1).MinWidth(0).FillHeight().Children(func() {
			ui.Row(c).FillWidth().Height(56).PaddingX(24).Gap(10).
				Border(1, tk.BorderSecondary).BorderWidth(0, 0, 1, 0).Children(func() {
				ui.Text(c, "工作空间").FontSize(13).TextColor(tk.TextSecondary)
				ui.Text(c, "/").TextColor(tk.TextSecondary)
				ui.Text(c, page).FontSize(13).FontWeight(500).Grow(1)
				ui.Text(c, "个人").FontSize(11).TextColor(tk.TextSecondary)
			})
			ui.Column(c).Grow(1).MinHeight(0).FillWidth().Padding(24).Gap(20).Children(func() {
				content()
			})
		})
	})
}
