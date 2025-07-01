package cmd

import (
	"GophKeeper/client/internal/config"

	"github.com/spf13/cobra"
)

var Config config.Config
var Token string

var rootCmd = &cobra.Command{
	Use:   "gophkeeper",
	Short: "CLI клиент для GophKeeper",
}

func Execute() error {
	return rootCmd.Execute()
}

func getToken() string {
	return Token
}
