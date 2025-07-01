package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

var authToken string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Аутентификация пользователя",
	Run: func(cmd *cobra.Command, args []string) {
		login, _ := cmd.Flags().GetString("login")
		password, _ := cmd.Flags().GetString("password")

		if login == "" || password == "" {
			fmt.Println("Пожалуйста, укажите --login и --password")
			return
		}

		data := map[string]string{"login": login, "password": password}
		body, _ := json.Marshal(data)

		resp, err := http.Post(Config.RunAddress+"/api/login", "application/json", bytes.NewReader(body))
		if err != nil {
			fmt.Println("Ошибка запроса:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var respData map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&respData); err == nil {
				Token = respData["token"]
				fmt.Println("Вход выполнен успешно!")
				// TODO: Сохраняем токен в файл для дальнейшего использования
			} else {
				fmt.Println("Не удалось получить токен")
			}
		} else {
			b, _ := io.ReadAll(resp.Body)
			fmt.Println("Ошибка входа:", string(b))
		}
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().String("login", "", "Логин пользователя")
	loginCmd.Flags().String("password", "", "Пароль пользователя")
}
