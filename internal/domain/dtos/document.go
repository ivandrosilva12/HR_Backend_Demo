package dtos

import (
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/vos"
	"strings"
	"time"
)

type UpdateDocumentDTO struct {
	Name      string `json:"name" binding:"omitempty,min=3,max=100"`
	FileURL   string `json:"file_url" binding:"omitempty,url"`
	Extension string `json:"file_ext" binding:"omitempty,oneof=pdf jpg jpeg png doc docx"`
	ObjectKey string `json:"object_key" binding:"omitempty"`
}

type DocumentResponseDTO struct {
	ID         string    `json:"id"`
	OwnerType  string    `json:"owner_type"`
	OwnerID    string    `json:"owner_id"`
	Type       string    `json:"type"`
	FileName   string    `json:"name"`
	FileURL    string    `json:"file_url"`
	IsActive   bool      `json:"is_active"`
	UploadedAt time.Time `json:"uploaded_at"`
	ObjectKey  string    `json:"object_key"`
}

type UploadDocumentForm struct {
	OwnerType string `form:"owner_type" binding:"required"`
	OwnerID   string `form:"owner_id" binding:"required,uuid4"`
	Type      string `form:"type" binding:"required"`
}

type ListDocumentsByOwnerDTO struct {
	OwnerID string `form:"owner_id" binding:"required,uuid4"` // ← era json:"owner_id"
}

func ToDocumentResponseDTO(d entities.Document) DocumentResponseDTO {
	return DocumentResponseDTO{
		ID:         d.ID.String(),
		OwnerType:  d.OwnerType.String(),
		OwnerID:    d.OwnerID.String(),
		Type:       d.Type.String(),
		FileName:   d.FileName.String(),
		FileURL:    d.FileURL.String(),
		IsActive:   d.IsActive,
		UploadedAt: d.UploadedAt,
		ObjectKey:  d.ObjectKey,
	}
}

func ApplyUpdateToDocument(doc *entities.Document, input UpdateDocumentDTO) error {
	if input.Name != "" {
		docName, err := vos.NewFilename(input.Name)
		if err != nil {
			return err
		}
		doc.FileName = docName
	}

	if input.FileURL != "" {
		u, err := vos.NewDocumentURL(input.FileURL)
		if err != nil {
			return err
		}
		doc.FileURL = u
	}

	if input.Extension != "" {
		ext, err := vos.NewFileExtension(input.Extension)
		if err != nil {
			return err
		}
		doc.Extension = ext
	}
	if input.ObjectKey != "" {
		doc.ObjectKey = strings.TrimSpace(input.ObjectKey)
	}

	doc.UploadedAt = time.Now()
	return nil
}

func ToDocumentResponseDTOList(list []entities.Document) []DocumentResponseDTO {
	result := make([]DocumentResponseDTO, len(list))
	for i, d := range list {
		result[i] = ToDocumentResponseDTO(d)
	}
	return result
}
