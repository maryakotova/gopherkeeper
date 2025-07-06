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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"file_id": id})
}

func (h *ControllerHandler) DownloadHandler(c *gin.Context) {

	fileIDStr := c.Param("id")

	if fileIDStr == "" {
		c.String(http.StatusBadRequest, "ID файла не указан в запросе")
		return
	}

	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	userID := c.GetInt("user_id")

	data, fileName, meta, updatedAt, err := h.service.Download(c.Request.Context(), fileID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	defer mw.Close()

	fw, err := mw.CreateFormFile("file", fileName)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create form file: %v", err)
		return
	}
	_, err = fw.Write(data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to write file data: %v", err)
		return
	}

	err = mw.WriteField("meta", meta)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to write meta field: %v", err)
		return
	}
	err = mw.WriteField("updated_at", updatedAt.String())
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to write date field: %v", err)
		return
	}

	c.Header("Content-Type", mw.FormDataContentType())
	c.Header("Content-Length", fmt.Sprint(buf.Len()))
	c.Writer.WriteHeader(http.StatusOK)
	_, err = c.Writer.Write(buf.Bytes())
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to write response: %v", err)
		return
	}

}

func (h *ControllerHandler) UpdateHandler(c *gin.Context) {

	// fileName := c.GetHeader("X-Filename")
	// fileIDStr := c.GetHeader("X-File-ID")
	// metadata := c.GetHeader("X-Meta")
	// userID := c.GetInt("user_id")

	// var updatedAt time.Time
	// updatedAtStr := c.GetHeader("X-Updated-At")
	// updatedAt, err := h.parseTime(updatedAtStr)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "incorrect time format"})
	// 	return
	// }

	// if fileName == "" {
	// 	c.String(http.StatusBadRequest, "X-Filename header is required")
	// 	return
	// }

	// if fileIDStr == "" {
	// 	c.String(http.StatusBadRequest, "X-File-ID header is required")
	// 	return
	// }

	// fileID, err := uuid.Parse(fileIDStr)
	// if err != nil {
	// 	log.Print(err)
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err})
	// 	return
	// }

	// data, err := io.ReadAll(c.Request.Body)
	// if err != nil {
	// 	c.String(http.StatusInternalServerError, "Failed to read body: %v", err)
	// 	return
	// }

	// err = h.service.Update(c.Request.Context(), fileID, userID, data, fileName, metadata, updatedAt)
	// if err != nil {
	// 	c.String(http.StatusInternalServerError, "Failed to update data: %v", err)
	// 	return
	// }

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
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
		c.String(http.StatusInternalServerError, "Failed to get data for sync: %v", err)
		return
	}

	if data == nil {
		c.Writer.WriteHeader(http.StatusNoContent)
		return
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	c.Header("Content-Type", "multipart/mixed; boundary="+mw.Boundary())

	for _, row := range data {
		// Создаем часть для файла
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, row.FileName))
		header.Set("Content-Type", "application/octet-stream")
		header.Set("X-Meta", row.Metadata)

		part, err := mw.CreatePart(header)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error creating part: %v", err)
			return
		}

		_, err = part.Write(row.Content)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error writing file data: %v", err)
			return
		}
	}

	err = mw.Close()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error closing multipart writer: %v", err)
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
