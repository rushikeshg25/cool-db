package core

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rushikeshg25/cool-wire/wire"
	"github.com/rushikeshg25/coolDb/internal/database"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSendQueryExecutesAgainstDatabase(t *testing.T) {
	engine, err := database.Open(filepath.Join(t.TempDir(), "test.cooldb"))
	if err != nil {
		t.Fatalf("database.Open() error = %v", err)
	}
	service := &CoreServerGRPC{executor: engine}

	queries := []string{
		"CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)",
		"INSERT INTO users VALUES (1, 'Ada')",
	}
	for _, query := range queries {
		if _, err := service.SendQuery(context.Background(), &wire.Query{Query: query}); err != nil {
			t.Fatalf("SendQuery(%q) error = %v", query, err)
		}
	}

	response, err := service.SendQuery(context.Background(), &wire.Query{Query: "SELECT * FROM users"})
	if err != nil {
		t.Fatalf("SendQuery(SELECT) error = %v", err)
	}
	for _, expected := range []string{"id", "name", "1", "Ada"} {
		if !strings.Contains(response.GetResponse(), expected) {
			t.Errorf("response = %q, want it to contain %q", response.GetResponse(), expected)
		}
	}
}

func TestSendQueryMapsDatabaseErrorsToGRPC(t *testing.T) {
	engine, err := database.Open(filepath.Join(t.TempDir(), "test.cooldb"))
	if err != nil {
		t.Fatalf("database.Open() error = %v", err)
	}
	service := &CoreServerGRPC{executor: engine}

	tests := []struct {
		name  string
		query *wire.Query
		code  codes.Code
	}{
		{name: "empty", query: &wire.Query{}, code: codes.InvalidArgument},
		{name: "nil", query: nil, code: codes.InvalidArgument},
		{name: "syntax", query: &wire.Query{Query: "hello"}, code: codes.InvalidArgument},
		{name: "not found", query: &wire.Query{Query: "SELECT * FROM missing"}, code: codes.NotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.SendQuery(context.Background(), test.query)
			if got := status.Code(err); got != test.code {
				t.Errorf("status code = %s, want %s (error: %v)", got, test.code, err)
			}
		})
	}
}
