package cmd

import (
	"GophKeeper/client/internal/config"
	contextutils "GophKeeper/client/internal/context_utils"

	"github.com/spf13/cobra"
)

// var Config config.Config

var rootCmd = &cobra.Command{
	Use:   "gophkeeper",
	Short: "CLI клиент для GophKeeper",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()

		ctx := cmd.Context()
		ctx = contextutils.WithDatabaseURL(ctx, cfg.DatabaseURI)
		ctx = contextutils.WithTokenPath(ctx, cfg.TokenPath)
		ctx = contextutils.WithServerAddr(ctx, cfg.RunAddress)
		cmd.SetContext(ctx)
	},
}

func Execute() error {
	return rootCmd.Execute()
}
