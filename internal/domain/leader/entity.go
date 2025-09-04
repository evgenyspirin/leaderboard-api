package leader

type (
	Leader struct {
		Rank     int
		TalentID string
		Score    float64
	}
	Leaders []*Leader
)
