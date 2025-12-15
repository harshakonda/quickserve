// Package main demonstrates a simple REST API with heapcheck integration.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// User represents a user in the system
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserStore is an in-memory user store
type UserStore struct {
	mu    sync.RWMutex
	users map[int]User
	next  int
}

// NewUserStore creates a new user store
func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[int]User),
		next:  1,
	}
}

// Create adds a new user
func (s *UserStore) Create(name, email string) User {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := User{
		ID:    s.next,
		Name:  name,
		Email: email,
	}
	s.users[s.next] = user
	s.next++
	return user
}

// Get retrieves a user by ID
func (s *UserStore) Get(id int) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[id]
	return user, ok
}

// List returns all users
func (s *UserStore) List() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}
	return users
}

// Delete removes a user
func (s *UserStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[id]; ok {
		delete(s.users, id)
		return true
	}
	return false
}

// Server holds the HTTP server dependencies
type Server struct {
	store *UserStore
}

// NewServer creates a new server
func NewServer() *Server {
	return &Server{
		store: NewUserStore(),
	}
}

// HandleListUsers handles GET /users
func (s *Server) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	users := s.store.List()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// HandleGetUser handles GET /users/{id}
func (s *Server) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	user, ok := s.store.Get(id)
	if !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// HandleCreateUser handles POST /users
func (s *Server) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user := s.store.Create(req.Name, req.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// HandleDeleteUser handles DELETE /users/{id}
func (s *Server) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if !s.store.Delete(id) {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Routes returns the HTTP handler with all routes
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /users", s.HandleListUsers)
	mux.HandleFunc("GET /users/{id}", s.HandleGetUser)
	mux.HandleFunc("POST /users", s.HandleCreateUser)
	mux.HandleFunc("DELETE /users/{id}", s.HandleDeleteUser)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	return mux
}

func main() {
	server := NewServer()

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", server.Routes()); err != nil {
		log.Fatal(err)
	}
}
