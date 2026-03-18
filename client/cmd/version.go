package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	buildVersion string = "dev"
	buildDate    string = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Показать версию и дату сборки",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GophKeeper CLI\nВерсия: %s\nСборка: %s\n", buildVersion, buildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
