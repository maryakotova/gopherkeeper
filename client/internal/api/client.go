package api

import (
	contextutils "GophKeeper/client/internal/context_utils"
	"GophKeeper/client/internal/models"
	"GophKeeper/client/internal/storage"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	client     *http.Client
	repo       storage.Repository
	serverAddr string
}

func NewClient(repo storage.Repository, serverAddr string) *Client {
	return &Client{
		client:     &http.Client{},
		repo:       repo,
		serverAddr: serverAddr,
	}
}

// Запрос на регистрацию
func (c *Client) Register(ctx context.Context, username, password string) error {
	reqBody := models.AuthRequest{
		Username: username,
		Password: password,
	}
	data, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+c.serverAddr+"/api/register", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ошибка %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Запрос на аутентификацию
func (c *Client) Login(ctx context.Context, username, password string) (string, error) {

	reqBody := models.AuthRequest{
		Username: username,
		Password: password,
	}
	data, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+c.serverAddr+"/api/login", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ошибка %d: %s", resp.StatusCode, string(body))
	}

	var authResp models.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", err
	}
	return authResp.Token, nil

}

// Запрос на обновление файла
func (c *Client) UpdateFile(ctx context.Context, id uuid.UUID, filePath string, meta string) (err error) {

	if c.repo == nil {
		return fmt.Errorf("подключение к БД не установлено")
	}

	token, err := contextutils.LoadTokenFromFile(ctx)
	if err != nil {
		return err
	}

	updatedAt := time.Now()

	file, fileData, fileName, err := readAndPrepareFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	serverID, err := c.repo.GetServerIDByFileID(ctx, id)
	if err != nil {
		return fmt.Errorf("файл не найден в БД на стороне клиента: %w", err)
	}

	body, contentType, err := buildMultipartBodyForUpdate(file, fileName, meta, updatedAt, serverID)
	if err != nil {
		return err
	}

	err = c.sendUpdateRequest(ctx, &body, contentType, token)
	if err != nil {
		return err
	}

	if err := c.updateFileInRepo(ctx, id, fileData, fileName, meta, updatedAt); err != nil {
		return err
	}

	fmt.Println("✅ Файл успешно обновлён")
	return nil
}

// Запрос на добавление файла
func (c *Client) AddFile(ctx context.Context, filePath string, meta string) (err error) {

	if c.repo == nil {
		return fmt.Errorf("подключение к БД не установлено")
	}

	token, err := contextutils.LoadTokenFromFile(ctx)
	if err != nil {
		return err
	}

	createdAt := time.Now()

	file, fileData, fileName, err := readAndPrepareFile(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body, contentType, err := buildMultipartBody(file, fileName, meta, createdAt)
	if err != nil {
		return err
	}

	response, err := c.sendUploadRequest(ctx, &body, contentType, token)
	if err != nil {
		return err
	}

	fileID, err := uuid.Parse(response.FileID)
	if err != nil {
		return fmt.Errorf("ошибка при преобразовании ID файла: %w", err)
	}

	if err := c.saveFileToRepo(ctx, fileData, fileName, meta, fileID, createdAt); err != nil {
		return err
	}

	fmt.Println("Файл успешно загружен!")
	return nil
}

// Запрос на получение файла
func (c *Client) GetFile(ctx context.Context, id uuid.UUID) (err error) {

	if c.repo == nil {
		return fmt.Errorf("подключение к БД не установлено")
	}

	token, err := contextutils.LoadTokenFromFile(ctx)
	if err != nil {
		return err
	}

	resp, err := c.sendDownloadRequest(ctx, token, id)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fileData, fileName, meta, updatedAt, err := parseMultipartResponse(resp)
	if err != nil {
		return err
	}

	if err := c.repo.UpdateFile(ctx, id, fileData, fileName, meta, updatedAt); err != nil {
		return fmt.Errorf("ошибка обновления в локальной базе: %w", err)
	}

	fmt.Println("Файл успешно загружен!")
	return nil

}

func (c *Client) Sync(ctx context.Context) (err error) {

	newLastSync := time.Now()

	if c.repo == nil {
		return fmt.Errorf("подключение к БД не установлено")
	}

	lastSync := c.repo.GetLastSync(ctx)

	resp, err := c.sendSyncRequest(ctx, lastSync)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent:
		fmt.Println("✅ Нечего синхронизировать — сервер не вернул данных")
		return nil
	case http.StatusOK:
		// продолжаем
	default:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("сервер вернул ошибку: %s", string(body))
	}

	err = c.processMultipartResponse(ctx, resp)
	if err != nil {
		return err
	}

	return c.repo.UpdateLastSync(ctx, newLastSync)

}

