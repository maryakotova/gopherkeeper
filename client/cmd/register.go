package cmd

import (
	"GophKeeper/client/internal/api"
	contextutils "GophKeeper/client/internal/context_utils"
	"fmt"

	"github.com/spf13/cobra"
)

var regUsername, regPassword string

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Регистрация нового пользователя",
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx := cmd.Context()
		serverAddr, err := contextutils.GetServerAddr(ctx)
		if err != nil {
			return err
		}

		client := api.NewClient(nil, serverAddr)
		err = client.Register(ctx, regUsername, regPassword)
		if err != nil {
			return fmt.Errorf("ошибка регистрации: %w", err)
		}
		fmt.Println("Пользователь успешно зарегистрирован")

		return nil

	},
}

func init() {
	registerCmd.Flags().StringVar(&regUsername, "username", "", "Имя пользователя")
	registerCmd.Flags().StringVar(&regPassword, "password", "", "Пароль")
	registerCmd.MarkFlagRequired("username")
	registerCmd.MarkFlagRequired("password")
	rootCmd.AddCommand(registerCmd)
}
