package pages

import (
	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/action"
	"github.com/lobehub/lobe-mygo/pkg/theme"

	"go-desktop-template/internal/store"
)

type Config struct {
	config  *store.Config
	dataDir string
}

func NewConfig(config *store.Config, dataDir string) Config {
	return Config{config: config, dataDir: dataDir}
}

func (v *Config) Render(c *ui.Context) {
	state := v.config.State()
	tk := theme.TokensOf(c)
	ui.Text(c, "外观").FontSize(18).Bold()
	description := "主题设置自动保存在本机。"
	if state.Busy {
		description = "正在保存主题设置…"
	}
	ui.Row(c).FillWidth().Height(32).Gap(12).Children(func() {
		ui.Text(c, description).TextColor(tk.TextSecondary)
		ui.Text(c, state.Message).Grow(1).MinWidth(0).SingleLine().TextColor(tk.Error)
	})
	ui.Row(c).Gap(10).Children(func() {
		for _, option := range []struct{ label, value string }{{"跟随系统", "system"}, {"浅色", "light"}, {"深色", "dark"}} {
			variant := action.Default
			if state.Value.Theme == option.value {
				variant = action.Primary
			}
			c.Key(option.value)
			action.Button(c, option.label, action.ButtonProps{Variant: variant}).Disabled(state.Busy).OnClick(func() { v.config.SetTheme(option.value) })
		}
	})
	ui.Text(c, "本地数据目录").FontSize(18).Bold()
	ui.Text(c, v.dataDir).Wrap().TextColor(tk.TextSecondary)
	ui.Text(c, "任务保存在 todos.db；关闭程序后可复制此文件备份。").Wrap().TextColor(tk.TextSecondary)
	ui.Text(c, "纯 Go 原生界面，无 WebView，无本地 HTTP 服务。").Wrap().TextColor(tk.TextSecondary)
}