// Загрузка и подготовка файла
func readAndPrepareFile(filePath string) (file *os.File, fileData []byte, fileName string, err error) {

	file, err = os.Open(filePath)
	if err != nil {
		return nil, nil, "", fmt.Errorf("ошибка открытия файла: %w", err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		file.Close()
		return nil, nil, "", fmt.Errorf("ошибка чтения файла: %w", err)
	}
	if len(data) == 0 {
		file.Close()
		return nil, nil, "", fmt.Errorf("файл пустой")
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		file.Close()
		return nil, nil, "", fmt.Errorf("ошибка сброса позиции курсора: %w", err)
	}

	fileName = filepath.Base(filePath)
	return file, data, fileName, nil

}

// Сборка multipart запроса для Add
func buildMultipartBody(file *os.File, fileName, meta string, createdAt time.Time) (body bytes.Buffer, contentType string, err error) {

	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return body, "", fmt.Errorf("ошибка создания multipart: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return body, "", fmt.Errorf("ошибка копирования файла: %w", err)
	}

	if err := writer.WriteField("metadata", meta); err != nil {
		return body, "", fmt.Errorf("ошибка добавления метаданных: %w", err)
	}

	if err := writer.WriteField("created_at", createdAt.Format(time.RFC3339Nano)); err != nil {
		return body, "", fmt.Errorf("ошибка добавления времени: %w", err)
	}

	if err := writer.Close(); err != nil {
		return body, "", fmt.Errorf("ошибка закрытия writer: %w", err)
	}

	return body, writer.FormDataContentType(), nil
}

// Отправка запроса Add и обработка ответа
func (c *Client) sendUploadRequest(ctx context.Context, body *bytes.Buffer, contentType, token string) (*models.UploadResponse, error) {

	req, err := http.NewRequestWithContext(ctx, "POST", "http://"+c.serverAddr+"/data/upload", body)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", contentType)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("ошибка загрузки: %s", string(respBody))
	}

	var uploadResp models.UploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	if uploadResp.FileID == "" {
		return nil, fmt.Errorf("сервер не вернул file_id")
	}

	return &uploadResp, nil
}

// Сохранение файла в хранилище
func (c *Client) saveFileToRepo(ctx context.Context, fileData []byte, fileName, meta string, fileID uuid.UUID, createdAt time.Time) error {

	_, err := c.repo.InsertFile(ctx, fileData, fileName, meta, fileID, createdAt)
	if err != nil {
		return fmt.Errorf("ошибка сохранения в хранилище: %w", err)
	}
	return nil
}

// Сборка multipart запроса для Update
func buildMultipartBodyForUpdate(file *os.File, fileName, meta string, updatedAt time.Time, serverID uuid.UUID) (bytes.Buffer, string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return body, "", fmt.Errorf("ошибка создания multipart части: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return body, "", fmt.Errorf("ошибка копирования файла: %w", err)
	}

	if err := writer.WriteField("metadata", meta); err != nil {
		return body, "", fmt.Errorf("ошибка записи метаданных: %w", err)
	}

	if err := writer.WriteField("updated_at", updatedAt.Format(time.RFC3339Nano)); err != nil {
		return body, "", fmt.Errorf("ошибка записи updated_at: %w", err)
	}

	if err := writer.WriteField("file_id", serverID.String()); err != nil {
		return body, "", fmt.Errorf("ошибка записи file_id: %w", err)
	}

	if err := writer.Close(); err != nil {
		return body, "", fmt.Errorf("ошибка закрытия writer: %w", err)
	}

	return body, writer.FormDataContentType(), nil
}

// Отправка запроса Update и обработка ответа
func (c *Client) sendUpdateRequest(ctx context.Context, body *bytes.Buffer, contentType, token string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", "http://"+c.serverAddr+"/data/update", body)
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", contentType)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("ошибка чтения тела ответа: %w", err)
		}
		return fmt.Errorf("сервер вернул ошибку: %s", string(respBody))
	}

	return nil
}

