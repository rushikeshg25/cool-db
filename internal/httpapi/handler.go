package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rushikeshg25/coolDb/internal/database"
)

const maxQueryBody = 64 << 10

type queryExecutor interface {
	Execute(string) (database.Result, error)
}

type Handler struct {
	executor queryExecutor
}

func NewHandler(executor queryExecutor) http.Handler {
	handler := &Handler{executor: executor}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handler.health)
	mux.HandleFunc("POST /api/query", handler.query)
	return securityHeaders(mux)
}

type queryRequest struct {
	Query string `json:"query"`
}

type queryResponse struct {
	Output string        `json:"output,omitempty"`
	Error  *errorPayload `json:"error,omitempty"`
}

type errorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (h *Handler) health(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) query(writer http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(writer, request.Body, maxQueryBody)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	var payload queryRequest
	if err := decoder.Decode(&payload); err != nil {
		writeError(writer, http.StatusBadRequest, "INVALID_REQUEST", "request body must contain a JSON query string")
		return
	}
	if err := ensureEOF(decoder); err != nil {
		writeError(writer, http.StatusBadRequest, "INVALID_REQUEST", "request body must contain one JSON object")
		return
	}
	payload.Query = strings.TrimSpace(payload.Query)
	if payload.Query == "" {
		writeError(writer, http.StatusBadRequest, "INVALID_REQUEST", "query cannot be empty")
		return
	}

	result, err := h.executor.Execute(payload.Query)
	if err != nil {
		writeDatabaseError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, queryResponse{Output: result.Format()})
}

func ensureEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("unexpected JSON value")
		}
		return err
	}
	return nil
}

func writeDatabaseError(writer http.ResponseWriter, err error) {
	var databaseError *database.Error
	if !errors.As(err, &databaseError) {
		writeError(writer, http.StatusInternalServerError, "INTERNAL", "internal database error")
		return
	}
	status := http.StatusInternalServerError
	switch databaseError.Code {
	case database.CodeSyntax, database.CodeType:
		status = http.StatusBadRequest
	case database.CodeAlreadyExists:
		status = http.StatusConflict
	case database.CodeNotFound:
		status = http.StatusNotFound
	case database.CodeConstraint:
		status = http.StatusUnprocessableEntity
	}
	writeError(writer, status, string(databaseError.Code), databaseError.Message)
}

func writeError(writer http.ResponseWriter, status int, code, message string) {
	writeJSON(writer, status, queryResponse{Error: &errorPayload{Code: code, Message: message}})
}

func writeJSON(writer http.ResponseWriter, status int, payload any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(payload)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Cache-Control", "no-store")
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		writer.Header().Set("X-CoolDB-Demo", "true")
		next.ServeHTTP(writer, request)
	})
}
