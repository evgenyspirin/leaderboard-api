package event

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	EventID   uuid.UUID
	TalentID  string
	RawMetric float64
	Skill     string
	TS        time.Time
	Score     float64
}
