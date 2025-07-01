package main

import (
	"GophKeeper/client/cmd"
	"GophKeeper/client/internal/config"
	"fmt"
	"os"
)

func main() {

	config := config.NewConfig()

	cmd.Config = *config

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
