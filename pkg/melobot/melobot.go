package melobot

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	retry "github.com/avast/retry-go"
	"gopkg.in/irc.v4"

	mbErrors "github.com/laurentiuNiculae/MeloBot/pkg/errors"
	"github.com/laurentiuNiculae/MeloBot/pkg/log"
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

	StateChange    StateChange
	RestartCounter *sync.WaitGroup
}

func New(pass, nick, channel string, log log.MeloLog) (*MeloBot, error) {
	return &MeloBot{
		Pass:           pass,
		Nick:           nick,
		Channel:        channel,
		Log:            log,
		State:          Starting,
		Incoming:       make(chan *irc.Message, 50),
		StateChange:    make(StateChange, 2),
		RestartCounter: &sync.WaitGroup{},
	}, nil
}

func (mb *MeloBot) Start() {
	var (
		ctx             context.Context
		cancelContext   context.CancelFunc
		systemInterrupt = make(chan os.Signal, 10)
	)

	signal.Notify(systemInterrupt, os.Interrupt)

	mb.StateChange <- Starting

	for {
		select {
		case newState := <-mb.StateChange:
			switch newState {
			case Starting:
				ctx, cancelContext = context.WithCancel(context.Background())

				err := retry.Do(
					func() error {
						conn, err := tls.Dial("tcp", TwitchIRSAddress, nil)
						if err != nil {
							mb.Log.Errorf("Error occured while re")

							return err
						}

						mb.Conn = conn

						return nil
					},
					retry.DelayType(retry.BackOffDelay),
					retry.Context(ctx),
					retry.Attempts(100),
				)

				if err != nil {
					mb.Log.Error("Connection failed, retrying agaain")

					continue
				}

				go mb.ListenReplies(ctx)

				mb.StateChange <- LoggingIn
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

				mb.StateChange <- Serving
				mb.State = Serving
			case Serving:
				go mb.StartCommandHandler(ctx)

				err := mb.Say("Hello, chat!")
				if err != nil {
					panic(err)
				}
			case Restarting:
				cancelContext()
				mb.RestartCounter.Wait()
				mb.Log.Info("Melobot is restarting")

				mb.StateChange <- Starting
				mb.State = Starting
			case Closing:
				cancelContext()
				mb.RestartCounter.Wait()
				mb.Log.Info("Melobot is closed")

				os.Exit(0)
			default:
				panic("all states cases should be handled")
			}
		case <-systemInterrupt:
			mb.Log.SetEnabled(false)

			reader := bufio.NewReader(os.Stdin)

			var repeatActionSelection bool = true

			fmt.Println("\nChoose action: ")
			fmt.Println("   - restart (r)")
			fmt.Println("   - quit (q)")

			for repeatActionSelection {
				action, err := reader.ReadString('\n')
				if err != nil {
					continue
				}

				action = strings.TrimSpace(action)

				switch action {
				case "restart", "r":
					repeatActionSelection = false

					err := mb.Conn.Close()
					if err != nil {
						fmt.Println("Failed to restart, try again.")
						repeatActionSelection = true
					}

					mb.StateChange <- Restarting
				case "quit", "q":
					repeatActionSelection = false

					err := mb.Conn.Close()
					if err != nil {
						fmt.Println("Failed to gracefully quit. Force quit.")
						os.Exit(0)
					}

					mb.StateChange <- Closing
				default:
					fmt.Println("Please retry :)")
				}
			}

			mb.Log.SetEnabled(true)
		}
	}
}

func (mb *MeloBot) StartCommandHandler(ctx context.Context) {
	mb.Log.Debug("Start listening to commands from chat.")
	mb.RestartCounter.Add(1)

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
			mb.Log.Debug("Closing command handler")
			mb.RestartCounter.Done()

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
	mb.Log.Info("Starting to listen to replyes.")
	mb.RestartCounter.Add(1)

	ircReader := irc.NewReader(mb.Conn)

	for {
		select {
		case <-ctx.Done():
			mb.Log.Debug("EXIT: Closing IRC reply listener")
			mb.RestartCounter.Done()

			return
		default:
			ircMsg, err := ircReader.ReadMessage()
			if err != nil {
				continue
			}

			mb.Log.Debugf("%s: %s", ircMsg.Command, ircMsg.Trailing())

			mb.Incoming <- ircMsg
		}
	}
}

func (mb *MeloBot) SendCredentials(ctx context.Context) error {
	mb.Log.Info("Sending credentials")

	err := mb.SendIRC(&irc.Message{Command: "PASS", Params: []string{mb.Pass}})
	if err != nil {
		mb.Log.Error("Error Writing PASS")

		return fmt.Errorf("%w : %w", mbErrors.ErrFailedIRCSend, err)
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
				if strings.Contains(confirmMsg.String(), "Login authentication failed") {
					mb.Log.Error("Connection FAILED because of bad credentials")

					return mbErrors.ErrBadCredentials
				}
			}
		case <-timeout:
			mb.Log.Error("Connection timeout")

			return mbErrors.ErrTimeout
		case <-ctx.Done():
			mb.Log.Info("Program is being closed")

			return nil
		}
	}
}

func (mb *MeloBot) JoinChannel(ctx context.Context) error {
	mb.Log.Infof("Joinging channel %s", mb.Channel)

	err := mb.SendIRC(&irc.Message{Command: "JOIN", Params: []string{mb.Channel}})
	if err != nil {
		mb.Log.Error("Error sending JOIN")

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
			mb.Log.Errorf("Joing channel TIMEOUT, make sure the channel '%s' is spelled right or that it exists", mb.Channel)

			return mbErrors.ErrTimeout
		case <-ctx.Done():
			mb.Log.Error("Program is being closed ")

			return nil
		}
	}
}
