package melobot

import (
	"crypto/tls"
	"fmt"
	"os"

	"gopkg.in/irc.v4"
)

func F() {
	conn, err := tls.Dial("tcp", TwitchIRSAddress, nil)
	if err != nil {
		fmt.Println("Error tdl Dial")

		panic(err)
	}

	go ListenReplies(conn)

	_, err = conn.Write([]byte("PASS oauth:nrj8auo57q3pfse5289kq4xgco60pq" + "\r\n"))
	if err != nil {
		fmt.Println("Error Writing PASS")

		panic(err)
	}

	_, err = conn.Write([]byte("NICK melopeusbot" + "\r\n"))
	if err != nil {
		fmt.Println("Error Writing NICK")

		panic(err)
	}

	_, err = conn.Write([]byte("JOIN #themelopeus" + "\r\n"))
	if err != nil {
		fmt.Println("Error Writing NICK")

		panic(err)
	}

	_, err = conn.Write([]byte("PRIVMSG #themelopeus :Hello, chat!\r\n"))
	if err != nil {
		fmt.Println("Error Writing PRIVMSG")

		panic(err)
	}

	c := make(chan os.Signal, 1)
	<-c
}

func ListenReplies(conn *tls.Conn) {
	ircReader := irc.NewReader(conn)

	for {

		ircMsg, err := ircReader.ReadMessage()
		if err != nil {
			continue
		}

		if ircMsg != nil {
			fmt.Println(ircMsg.String())
		}
	}
}
