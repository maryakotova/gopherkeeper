package cmd

import (
	"GophKeeper/client/internal/api"
	contextutils "GophKeeper/client/internal/context_utils"
	"GophKeeper/client/internal/storage"
	"fmt"

	"github.com/spf13/cobra"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var addFilePath, addMetadata string

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Добавить новый файл",
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

		if addFilePath == "" {
			err = fmt.Errorf("пожалуйста, укажите путь к файлу через --path")
			return err
		}

		client := api.NewClient(storage, serverAddr)
		err = client.AddFile(ctx, addFilePath, addMetadata)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addFilePath, "path", "", "Путь к файлу для загрузки")
	addCmd.Flags().StringVar(&addMetadata, "meta", "", "Метаданные для файла")
	addCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(addCmd)
}
