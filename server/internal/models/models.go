package models

import "github.com/google/uuid"

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// type UploadRequest struct {
// 	Type     string            `json:"type"`
// 	Metadata string 		   `json:"metadata"`
// 	Data     []byte            `json:"data"` // Для передачи бинарных данных используем base64 в JSON, но лучше - multipart/form-data
// }

type SyncData struct {
	FileID   uuid.UUID
	FileName string
	Metadata string
	Content  []byte
}
