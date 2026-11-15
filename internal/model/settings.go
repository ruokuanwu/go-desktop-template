package model

type Settings struct {
	Theme       string      `json:"theme"`
	CloseToTray bool        `json:"close_to_tray"`
	Log         LogSettings `json:"log"`
}

type LogSettings struct {
	MaxAge int    `json:"max_age"`
	Level  string `json:"level"`
}
