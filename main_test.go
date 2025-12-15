package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/harshakonda/heapcheck/guard"
)

func TestHandleListUsers(t *testing.T) {
	defer guard.VerifyNone(t)

	server := NewServer()
	server.store.Create("Alice", "alice@test.com")
	server.store.Create("Bob", "bob@test.com")

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	server.HandleListUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var users []User
	json.NewDecoder(w.Body).Decode(&users)

	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestHandleCreateUser(t *testing.T) {
	defer guard.VerifyNone(t)

	server := NewServer()

	body := bytes.NewBufferString(`{"name":"Test","email":"test@test.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	w := httptest.NewRecorder()

	server.HandleCreateUser(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	var user User
	json.NewDecoder(w.Body).Decode(&user)

	if user.Name != "Test" {
		t.Errorf("expected 'Test', got '%s'", user.Name)
	}
	if user.ID != 1 {
		t.Errorf("expected ID 1, got %d", user.ID)
	}
}

func TestHandleGetUser(t *testing.T) {
	defer guard.VerifyNone(t)

	server := NewServer()
	server.store.Create("Alice", "alice@test.com")

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	server.HandleGetUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var user User
	json.NewDecoder(w.Body).Decode(&user)

	if user.Name != "Alice" {
		t.Errorf("expected 'Alice', got '%s'", user.Name)
	}
}

func TestHandleGetUserNotFound(t *testing.T) {
	defer guard.VerifyNone(t)

	server := NewServer()

	req := httptest.NewRequest(http.MethodGet, "/users/999", nil)
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()

	server.HandleGetUser(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandleDeleteUser(t *testing.T) {
	defer guard.VerifyNone(t)

	server := NewServer()
	server.store.Create("Alice", "alice@test.com")

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	server.HandleDeleteUser(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}

	// Verify deleted
	_, ok := server.store.Get(1)
	if ok {
		t.Error("expected user to be deleted")
	}
}

func TestUserStoreConcurrent(t *testing.T) {
	defer guard.VerifyNone(t,
		guard.MaxGoroutines(10),
	)

	server := NewServer()

	// Concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			server.store.Create("User", "user@test.com")
			done <- true
		}(i)
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}

	users := server.store.List()
	if len(users) != 10 {
		t.Errorf("expected 10 users, got %d", len(users))
	}
}

func TestHealthCheck(t *testing.T) {
	defer guard.VerifyNone(t)

	server := NewServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "OK" {
		t.Errorf("expected 'OK', got '%s'", w.Body.String())
	}
}
