package cmd

import (
	"GophKeeper/client/internal/api"
	contextutils "GophKeeper/client/internal/context_utils"
	"GophKeeper/client/internal/storage"

	"github.com/spf13/cobra"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Синхронизация данных в сервером",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		ctx := cmd.Context()
		serverAddr, err := contextutils.GetServerAddr(ctx)
		if err != nil {
			return err
		}

		dbDSN, err := contextutils.GetDatabaseURL(ctx)
		if err != nil {
			return err
		}

		factory := &storage.StorageFactory{}
		storage, err := factory.NewStorage(dbDSN)
		if err != nil {
			return err
		}

		client := api.NewClient(storage, serverAddr)
		err = client.Sync(ctx)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
