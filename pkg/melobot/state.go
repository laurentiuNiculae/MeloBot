package melobot

type BotState int

const (
	Starting BotState = iota
	Connecting
	LoggingIn
	Serving
	Restarting
)
