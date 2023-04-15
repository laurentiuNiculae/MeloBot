package main

import (
	"fmt"

	"github.com/laurentiuNiculae/MeloBot/cmd"
)

func main() {
	melobot, err := cmd.NewMeloBot(
		"oauth:egg5z1bi7td00bapv3aien22gphni1",
		"melopeusbot",
		"#themelopeus",
	)

	if err != nil {
		panic(fmt.Errorf("errors creating bot: %w", err))
	}

	melobot.Start()
}
