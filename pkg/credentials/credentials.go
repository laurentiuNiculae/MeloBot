package credentials

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/laurentiuNiculae/MeloBot/pkg/flags"
)

const (
	MeloBotTwitchOauthToken = "MELOBOT_TWITCH_OAUTH_TOKEN"
	MeloBotTwitchNickname   = "MELOBOT_TWITCH_NICKNAME"
	MeloBotTwitchChannel    = "MELOBOT_TWITCH_CHANNEL"
)

type TwitchCredentials struct {
	OauthToken string
	Nick       string
	Channel    string
}

func GetCredentials(flags flags.Flags) (TwitchCredentials, error) {
	if flags.CredentialFilePath != nil && *flags.CredentialFilePath != "" {
		return getCredentialsFromFile(*flags.CredentialFilePath)
	}

	return getCredentialsFromEnv()
}

func getCredentialsFromFile(filePath string) (TwitchCredentials, error) {
	fileBlob, err := os.ReadFile(filePath)
	if err != nil {
		return TwitchCredentials{}, err
	}

	var twitchCredentials TwitchCredentials

	err = json.Unmarshal(fileBlob, &twitchCredentials)
	if err != nil {
		return TwitchCredentials{}, err
	}

	return twitchCredentials, nil
}

func getCredentialsFromEnv() (TwitchCredentials, error) {
	oauthToken, ok := os.LookupEnv(MeloBotTwitchOauthToken)
	if !ok {
		panic(fmt.Sprintf("Missing ENV Variable '%s'", MeloBotTwitchOauthToken))
	}

	nickname, ok := os.LookupEnv(MeloBotTwitchNickname)
	if !ok {
		panic(fmt.Sprintf("Missing ENV Variable '%s'", MeloBotTwitchNickname))
	}

	channel, ok := os.LookupEnv(MeloBotTwitchChannel)
	if !ok {
		panic(fmt.Sprintf("Missing ENV Variable '%s'", MeloBotTwitchChannel))
	}

	return TwitchCredentials{
		OauthToken: oauthToken,
		Nick:       nickname,
		Channel:    channel,
	}, nil
}
