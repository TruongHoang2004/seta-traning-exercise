package entity

import (
	"time"

	"github.com/google/uuid"
)

type NoteShare struct {
	ID     uuid.UUID
	NoteID uuid.UUID
	UserID uuid.UUID

	AccessLevel AccessLevel

	CreatedAt time.Time
}
