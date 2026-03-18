package handlers

import (
	"GophKeeper/server/internal/service"
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxFileSize = 2 * 1024 * 1024 * 1024 // 2GB

type ControllerHandler struct {
	service service.ControllerService
}

func NewControllerHandler(s service.ControllerService) *ControllerHandler {
	return &ControllerHandler{
		service: s,
	}
}

// TODO: для больших файлов добавить MinIO

func (h *ControllerHandler) UploadHandler(c *gin.Context) {

	userID := c.GetInt("user_id")
	var createdAt time.Time
	var err error

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	if fileHeader.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is empty"})
		return
	}

	if fileHeader.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "размер файл преавышает 2 ГБ"})
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	createdAtStr := c.PostForm("created_at")
	if createdAtStr == "" {
		createdAt = time.Now()
	} else {
		createdAt, err = h.parseTime(createdAtStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "incorrect time format"})
			return
		}
	}

	metadata := c.PostForm("metadata")
	fileName := fileHeader.Filename

	id, err := h.service.Upload(c.Request.Context(), userID, data, fileName, metadata, createdAt)

	if err != nil {
		err = fmt.Errorf("[UploadHandler]: failed to save data: %w", err)
		log.Print(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"file_id": id})
}

func (h *ControllerHandler) DownloadHandler(c *gin.Context) {

	fileIDStr := c.Param("id")
	if fileIDStr == "" {
		c.String(http.StatusBadRequest, "ID файла не указан")
		return
	}

	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Неверный формат ID файла")
		return
	}

	userID := c.GetInt("user_id") // если используете JWT/авторизацию

	// Получение данных из вашей бизнес-логики
	fileData, fileName, meta, updatedAt, err := h.service.Download(c.Request.Context(), fileID, userID)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Не удалось загрузить файл: %v", err))
		return
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Добавляем файл
	filePart, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Ошибка формирования файла: %v", err))
		return
	}
	_, err = filePart.Write(fileData)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Ошибка записи данных файла: %v", err))
		return
	}

	// Добавляем метаданные
	if err := writer.WriteField("meta", meta); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Ошибка записи мета: %v", err))
		return
	}

	// Добавляем дату обновления (в формате RFC3339Nano)
	if err := writer.WriteField("updated_at", updatedAt.Format(time.RFC3339Nano)); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Ошибка записи времени: %v", err)))
		return
	}

	writer.Close()

	// Устанавливаем заголовки и отправляем ответ
	c.Header("Content-Type", writer.FormDataContentType())
	c.Header("Content-Length", fmt.Sprint(buf.Len()))
	c.Writer.WriteHeader(http.StatusOK)
	_, err = c.Writer.Write(buf.Bytes())
	if err != nil {
		log.Printf("Ошибка отправки ответа: %v", err)
	}

}

func (h *ControllerHandler) UpdateHandler(c *gin.Context) {

	userID := c.GetInt("user_id")
	var err error

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	if fileHeader.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is empty"})
		return
	}

	if fileHeader.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "размер файл преавышает 2 ГБ"})
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	var updatedAt time.Time
	updatedAtStr := c.PostForm("updated_at")
	updatedAt, err = h.parseTime(updatedAtStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "incorrect time format"})
		return
	}

	metadata := c.PostForm("metadata")
	fileName := fileHeader.Filename

	idStr := c.PostForm("file_id")
	if idStr == "" {
		err = fmt.Errorf("ID файла не заполнено")
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "file ID is empty"})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "incorrect file id format"})
		return
	}

	err = h.service.Update(c.Request.Context(), id, userID, data, fileName, metadata, updatedAt)

	if err != nil {
		err = fmt.Errorf("[UpdateHandler]: failed to save data: %w", err)
		log.Print(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
}

func (h *ControllerHandler) SyncHandler(c *gin.Context) {

	userID := c.GetInt("user_id")

	lastSyncStr := c.Param("lastsync")

	var lastSync time.Time
	var err error

	if lastSyncStr != "" {
		lastSync, err = h.parseTime(lastSyncStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "incorrect time format"})
			return
		}
	}

	data, err := h.service.GetDataForSync(c.Request.Context(), userID, lastSync)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to get data for sync: %v", err))
		return
	}

	if len(data) == 0 {
		c.Writer.WriteHeader(http.StatusNoContent)
		return
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	c.Writer.Header("Content-Type", "multipart/mixed; boundary="+mw.Boundary())

	for _, row := range data {
		// Создаем часть для файла
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, row.FileName))
		header.Set("Content-Type", "application/octet-stream")
		header.Set("meta", row.Metadata)
		header.Set("updated_at", row.UpdatedAt.Format(time.RFC3339Nano))
		header.Set("created_at", row.CreatedAt.Format(time.RFC3339Nano))
		header.Set("file_id", row.FileID.String())

		part, err := mw.CreatePart(header)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error creating part: %v", err))
			return
		}

		_, err = part.Write(row.Content)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Error writing file data: %v", err))
			return
		}
	}

	err = mw.Close()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error closing multipart writer: %v", err))
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write(buf.Bytes())

}

func (h *ControllerHandler) parseTime(str string) (t time.Time, err error) {

	if str == "" {
		return time.Now(), nil
	}
	t, err = time.Parse(time.RFC3339Nano, str)
	if err != nil {
		return
	}

	return
}