// Обновление файла в хранелище
func (c *Client) updateFileInRepo(ctx context.Context, id uuid.UUID, data []byte, name, meta string, updatedAt time.Time) error {
	err := c.repo.UpdateFile(ctx, id, data, name, meta, updatedAt)
	if err != nil {
		return fmt.Errorf("ошибка обновления в локальной базе: %w", err)
	}
	return nil
}

// Извлечение границу (boundary) из заголовка Content-Type
func boundaryFromContentType(contentType string) string {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return ""
	}
	return params["boundary"]
}

// Отправка Get запроса и обработка ответа
func (c *Client) sendDownloadRequest(ctx context.Context, token string, id uuid.UUID) (*http.Response, error) {
	url := fmt.Sprintf("http://%s/data/download/%s", c.serverAddr, id.String())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("ошибка загрузки файла: %s", string(body))
	}

	return resp, nil
}

// Парсинг multipart ответа для Get
func parseMultipartResponse(resp *http.Response) ([]byte, string, string, time.Time, error) {
	contentType := resp.Header.Get("Content-Type")
	boundary := boundaryFromContentType(contentType)
	mr := multipart.NewReader(resp.Body, boundary)

	var (
		fileData  []byte
		fileName  string
		meta      string
		updatedAt time.Time
	)

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", "", time.Time{}, fmt.Errorf("ошибка чтения multipart: %w", err)
		}

		switch part.FormName() {
		case "file":
			fileName = part.FileName()
			fileData, err = io.ReadAll(part)
			if err != nil {
				return nil, "", "", time.Time{}, fmt.Errorf("ошибка чтения файла: %w", err)
			}

		case "meta":
			metaBytes, err := io.ReadAll(part)
			if err != nil {
				return nil, "", "", time.Time{}, fmt.Errorf("ошибка чтения метаданных: %w", err)
			}
			meta = string(metaBytes)

		case "updated_at":
			dateBytes, err := io.ReadAll(part)
			if err != nil {
				return nil, "", "", time.Time{}, fmt.Errorf("ошибка чтения даты обновления: %w", err)
			}
			updatedAt, err = time.Parse(time.RFC3339Nano, string(dateBytes))
			if err != nil {
				return nil, "", "", time.Time{}, err
			}
		}
	}

	return fileData, fileName, meta, updatedAt, nil
}

// Отправка Sync запроса
func (c *Client) sendSyncRequest(ctx context.Context, lastSync time.Time) (*http.Response, error) {
	token, err := contextutils.LoadTokenFromFile(ctx)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://%s/data/sync/%s", c.serverAddr, lastSync.Format(time.RFC3339Nano))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}

	return resp, nil
}

// Обработка multipart-ответа
func (c *Client) processMultipartResponse(ctx context.Context, resp *http.Response) error {
	contentType := resp.Header.Get("Content-Type")
	boundary := boundaryFromContentType(contentType)
	if boundary == "" {
		return fmt.Errorf("не удалось определить boundary из Content-Type: %s", contentType)
	}

	mr := multipart.NewReader(resp.Body, boundary)

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ошибка чтения multipart части: %w", err)
		}

		err = c.handlePart(ctx, part)
		if err != nil {
			return err
		}
	}

	return nil
}

// Обработка одной multipart части
func (c *Client) handlePart(ctx context.Context, part *multipart.Part) error {
	cd := part.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(cd)
	if err != nil {
		return fmt.Errorf("ошибка парсинга Content-Disposition: %w", err)
	}
	fileName := params["filename"]

	meta := part.Header.Get("meta")
	updatedAtStr := part.Header.Get("updated_at")
	createdAtStr := part.Header.Get("created_at")
	fileIDStr := part.Header.Get("file_id")

	updatedAt, err := time.Parse(time.RFC3339Nano, updatedAtStr)
	if err != nil {
		return fmt.Errorf("неверный формат времени обновления для файла %s", fileName)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
	if err != nil {
		return fmt.Errorf("неверный формат времени создания для файла %s", fileName)
	}
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		return fmt.Errorf("ошибка при преобразовании ID для файла %s: %w", fileName, err)
	}

	content, err := io.ReadAll(part)
	if err != nil {
		return fmt.Errorf("ошибка чтения содержимого файла: %w", err)
	}

	err = c.repo.SaveOrUpdate(ctx, fileID, fileName, content, meta, createdAt, updatedAt)
	if err != nil {
		return fmt.Errorf("ошибка сохранения файла %s в локальное хранилище: %w", fileName, err)
	}

	fmt.Printf("✅ Синхронизирован файл: %s\n", fileName)
	return nil
}
