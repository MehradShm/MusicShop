package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	repo "musicshop/backend/internal/repository"
)

type Server struct {
	Repo repo.UserRepository
}

func (s *Server) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/users", s.handleUsers)
	mux.HandleFunc("/users/", s.handleUserByID)
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		users, err := s.Repo.List(r.Context())
		respondJSON(w, users, err)
	case http.MethodPost:
		var in struct{ Username, Name, Email string }
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		u := repo.User{Username: in.Username, Name: in.Name, Email: in.Email}
		if err := s.Repo.Create(r.Context(), &u); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		respondJSON(w, u, nil)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/users/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "bad id", 400)
		return
	}

	switch r.Method {
	case http.MethodGet:
		u, err := s.Repo.Get(r.Context(), id)
		respondJSON(w, u, err)
	case http.MethodPut:
		var in struct{ Username, Name, Email string }
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		u := repo.User{Username: in.Username, Name: in.Name, Email: in.Email}
		respondJSON(w, map[string]string{"status": "updated"}, s.Repo.Update(r.Context(), id, &u))
	case http.MethodDelete:
		respondJSON(w, map[string]string{"status": "deleted"}, s.Repo.Delete(r.Context(), id))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func respondJSON(w http.ResponseWriter, v any, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

// Optional context helper
func withTimeout(ctx context.Context) (context.Context, func()) {
	return context.WithTimeout(ctx, 5_000_000_000) // 5s
}
