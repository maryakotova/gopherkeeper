package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Скачать файл с сервера",
	Run: func(cmd *cobra.Command, args []string) {
		filename, _ := cmd.Flags().GetString("file")
		if filename == "" {
			fmt.Println("Пожалуйста, укажите имя файла для загрузки через --file")
			return
		}

		token := getToken()

		var fileID uuid.UUID

		url := fmt.Sprintf("%s/data/download/%s", Config.RunAddress, fileID)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println("Ошибка создания запроса:", err)
			return
		}

		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Ошибка запроса:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Ошибка получения файла: %s\n", string(body))
			return
		}

		outFile := filepath.Base(filename)
		f, err := os.Create(outFile)
		if err != nil {
			fmt.Println("Ошибка создания файла:", err)
			return
		}
		defer f.Close()

		_, err = io.Copy(f, resp.Body)
		if err != nil {
			fmt.Println("Ошибка сохранения файла:", err)
			return
		}

		fmt.Printf("Файл %s успешно сохранён\n", outFile)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().String("file", "", "Имя файла для загрузки")
}
