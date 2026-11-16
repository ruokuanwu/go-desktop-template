package pages

import (
	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/action"
	"github.com/lobehub/lobe-mygo/pkg/choice"
	"github.com/lobehub/lobe-mygo/pkg/theme"

	"go-desktop-template/internal/store"
)

type SettingsPage struct {
	settings *store.SettingsStore
}

func NewSettingsPage(settings *store.SettingsStore) SettingsPage {
	return SettingsPage{settings: settings}
}

func (v *SettingsPage) Render(c *ui.Context) {
	state := v.settings.State()
	tk := theme.TokensOf(c)
	ui.Column(c).Gap(8).Children(func() {
		ui.Text(c, "偏好设置").FontSize(24).FontWeight(600)
		ui.Text(c, "按你的习惯，安排工作空间。").FontSize(13).TextColor(tk.TextSecondary)
	})
	ui.Scroll(c).Grow(1).MinHeight(0).FillWidth().Children(func() {
		ui.Column(c).FillWidth().Gap(20).Children(func() {
			ui.Column(c).FillWidth().Padding(20).Gap(12).Radius(12).Background(tk.BgContainer).Border(1, tk.BorderSecondary).Children(func() {
				ui.Text(c, "外观").FontSize(14).FontWeight(600)
				description := "设置自动保存在本机。"
				if state.Busy {
					description = "正在保存设置…"
				}
				ui.Row(c).FillWidth().Height(32).Gap(12).Children(func() {
					ui.Text(c, description).TextColor(tk.TextSecondary)
					ui.Text(c, state.Message).Grow(1).MinWidth(0).SingleLine().TextColor(tk.Error)
				})
				ui.Row(c).Gap(10).Children(func() {
					for _, option := range []struct{ label, value string }{{"跟随系统", "system"}, {"浅色", "light"}, {"深色", "dark"}} {
						variant := action.Default
						if state.Value.Theme == option.value {
							variant = action.Filled
						}
						c.Key(option.value)
						action.Button(c, option.label, action.ButtonProps{Variant: variant}).Disabled(state.Busy).OnClick(func() { v.settings.SetTheme(option.value) })
					}
				})
			})
			ui.Column(c).FillWidth().Padding(20).Gap(12).Radius(12).Background(tk.BgContainer).Border(1, tk.BorderSecondary).Children(func() {
				ui.Text(c, "窗口行为").FontSize(14).FontWeight(600)
				ui.Row(c).Gap(12).Children(func() {
					enabled := state.Value.CloseToTray
					choice.Switch(c.Key("close-to-tray"), &enabled, choice.SwitchProps{}).
						Label("关闭到系统托盘").Disabled(state.Busy).
						OnClick(func() { v.settings.SetCloseToTray(enabled) })
					ui.Text(c, "关闭到系统托盘")
				})
				ui.Text(c, "关闭窗口后继续在后台运行，可从托盘显示或退出。").Wrap().TextColor(tk.TextSecondary)
			})
			ui.Column(c).FillWidth().Padding(20).Gap(12).Radius(12).Background(tk.BgContainer).Border(1, tk.BorderSecondary).Children(func() {
				ui.Text(c, "本地数据").FontSize(14).FontWeight(600)
				ui.Text(c, "任务保存在 data.db；关闭程序后可复制此文件备份。").Wrap().TextColor(tk.TextSecondary)
				ui.Text(c, "纯 Go 原生界面，无 WebView，无本地 HTTP 服务。").Wrap().TextColor(tk.TextSecondary)
			})
		})
	})
}
