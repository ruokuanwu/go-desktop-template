package i18n

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/jeandeaual/go-locale"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var translations embed.FS

// Translator is owned by the UI thread. Translation resources are loaded once.
type Translator struct {
	localizers map[string]*goi18n.Localizer
	current    *goi18n.Localizer
}

func New(selected string) (*Translator, error) {
	bundle := goi18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	for _, name := range []string{"en", "zh-CN"} {
		if _, err := bundle.LoadMessageFileFS(translations, "locales/"+name+".json"); err != nil {
			return nil, fmt.Errorf("load translations: %w", err)
		}
	}
	languages, err := locale.GetLocales()
	if err != nil {
		zap.L().Warn("detect system language; using English", zap.Error(err))
	}
	t := &Translator{localizers: map[string]*goi18n.Localizer{
		"en":     goi18n.NewLocalizer(bundle, "en"),
		"zh-CN":  goi18n.NewLocalizer(bundle, "zh-CN"),
		"system": goi18n.NewLocalizer(bundle, languages...),
	}}
	t.SetLanguage(selected)
	return t, nil
}

func (t *Translator) SetLanguage(selected string) {
	if selected == "" {
		selected = "system"
	}
	t.current = t.localizers[selected]
}

func (t *Translator) T(id string, data ...any) string {
	config := &goi18n.LocalizeConfig{MessageID: id}
	if len(data) > 0 {
		config.TemplateData = data[0]
	}
	return t.current.MustLocalize(config)
}

func (t *Translator) Tasks(count int) string {
	return t.current.MustLocalize(&goi18n.LocalizeConfig{
		MessageID: "todos.count", PluralCount: count,
		TemplateData: map[string]any{"Count": count},
	})
}
