package melobot_test

import (
	"fmt"
	"testing"

	"gopkg.in/irc.v4"
)

func TestTest(t *testing.T) {
	msg := &irc.Message{
		Command: "PRIVMSG",
		Params:  []string{"#themelopeus", "Hello, chat!"},
	}
	x := []byte(msg.String() + "\r\n")
	fmt.Println(x)
}
