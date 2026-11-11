package pages

import (
	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/action"
	"github.com/lobehub/lobe-mygo/pkg/theme"

	"go-desktop-template/internal/store"
)

type Settings struct {
	preferences *store.Preferences
	dataDir     string
}

func NewSettings(preferences *store.Preferences, dataDir string) Settings {
	return Settings{preferences: preferences, dataDir: dataDir}
}

func (v *Settings) Render(c *ui.Context) {
	state := v.preferences.State()
	Status(c, state.Message, state.Busy)
	tk := theme.TokensOf(c)
	ui.Text(c, "外观").FontSize(18).Bold()
	ui.Text(c, "主题设置自动保存在本机。").TextColor(tk.TextSecondary)
	ui.Row(c).Gap(10).Children(func() {
		for _, option := range []struct{ label, value string }{{"跟随系统", "system"}, {"浅色", "light"}, {"深色", "dark"}} {
			variant := action.Default
			if state.Value.Appearance == option.value {
				variant = action.Primary
			}
			c.Key(option.value)
			action.Button(c, option.label, action.ButtonProps{Variant: variant}).Disabled(state.Busy).OnClick(func() { v.preferences.SetAppearance(option.value) })
		}
	})
	ui.Text(c, "本地数据目录").FontSize(18).Bold()
	ui.Text(c, v.dataDir).Wrap().TextColor(tk.TextSecondary)
	ui.Text(c, "任务保存在 todos.db；关闭程序后可复制此文件备份。").Wrap().TextColor(tk.TextSecondary)
	ui.Text(c, "纯 Go 原生界面，无 WebView，无本地 HTTP 服务。").Wrap().TextColor(tk.TextSecondary)
}
