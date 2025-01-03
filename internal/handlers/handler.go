package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/karthikeyaspace/game-leaderboard/internal/services"
)

type Handler struct {
	service services.Service
}

func NewHandler(service services.Service) *Handler {
	return &Handler{service: service}
}

type User struct {
	Name  string `json:"name"`
	ID    string `json:"userId"`
	Score int    `json:"score"`
}

// Create a new player - { name }
func (h *Handler) CreatePlayerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, "Bad Request", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	userId, err := h.service.CreatePlayerService(req.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to create player",
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user":    User{ID: userId, Name: req.Name, Score: 0},
	})

}

// Update the score of the player - {userId, newscore}
func (h *Handler) UpdateScoreHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID    string `json:"userId"`
		Score int    `json:"score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" || req.Score < 0 {
		http.Error(w, "Bad Request", http.StatusBadGateway)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := h.service.UpdateScoreService(req.ID, req.Score); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to update score",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"success": true})
}

// Get the leaderboard - /leaderboard?limit=n
func (h *Handler) GetLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	lim := r.URL.Query().Get("limit")
	if lim == "" {
		lim = "10"
	}

	w.Header().Set("Content-Type", "application/json")

	users, err := h.service.GetLeaderboardService(lim)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to get leaderboard",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"leaderboard": users,
	})
}

// realtime time event stream using sse
func (h *Handler) GetLeaderboardHandlerStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	clientDisconnect := r.Context().Done()
	flusher := w.(http.Flusher)

	for {
		select {
		case <-clientDisconnect:
			return
		default:
			users, err := h.service.GetLeaderboardService("10")
			if err != nil {
				return
			}

			data := map[string]interface{}{
				"success":     true,
				"leaderboard": users,
			}

			jsonData, err := json.Marshal(data)
			if err != nil {
				return
			}

			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			flusher.Flush()
			time.Sleep(2 * time.Second)
		}
	}
}

func (h *Handler) HomeHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
