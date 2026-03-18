package cmd

import (
	"GophKeeper/client/internal/api"
	contextutils "GophKeeper/client/internal/context_utils"
	"GophKeeper/client/internal/storage"
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var getFileID string

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Скачать файл с сервера",
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

		var fileID uuid.UUID
		if getFileID == "" {
			err = fmt.Errorf("пожалуйста, укажите ID файла через --id")
			return err
		} else {
			fileID, err = uuid.Parse(getFileID)
			if err != nil {
				return fmt.Errorf("неверный формат ID файла: %w", err)
			}
		}

		client := api.NewClient(storage, serverAddr)
		err = client.GetFile(ctx, fileID)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringVar(&getFileID, "id", "", "ID файла для загрузки")
	getCmd.MarkFlagRequired("id")
}
