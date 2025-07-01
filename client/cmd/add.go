package cmd

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Добавить новую запись из файла",
	Run: func(cmd *cobra.Command, args []string) {
		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			fmt.Println("Пожалуйста, укажите путь к файлу через --file")
			return
		}

		token := getToken()

		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println("Ошибка открытия файла:", err)
			return
		}
		defer file.Close()

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		part, err := writer.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			fmt.Println("Ошибка создания multipart части:", err)
			return
		}

		if _, err := io.Copy(part, file); err != nil {
			fmt.Println("Ошибка копирования файла в multipart:", err)
			return
		}

		writer.Close()

		req, err := http.NewRequest("POST", Config.RunAddress+"/data/upload", &body)
		if err != nil {
			fmt.Println("Ошибка создания запроса:", err)
			return
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Ошибка запроса:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated {
			fmt.Println("Файл успешно загружен!")
		} else {
			b, _ := io.ReadAll(resp.Body)
			fmt.Printf("Ошибка загрузки файла: %s\n", string(b))
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().String("file", "", "Путь к файлу для загрузки")
}
