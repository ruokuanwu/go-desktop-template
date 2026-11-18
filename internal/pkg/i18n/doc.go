// Package i18n provides UI-thread-owned Chinese and English localization using
// go-i18n/v2. JSON catalogs in locales are embedded into the executable.
//
// Settings accept system (the default), zh-CN and en. An empty setting from older
// callers also means system. System language preferences are detected at startup;
// unsupported locales fall back to English. Language changes apply after settings
// are saved successfully, without rebuilding page state or translating user data.
//
// Add stable message IDs to both catalogs and query them with Translator.T.
// Translate whole sentences; use named template data instead of concatenation.
// Translator.Tasks demonstrates plural forms. Missing English messages or invalid
// templates fail visibly rather than silently hiding translation errors.
//
// This follows the embedded JSON bundle approach used by Fyne's lang package:
// https://github.com/fyne-io/fyne/tree/master/lang
package i18n
