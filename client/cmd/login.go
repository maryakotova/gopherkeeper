package cmd

import (
	"GophKeeper/client/internal/api"
	contextutils "GophKeeper/client/internal/context_utils"
	"fmt"

	"github.com/spf13/cobra"
)

var loginUsername, loginPassword string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Аутентификация пользователя",
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx := cmd.Context()
		serverAddr, err := contextutils.GetServerAddr(ctx)
		if err != nil {
			return err
		}

		client := api.NewClient(nil, serverAddr)
		token, err := client.Login(ctx, loginUsername, loginPassword)
		if err != nil {
			return fmt.Errorf("ошибка входа: %w", err)
		}

		err = contextutils.SaveTokenToFile(ctx, token)
		if err != nil {
			return fmt.Errorf("вход выполнен, но не удалось сохранить токен: %w", err)
		}

		fmt.Println("🔓 Вход выполнен успешно")
		return nil

	},
}

func init() {
	loginCmd.Flags().StringVar(&loginUsername, "username", "", "Имя пользователя")
	loginCmd.Flags().StringVar(&loginPassword, "password", "", "Пароль")
	loginCmd.MarkFlagRequired("login")
	loginCmd.MarkFlagRequired("password")
	rootCmd.AddCommand(loginCmd)
}
