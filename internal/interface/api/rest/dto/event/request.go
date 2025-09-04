package event

import (
	"time"

	"github.com/google/uuid"
)

type Request struct {
	EventID   uuid.UUID `json:"event_id"`
	TalentID  string    `json:"talent_id"`
	RawMetric float64   `json:"raw_metric"`
	Skill     string    `json:"skill"`
	TS        time.Time `json:"ts"`
}
