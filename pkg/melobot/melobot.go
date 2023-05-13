package melobot

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/laurentiuNiculae/MeloBot/pkg/log"
	"gopkg.in/irc.v4"
)

const (
	TwitchIRSAddress = "irc.chat.twitch.tv:6697"
	TwitchIrcChannel = "#themelopeus"
	TwitchChannel    = "themelopeus"
)

type MeloBot struct {
	Pass     string
	Nick     string
	Channel  string
	Conn     *tls.Conn
	Log      log.MeloLog
	State    BotState
	Incoming chan *irc.Message

	reconectDelay time.Duration
}

func New(pass, nick, channel string, log log.MeloLog) (*MeloBot, error) {
	return &MeloBot{
		Pass:          pass,
		Nick:          nick,
		Channel:       channel,
		Log:           log,
		State:         Starting,
		Incoming:      make(chan *irc.Message),
		reconectDelay: time.Second,
	}, nil
}

func (mb *MeloBot) Start() {
	ctx := context.Background()

	for {
		switch mb.State {
		case Starting:
			conn, err := tls.Dial("tcp", TwitchIRSAddress, nil)
			if err != nil {
				mb.Log.Errorf("Error creating connection, retrying after %v", mb.reconectDelay.String())
				mb.reconectDelay = mb.reconectDelay * 2
				mb.State = Starting

				break
			}

			mb.reconectDelay = time.Second
			mb.Conn = conn

			go mb.ListenReplies(ctx)

			mb.State = LoggingIn
		case LoggingIn:
			err := mb.SendCredentials(ctx)
			if err != nil {
				panic(err)
			}

			err = mb.JoinChannel(ctx)
			if err != nil {
				panic(err)
			}

			mb.State = Serving
		case Serving:
			go mb.StartCommandHandler(ctx)

			err := mb.Say("Hello, chat!")
			if err != nil {
				panic(err)
			}
		case Restarting:
			ctx.Done()
			mb.State = Starting
		default:
		}
	}

	// c := make(chan os.Signal, 10)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// // wait until we get an interrupt call, then we'll close the main program context
	// <-c
	// ctx.Done()
}

func (mb *MeloBot) StartCommandHandler(ctx context.Context) {
	mb.Log.Debug("Start listening to commands from chat.")

	for {
		select {
		case ircMsg := <-mb.Incoming:
			switch ircMsg.Command {
			case "PRIVMSG":
				// TODO: maybe implement a structure that manages this, command, parameters etc
				// TODO: clean the command of white spaces and things like that
				command := strings.Split(ircMsg.Trailing(), " ")

				switch strings.ToLower(command[0]) {
				case "!hello":
					mb.Sayf("Hello back @%s", ircMsg.Name)
				case "!rand":
					mb.Say(ircMsg.Name + " your number is: " + strconv.FormatUint(rand.Uint64(), 10))
				case "!dice":
					mb.Say(ircMsg.Name + " your number is: " + strconv.FormatUint(rand.Uint64()%6, 10))
				case "!rand_b":
					if len(command) == 1 {
						mb.Say("I need a base input like this: \"!rand_b 2\"")
						continue
					}

					base, err := strconv.Atoi(command[1])

					if err != nil {
						mb.Sayf("Oops! The number given \"%s\" isn't a int or can't be parsed.", command[1])
						continue
					}

					mb.Say(ircMsg.Name + " your number is: " + strconv.FormatUint(rand.Uint64(), base))
				default:
					continue
				}
			case "PING":
				err := mb.SendIRC(&irc.Message{Command: "PONG", Params: []string{"tmi.twitch.tv"}})
				if err != nil {
					mb.Log.Error("Error when sending PONG message: " + err.Error())
				}

				mb.Log.Info("PONG")
			}

			continue
		case <-ctx.Done():
			mb.Log.Info("X Closing command handler")

			return
		}
	}
}

func (mb *MeloBot) Say(msg string) error {
	ircMsg := &irc.Message{Command: "PRIVMSG", Params: []string{mb.Channel, msg}}

	payload := IRCPayload(ircMsg)

	_, err := mb.Conn.Write(payload)
	if err != nil {
		return err
	}

	return nil
}

func (mb *MeloBot) Sayf(format string, params ...any) error {
	msg := fmt.Sprintf(format, params...)
	ircMsg := &irc.Message{Command: "PRIVMSG", Params: []string{mb.Channel, msg}}

	payload := IRCPayload(ircMsg)

	_, err := mb.Conn.Write(payload)
	if err != nil {
		return err
	}

	return nil
}

func (mb *MeloBot) SendIRC(ircMsg *irc.Message) error {
	payload := IRCPayload(ircMsg)

	_, err := mb.Conn.Write(payload)
	if err != nil {
		return err
	}

	return nil
}

func (mb *MeloBot) ListenReplies(ctx context.Context) {
	mb.Log.Debug("Starting to listen to replyes.")

	ircReader := irc.NewReader(mb.Conn)

	for {
		select {
		default:
			ircMsg, err := ircReader.ReadMessage()
			if err != nil {
				continue
			}

			mb.Log.Debugf("%s: %s", ircMsg.Command, ircMsg.Trailing())

			mb.Incoming <- ircMsg
		case <-ctx.Done():
			mb.Log.Debug("EXIT: Closing IRC reply listener")

			return
		}
	}
}

func (mb *MeloBot) SendCredentials(ctx context.Context) error {
	err := mb.SendIRC(&irc.Message{Command: "PASS", Params: []string{mb.Pass}})
	if err != nil {
		mb.Log.Error("Error Writing PASS")

		return err
	}

	err = mb.SendIRC(&irc.Message{Command: "NICK", Params: []string{mb.Nick}})
	if err != nil {
		mb.Log.Error("Error Writing NICK")

		return err
	}

	// wait for confirmation
	timeout := time.After(5 * time.Second)

	for {
		select {
		case confirmMsg := <-mb.Incoming:
			// mb.Log.Debug("COMMAND: " + confirmMsg.Command + " " + confirmMsg.Param(0))

			switch confirmMsg.Command {
			case irc.RPL_WELCOME:
				mb.Log.Info("Connection SUCCESS")

				return nil
			case "NOTICE":
				mb.Log.Error("Connection FAILED")

				if strings.Contains(confirmMsg.String(), "Login authentication failed") {
					panic("Authentification failed")
				}
			}
		case <-timeout:
			mb.Log.Error("Connection timeout")

			return fmt.Errorf("login failed")
		case <-ctx.Done():
			mb.Log.Info("Program is being closed")

			return nil
		}
	}
}

func (mb *MeloBot) JoinChannel(ctx context.Context) error {
	_, err := mb.Conn.Write([]byte("JOIN #themelopeus" + "\r\n"))
	if err != nil {
		fmt.Println("LOG: Error sending JOIN")

		panic(err)
	}

	timeout := time.After(5 * time.Second)

	for {
		select {
		case confirmMsg := <-mb.Incoming:
			if confirmMsg.Command == irc.RPL_NAMREPLY {
				mb.Log.Info("Joing channel SUCCESSFUL")

				return nil
			}
		case <-timeout:
			mb.Log.Error("Connection TIMEOUT")

			return fmt.Errorf("login failed")
		case <-ctx.Done():
			mb.Log.Error("Program is being closed ")

			fmt.Println("EXIT: Program is being closed")

			return nil
		}
	}
}
