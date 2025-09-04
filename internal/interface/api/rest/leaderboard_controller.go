package rest

import (
	"encoding/json"
	"leaderboard-api/internal/application/ports"
	"net/http"
	"strconv"
	"strings"
)

const (
	defaultLimit = 10
	maxLimit     = 100
)

type LeaderboardController struct {
	lbService ports.LeaderboardService
}

func NewLeaderboardController(
	m *http.ServeMux,
	leaderboard ports.LeaderboardService,
) *LeaderboardController {
	ec := &LeaderboardController{
		lbService: leaderboard,
	}

	m.HandleFunc(http.MethodGet+Space+RouteLeaderboard, ec.GetBboard)
	m.HandleFunc(http.MethodGet+Space+RouteRank+Slash, ec.GetRankByID)

	return ec
}

func (lc *LeaderboardController) GetBboard(w http.ResponseWriter, r *http.Request) {
	limit := defaultLimit
	if s := r.URL.Query().Get("limit"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v <= 0 || v > maxLimit {
			http.Error(w, "invalid limit (must be 1..100)", http.StatusBadRequest)
			return
		}
		limit = v
	}

	leaders, err := lc.lbService.GetBboard(r.Context(), limit)
	if err != nil {
		http.Error(w, "failed to get a leaderboard", http.StatusInternalServerError)
		return
	}

	w.Header().Set(HeaderContentType, ContentTypeJSON)
	if err := json.NewEncoder(w).Encode(leaders); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}

func (lc *LeaderboardController) GetRankByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, RouteRank+Slash) {
		http.NotFound(w, r)
		return
	}
	id := strings.TrimPrefix(path, RouteRank+Slash)
	if id == "" || strings.Contains(id, Slash) {
		http.Error(w, "Invalid or missing ID", http.StatusBadRequest)
		return
	}

	leader, err := lc.lbService.GetRankByID(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to get a rank", http.StatusInternalServerError)
		return
	}
	if leader.Rank == 0 {
		http.Error(w, "leader not found", http.StatusNotFound)
		return
	}

	w.Header().Set(HeaderContentType, ContentTypeJSON)
	if err := json.NewEncoder(w).Encode(leader); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}
