package database

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestEngineCRUDAndPersistence(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "test.cooldb")
	engine, err := Open(databasePath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	queries := []string{
		"CREATE TABLE users (id INTEGER PRIMARY KEY, name VARCHAR(20) NOT NULL, active BOOLEAN, score FLOAT)",
		"INSERT INTO users VALUES (1, 'Ada', true, 9.5)",
		"INSERT INTO users (id, name, active, score) VALUES (2, 'Grace', false, 8)",
		"UPDATE users SET active = true, score = 8.75 WHERE id = 2",
	}
	for _, query := range queries {
		if _, err := engine.Execute(query); err != nil {
			t.Fatalf("Execute(%q) error = %v", query, err)
		}
	}

	result, err := engine.Execute("SELECT name, score FROM users WHERE active = true;")
	if err != nil {
		t.Fatalf("SELECT error = %v", err)
	}
	if got, want := len(result.Rows), 2; got != want {
		t.Fatalf("row count = %d, want %d", got, want)
	}
	if got, want := result.Rows[1][0].String(), "Grace"; got != want {
		t.Errorf("selected name = %q, want %q", got, want)
	}
	if got, want := result.Rows[1][1].String(), "8.75"; got != want {
		t.Errorf("selected score = %q, want %q", got, want)
	}

	reopened, err := Open(databasePath)
	if err != nil {
		t.Fatalf("reopen error = %v", err)
	}
	result, err = reopened.Execute("SELECT * FROM users WHERE id = 1")
	if err != nil {
		t.Fatalf("SELECT after reopen error = %v", err)
	}
	if got, want := len(result.Rows), 1; got != want {
		t.Fatalf("persisted row count = %d, want %d", got, want)
	}

	result, err = reopened.Execute("DELETE FROM users WHERE id = 1")
	if err != nil {
		t.Fatalf("DELETE error = %v", err)
	}
	if got, want := result.AffectedRows, 1; got != want {
		t.Errorf("deleted rows = %d, want %d", got, want)
	}
}

func TestEngineValidatesQueriesAndConstraints(t *testing.T) {
	engine, err := Open(filepath.Join(t.TempDir(), "test.cooldb"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, err := engine.Execute("CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT UNIQUE, name TEXT NOT NULL)"); err != nil {
		t.Fatalf("CREATE TABLE error = %v", err)
	}
	if _, err := engine.Execute("INSERT INTO users VALUES (1, 'one@example.com', 'One')"); err != nil {
		t.Fatalf("INSERT error = %v", err)
	}

	tests := []struct {
		name string
		sql  string
		code ErrorCode
	}{
		{name: "duplicate primary key", sql: "INSERT INTO users VALUES (1, 'two@example.com', 'Two')", code: CodeConstraint},
		{name: "duplicate unique value", sql: "INSERT INTO users VALUES (2, 'one@example.com', 'Two')", code: CodeConstraint},
		{name: "missing required value", sql: "INSERT INTO users (id, email) VALUES (2, 'two@example.com')", code: CodeConstraint},
		{name: "wrong type", sql: "INSERT INTO users VALUES ('two', 'two@example.com', 'Two')", code: CodeType},
		{name: "unknown table", sql: "SELECT * FROM missing", code: CodeNotFound},
		{name: "invalid syntax", sql: "SELECT FROM users", code: CodeSyntax},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := engine.Execute(test.sql)
			var databaseError *Error
			if !errors.As(err, &databaseError) {
				t.Fatalf("error = %v, want *database.Error", err)
			}
			if databaseError.Code != test.code {
				t.Errorf("error code = %s, want %s", databaseError.Code, test.code)
			}
		})
	}
}

func TestFailedPersistenceDoesNotMutateEngine(t *testing.T) {
	engine, err := Open(filepath.Join(t.TempDir(), "test.cooldb"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	engine.persist = func(databaseState) error { return errors.New("disk full") }

	if _, err := engine.Execute("CREATE TABLE users (id INTEGER)"); err == nil {
		t.Fatal("CREATE TABLE error = nil, want storage error")
	}
	if _, err := engine.Execute("SELECT * FROM users"); err == nil {
		t.Fatal("SELECT found table after failed persistence")
	}
}

func TestResultFormat(t *testing.T) {
	result := Result{
		Columns: []string{"id", "name"},
		Rows: [][]Value{{
			{Type: TypeInteger, Integer: 1},
			{Type: TypeText, Text: "Ada"},
		}},
	}
	formatted := result.Format()
	for _, expected := range []string{"id", "name", "1", "Ada", "(1 row(s))"} {
		if !strings.Contains(formatted, expected) {
			t.Errorf("Format() = %q, want it to contain %q", formatted, expected)
		}
	}
	if got, want := (Result{AffectedRows: 2}).Format(), "2 row(s) affected"; got != want {
		t.Errorf("mutation Format() = %q, want %q", got, want)
	}
}

func TestNullComparisonsFollowThreeValuedLogic(t *testing.T) {
	engine, err := Open(filepath.Join(t.TempDir(), "nulls.cooldb"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	setup := []string{
		"CREATE TABLE people (id INTEGER PRIMARY KEY, name TEXT)",
		"INSERT INTO people VALUES (1, 'Ada')",
		"INSERT INTO people VALUES (2, NULL)",
	}
	for _, query := range setup {
		if _, err := engine.Execute(query); err != nil {
			t.Fatalf("Execute(%q) error = %v", query, err)
		}
	}

	result, err := engine.Execute("SELECT id FROM people WHERE name = NULL")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := len(result.Rows); got != 0 {
		t.Errorf("`= NULL` matched %d row(s), want 0", got)
	}

	result, err = engine.Execute("SELECT id FROM people WHERE name = 'Ada'")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := len(result.Rows); got != 1 {
		t.Errorf("`= 'Ada'` matched %d row(s), want 1", got)
	}
}

// UNIQUE ignores NULLs, and must keep ignoring them now that valuesEqual
// reports NULL comparisons as unequal.
func TestUniqueColumnAcceptsRepeatedNulls(t *testing.T) {
	engine, err := Open(filepath.Join(t.TempDir(), "unique.cooldb"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	setup := []string{
		"CREATE TABLE accounts (id INTEGER PRIMARY KEY, email TEXT UNIQUE)",
		"INSERT INTO accounts VALUES (1, NULL)",
		"INSERT INTO accounts VALUES (2, NULL)",
	}
	for _, query := range setup {
		if _, err := engine.Execute(query); err != nil {
			t.Fatalf("Execute(%q) error = %v", query, err)
		}
	}
	if _, err := engine.Execute("INSERT INTO accounts VALUES (3, 'a@example.com')"); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if _, err := engine.Execute("INSERT INTO accounts VALUES (4, 'a@example.com')"); err == nil {
		t.Fatal("duplicate non-NULL value was accepted, want a constraint error")
	}
}
