package view

import (
	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/theme"
)

var linearTheme = newLinearTheme()

func newLinearTheme() *theme.Config {
	light, dark := theme.Light(), theme.Dark()
	for _, tk := range []*theme.Tokens{light, dark} {
		tk.Primary = ui.Hex("#5e6ad2")
		tk.PrimaryHover = ui.Hex("#828fff")
		tk.PrimaryActive = ui.Hex("#5e69d1")
		tk.PrimaryForeground = ui.Hex("#ffffff")
		tk.PrimaryText, tk.PrimaryTextHover, tk.PrimaryTextActive = tk.Primary, tk.PrimaryHover, tk.PrimaryActive
		tk.Link, tk.LinkHover, tk.LinkActive = tk.Primary, tk.PrimaryHover, tk.PrimaryActive
		tk.PrimaryBorder, tk.PrimaryBorderHover = tk.Primary, tk.PrimaryHover
		tk.BorderRadiusSM, tk.BorderRadius, tk.BorderRadiusLG = 6, 8, 12
	}
	dark.BgBase, dark.BgLayout = ui.Hex("#010102"), ui.Hex("#010102")
	dark.BgContainer = ui.Hex("#0f1011")
	dark.BgContainerSecondary = ui.Hex("#141516")
	dark.BgElevated = ui.Hex("#191a1b")
	dark.FillQuaternary, dark.FillTertiary = ui.Hex("#141516"), ui.Hex("#18191a")
	dark.FillSecondary, dark.Fill = ui.Hex("#191a1b"), ui.Hex("#23252a")
	dark.Border, dark.BorderSecondary, dark.Split = ui.Hex("#34343a"), ui.Hex("#23252a"), ui.Hex("#23252a")
	dark.Text, dark.TextBase, dark.TextHeading = ui.Hex("#f7f8f8"), ui.Hex("#f7f8f8"), ui.Hex("#f7f8f8")
	dark.TextLabel = dark.Text
	dark.TextSecondary = ui.Hex("#8a8f98")
	dark.TextTertiary, dark.TextDescription, dark.TextPlaceholder = dark.TextSecondary, dark.TextSecondary, dark.TextSecondary
	dark.Icon, dark.IconHover = dark.TextSecondary, dark.Text
	light.BgLayout, light.BgContainer = ui.Hex("#ffffff"), ui.Hex("#f5f6f6")
	light.BgContainerSecondary = ui.Hex("#f6f7f7")
	return theme.NewConfig(*light, *dark)
}
