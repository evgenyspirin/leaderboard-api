package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"leaderboard-api/internal/application/ports"
	"leaderboard-api/internal/interface/api/rest/dto/event"
)

type EventsController struct {
	eventService ports.EventService
}

func NewEventController(m *http.ServeMux, eventService ports.EventService) *EventsController {
	ec := &EventsController{
		eventService: eventService,
	}

	m.HandleFunc(http.MethodPost+Space+RouteEvents, ec.PostEventHandler)
	m.HandleFunc(http.MethodGet+Space+RouteSeed, ec.SeedHandler)

	return ec
}

func (ec *EventsController) PostEventHandler(w http.ResponseWriter, r *http.Request) {
	var req event.Request
	// for big performance and to avoid reflection under the hood
	// better to use codegen for marshal/unmarshal for example "easyjson" pkg
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json request", http.StatusBadRequest)
		return
	}
	// example
	//if err := validate(req); err != nil {
	//	http.Error(w, "validation error", http.StatusBadRequest)
	//	return
	//}

	duplicate, err := ec.eventService.Create(r.Context(), event.FromRequest(req))
	if err != nil {
		http.Error(w, "failed to create an event", http.StatusInternalServerError)
		return
	}
	if !duplicate {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// SeedHandler - En extra endpoint to seed real N events
func (ec *EventsController) SeedHandler(w http.ResponseWriter, r *http.Request) {
	cnt := defaultLimit
	s := r.URL.Query().Get("count")
	if s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v <= 0 {
			http.Error(w, "invalid count (must be > 0)", http.StatusBadRequest)
			return
		}
		cnt = v
	}

	ec.eventService.Seed(r.Context(), cnt)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Success string `json:"Seed"`
	}{Success: "Database seeded successfully with " + s + " records"})
}
