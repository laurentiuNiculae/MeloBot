package melobot

type BotState int

const (
	Starting BotState = iota
	Connecting
	LoggingIn
	Serving
	Restarting
	Closing
)

type StateChange chan BotState
