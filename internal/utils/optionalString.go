package utils

import "github.com/google/uuid"

func ToOptionalString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}
