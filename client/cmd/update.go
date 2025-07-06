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

var updateFilePath, updateMetadata, updateFileID string

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Обновить существующий файл",
	RunE: func(cmd *cobra.Command, args []string) error {

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

		if updateFilePath == "" {
			err = fmt.Errorf("пожалуйста, укажите путь к файлу через --path")
			return err
		}

		var fileID uuid.UUID
		if updateFileID == "" {
			err = fmt.Errorf("пожалуйста, укажите ID файла через --id")
			return err
		} else {
			fileID, err = uuid.Parse(updateFileID)
			if err != nil {
				return fmt.Errorf("неверный формат ID файла: %w", err)
			}
		}

		client := api.NewClient(storage, serverAddr)
		err = client.UpdateFile(ctx, fileID, updateFilePath, updateMetadata)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateFilePath, "path", "", "Путь к файлу для загрузки")
	updateCmd.Flags().StringVar(&updateMetadata, "meta", "", "Метаданные для файла")
	updateCmd.Flags().StringVar(&updateFileID, "id", "", "ID файла для обновления")
	updateCmd.MarkFlagRequired("path")
	updateCmd.MarkFlagRequired("id")
	rootCmd.AddCommand(updateCmd)
}
