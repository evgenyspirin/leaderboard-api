package event

import (
	"leaderboard-api/internal/domain/event"
)

func FromRequest(r Request) *event.Event {
	return &event.Event{
		EventID:   r.EventID,
		TalentID:  r.TalentID,
		RawMetric: r.RawMetric,
		Skill:     r.Skill,
		TS:        r.TS,
	}
}
