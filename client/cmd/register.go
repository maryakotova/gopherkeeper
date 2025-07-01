package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Регистрация нового пользователя",
	Run: func(cmd *cobra.Command, args []string) {
		login, _ := cmd.Flags().GetString("login")
		password, _ := cmd.Flags().GetString("password")

		if login == "" || password == "" {
			fmt.Println("Пожалуйста, укажите --login и --password")
			return
		}

		data := map[string]string{"login": login, "password": password}
		body, _ := json.Marshal(data)

		resp, err := http.Post(Config.RunAddress+"/api/register", "application/json", bytes.NewReader(body))
		if err != nil {
			fmt.Println("Ошибка запроса:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated {
			fmt.Println("Регистрация прошла успешно!")
		} else {
			b, _ := io.ReadAll(resp.Body)
			fmt.Println("Ошибка регистрации:", string(b))
		}
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)

	registerCmd.Flags().String("login", "", "Логин пользователя")
	registerCmd.Flags().String("password", "", "Пароль пользователя")
}
