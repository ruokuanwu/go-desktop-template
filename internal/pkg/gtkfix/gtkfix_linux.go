package gtkfix

import (
	"fmt"

	"github.com/ebitengine/purego"
)

// TextColor must run after GTK initializes and before the first window creates fonts.
// Older cairo can promote grayscale glyphs to subpixel glyphs, which MyGo
// mistakes for color emoji and draws without the requested text color.
func TextColor() error {
	current, err := purego.Dlsym(purego.RTLD_DEFAULT, "FcConfigGetCurrent")
	if err != nil {
		return fmt.Errorf("find fontconfig configuration: %w", err)
	}
	parse, err := purego.Dlsym(purego.RTLD_DEFAULT, "FcConfigParseAndLoadFromMemory")
	if err != nil {
		return fmt.Errorf("find fontconfig parser: %w", err)
	}
	var getCurrent func() uintptr
	var parseAndLoad func(uintptr, string, int32) int32
	purego.RegisterFunc(&getCurrent, current)
	purego.RegisterFunc(&parseAndLoad, parse)
	config := getCurrent()
	if config == 0 {
		return fmt.Errorf("fontconfig configuration is unavailable")
	}
	if parseAndLoad(config, fontconfig, 1) == 0 {
		return fmt.Errorf("load fontconfig text color workaround")
	}
	return nil
}

// Clear the matched subpixel order, not antialiasing or color-font support.
// MyGo's explicitly requested subpixel rendering retains its own settings.
const fontconfig = `<?xml version="1.0"?>
<!DOCTYPE fontconfig SYSTEM "urn:fontconfig:fonts.dtd">
<fontconfig>
  <match target="pattern"><edit name="rgba" mode="assign"><const>none</const></edit></match>
  <match target="font"><edit name="rgba" mode="assign"><const>none</const></edit></match>
</fontconfig>
`
