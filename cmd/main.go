package main

import (
	"fmt"
	"os"

	"github.com/laurentiuNiculae/MeloBot/pkg/credentials"
	"github.com/laurentiuNiculae/MeloBot/pkg/flags"
	"github.com/laurentiuNiculae/MeloBot/pkg/log"
	"github.com/laurentiuNiculae/MeloBot/pkg/melobot"
)

func main() {
	flags := flags.GetFlags()
	log := log.New(os.Stdout)

	log.Info("Starting Twitch Bot")
	log.Error("Starting Twitch Bot")
	log.Debug("Starting Twitch Bot")

	twitchCredentials, err := credentials.GetCredentials(flags)
	if err != nil {
		log.Error("Coudn't get credentials: " + err.Error())
		panic("couldn't get credentials from either file or env")
	}

	melobot, err := melobot.New(
		twitchCredentials.OauthToken,
		twitchCredentials.Nick,
		twitchCredentials.Channel,
		log,
	)

	if err != nil {
		panic(fmt.Errorf("errors creating bot: %w", err))
	}

	melobot.Start()
}
