package tclient

const (
	AppBuiltin = "builtin"
	AppDesktop = "desktop"
	AppCustom  = "custom"
)

type App struct {
	AppID   int
	AppHash string
}

var Apps = map[string]App{
	// application created by iyear
	AppBuiltin: {AppID: 15055931, AppHash: "021d433426cbb920eeb95164498fe3d3"},
	// application created by tdesktop.
	// https://opentele.readthedocs.io/en/latest/documentation/authorization/api/#class-telegramdesktop
	AppDesktop: {AppID: 2040, AppHash: "b18441a1ff607e10a989891a5462e627"},
	// custom application
	AppCustom: {AppID: 27897654, AppHash: "41526aab858d7d7d903e68115a434871"},
}
