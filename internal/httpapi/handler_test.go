package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/rushikeshg25/coolDb/internal/database"
)

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	engine, err := database.Open(filepath.Join(t.TempDir(), "http-test.cooldb"))
	if err != nil {
		t.Fatalf("database.Open() error = %v", err)
	}
	return NewHandler(engine)
}

func TestHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	response := httptest.NewRecorder()
	newTestHandler(t).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Header().Get("X-CoolDB-Demo") != "true" {
		t.Error("missing demo response header")
	}
}

func TestQueryExecutesSQL(t *testing.T) {
	handler := newTestHandler(t)
	queries := []string{
		"CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)",
		"INSERT INTO users VALUES (1, 'Ada')",
	}
	for _, query := range queries {
		response := performQuery(t, handler, query)
		if response.Code != http.StatusOK {
			t.Fatalf("query %q status = %d, body = %s", query, response.Code, response.Body.String())
		}
	}
	response := performQuery(t, handler, "SELECT * FROM users")
	if response.Code != http.StatusOK {
		t.Fatalf("SELECT status = %d", response.Code)
	}
	var payload queryResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Output == "" || payload.Error != nil {
		t.Errorf("response = %#v", payload)
	}
}

func TestQueryReturnsStructuredErrors(t *testing.T) {
	response := performQuery(t, newTestHandler(t), "SELECT * FROM missing")
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	var payload queryResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error == nil || payload.Error.Code != string(database.CodeNotFound) {
		t.Errorf("response = %#v", payload)
	}
}

func TestQueryRejectsInvalidBody(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/query", bytes.NewBufferString(`{"query":"SELECT 1","extra":true}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	newTestHandler(t).ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func performQuery(t *testing.T, handler http.Handler, query string) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(queryRequest{Query: query})
	if err != nil {
		t.Fatalf("encode query: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/query", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
