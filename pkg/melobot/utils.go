package melobot

import "gopkg.in/irc.v4"

func IRCPayload(ircMsg *irc.Message) []byte {
	return []byte(ircMsg.String() + "\n\r")
}
