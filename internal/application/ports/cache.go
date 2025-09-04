package ports

import (
	"github.com/google/uuid"
)

type Cache interface {
	Set(eventID uuid.UUID)
	IsSet(id uuid.UUID) bool
}
