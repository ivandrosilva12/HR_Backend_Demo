package entities

import (
	"rhapp/internal/domain/vos"
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID         uuid.UUID
	OwnerType  vos.DocumentOwnerType // "employee", "dependent", etc.
	OwnerID    uuid.UUID             // ID do proprietário (Employee ou Dependent)
	Type       vos.DocumentType      // BI, Contrato, Diploma, etc.
	FileName   vos.Filename
	FileURL    vos.DocumentURL
	Extension  vos.FileExtension // jpg, pdf, png, etc.
	IsActive   bool
	UploadedAt time.Time
	ObjectKey  string
}
