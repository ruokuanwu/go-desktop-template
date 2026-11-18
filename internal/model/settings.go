package model

type Settings struct {
	Language    string      `json:"language"`
	Theme       string      `json:"theme"`
	CloseToTray bool        `json:"close_to_tray"`
	Log         LogSettings `json:"log"`
}

type LogSettings struct {
	MaxAge int    `json:"max_age"`
	Level  string `json:"level"`
}
