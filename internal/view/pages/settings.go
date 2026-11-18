package pages

import (
	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/action"
	"github.com/lobehub/lobe-mygo/pkg/choice"
	"github.com/lobehub/lobe-mygo/pkg/theme"

	"go-desktop-template/internal/pkg/i18n"
	"go-desktop-template/internal/store"
)

type SettingsPage struct {
	settings *store.SettingsStore
}

func NewSettingsPage(settings *store.SettingsStore) SettingsPage {
	return SettingsPage{settings: settings}
}

func (v *SettingsPage) Render(c *ui.Context, tr *i18n.Translator) {
	state := v.settings.State()
	tk := theme.TokensOf(c)
	ui.Column(c).Gap(8).Children(func() {
		ui.Text(c, tr.T("settings.title")).FontSize(24).FontWeight(600)
		ui.Text(c, tr.T("settings.subtitle")).FontSize(13).TextColor(tk.TextSecondary)
	})
	ui.Scroll(c).Grow(1).MinHeight(0).FillWidth().Children(func() {
		ui.Column(c).FillWidth().Gap(20).Children(func() {
			ui.Column(c).FillWidth().Padding(20).Gap(12).Radius(12).Background(tk.BgContainer).Border(1, tk.BorderSecondary).Children(func() {
				ui.Text(c, tr.T("settings.appearance")).FontSize(14).FontWeight(600)
				description := tr.T("settings.autoSave")
				if state.Busy {
					description = tr.T("settings.saving")
				}
				ui.Row(c).FillWidth().Height(32).Gap(12).Children(func() {
					ui.Text(c, description).TextColor(tk.TextSecondary)
					message := ""
					if state.Message != "" {
						message = tr.T("error.settings")
					}
					ui.Text(c, message).Grow(1).MinWidth(0).SingleLine().TextColor(tk.Error)
				})
				ui.Row(c).Gap(10).Children(func() {
					for _, option := range []struct{ label, value string }{{tr.T("common.system"), "system"}, {tr.T("theme.light"), "light"}, {tr.T("theme.dark"), "dark"}} {
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
				ui.Text(c, tr.T("settings.language")).FontSize(14).FontWeight(600)
				ui.Text(c, tr.T("settings.languageHint")).Wrap().TextColor(tk.TextSecondary)
				ui.Row(c).Gap(10).Children(func() {
					for _, option := range []struct{ label, value string }{{tr.T("common.system"), "system"}, {"简体中文", "zh-CN"}, {"English", "en"}} {
						variant := action.Default
						if state.Value.Language == option.value || state.Value.Language == "" && option.value == "system" {
							variant = action.Filled
						}
						action.Button(c.Key("language-"+option.value), option.label, action.ButtonProps{Variant: variant}).Disabled(state.Busy).OnClick(func() { v.settings.SetLanguage(option.value) })
					}
				})
			})
			ui.Column(c).FillWidth().Padding(20).Gap(12).Radius(12).Background(tk.BgContainer).Border(1, tk.BorderSecondary).Children(func() {
				ui.Text(c, tr.T("settings.window")).FontSize(14).FontWeight(600)
				ui.Row(c).Gap(12).Children(func() {
					enabled := state.Value.CloseToTray
					choice.Switch(c.Key("close-to-tray"), &enabled, choice.SwitchProps{}).
						Label(tr.T("settings.closeToTray")).Disabled(state.Busy).
						OnClick(func() { v.settings.SetCloseToTray(enabled) })
					ui.Text(c, tr.T("settings.closeToTray"))
				})
				ui.Text(c, tr.T("settings.trayHint")).Wrap().TextColor(tk.TextSecondary)
			})
			ui.Column(c).FillWidth().Padding(20).Gap(12).Radius(12).Background(tk.BgContainer).Border(1, tk.BorderSecondary).Children(func() {
				ui.Text(c, tr.T("settings.data")).FontSize(14).FontWeight(600)
				ui.Text(c, tr.T("settings.backup")).Wrap().TextColor(tk.TextSecondary)
				ui.Text(c, tr.T("settings.native")).Wrap().TextColor(tk.TextSecondary)
			})
		})
	})
}
