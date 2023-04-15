package cmd

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"gopkg.in/irc.v4"
)

const (
	TwitchIRSAddress = "irc.chat.twitch.tv:6697"
	TwitchIrcChannel = "#themelopeus"
	TwitchChannel    = "themelopeus"
)

type MeloBot struct {
	Pass    string
	Nick    string
	Channel string
	Conn    *tls.Conn

	Incoming chan *irc.Message
}

func NewMeloBot(pass, nick, channel string) (*MeloBot, error) {
	conn, err := tls.Dial("tcp", TwitchIRSAddress, nil)
	if err != nil {
		fmt.Println("Error tdl Dial")

		return nil, err
	}

	return &MeloBot{
		Pass:     pass,
		Nick:     nick,
		Channel:  channel,
		Conn:     conn,
		Incoming: make(chan *irc.Message),
	}, nil
}

func (mb *MeloBot) Start() {
	go mb.ListenReplies()

	err := mb.SendCredentials()
	if err != nil {
		panic(err)
	}

	err = mb.JoinChannel()
	if err != nil {
		panic(err)
	}

	err = mb.Say("Hello, chat!")
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	<-c
}

func (mb *MeloBot) Say(msg string) error {
	ircMsg := &irc.Message{
		Command: "PRIVMSG",
		Params:  []string{mb.Channel, msg},
	}

	payload := IRCPayload(ircMsg)

	_, err := mb.Conn.Write(payload)
	if err != nil {
		return err
	}

	return nil
}

func (mb *MeloBot) ListenReplies() {
	ircReader := irc.NewReader(mb.Conn)

	for {
		ircMsg, err := ircReader.ReadMessage()
		if err != nil {
			fmt.Println("Error reading", err)
		}

		if ircMsg != nil {
			mb.Incoming <- ircMsg

			fmt.Printf("%s: %s\n", ircMsg.User, ircMsg.Param(1))
		}
	}
}

func (mb *MeloBot) SendCredentials() error {
	_, err := mb.Conn.Write([]byte("PASS oauth:egg5z1bi7td00bapv3aien22gphni1" + "\r\n"))
	if err != nil {
		fmt.Println("Error Writing PASS")

		return err
	}

	_, err = mb.Conn.Write([]byte("NICK melopeusbot" + "\r\n"))
	if err != nil {
		fmt.Println("Error Writing NICK")

		return err
	}

	// wait for confirmation
	timeout := time.After(5 * time.Second)

	for {
		select {
		case confirmMsg := <-mb.Incoming:
			if confirmMsg.Command == irc.RPL_WELCOME {
				fmt.Println("Connection SUCCESS")

				return nil
			}
		case <-timeout:
			fmt.Println("Connection TIMEOUT")

			return fmt.Errorf("login failed")
		}
	}
}

func (mb *MeloBot) JoinChannel() error {
	_, err := mb.Conn.Write([]byte("JOIN #themelopeus" + "\r\n"))
	if err != nil {
		fmt.Println("Error sending JOIN")

		panic(err)
	}

	timeout := time.After(5 * time.Second)

	for {
		select {
		case confirmMsg := <-mb.Incoming:
			if confirmMsg.Command == irc.RPL_NAMREPLY {
				fmt.Println("Connection SUCCESS")

				return nil
			}
		case <-timeout:
			fmt.Println("COnnection TIMEOUT")

			return fmt.Errorf("login failed")
		}
	}
}

func F() {

}
