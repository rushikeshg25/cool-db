package database

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSnapshot(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.cooldb")
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}
	return path
}

func TestOpenRejectsCorruptSnapshots(t *testing.T) {
	cases := []struct {
		name     string
		contents string
		wantIn   string
	}{
		{
			name:     "short row",
			contents: `{"version":1,"tables":{"t":{"name":"t","columns":[{"name":"id","type":"INTEGER"},{"name":"note","type":"TEXT"}],"rows":[[{"type":"INTEGER","integer":1}]]}}}`,
			wantIn:   "has 1 value(s) but the table has 2 column(s)",
		},
		{
			name:     "long row",
			contents: `{"version":1,"tables":{"t":{"name":"t","columns":[{"name":"id","type":"INTEGER"}],"rows":[[{"type":"INTEGER","integer":1},{"type":"TEXT","text":"x"}]]}}}`,
			wantIn:   "has 2 value(s) but the table has 1 column(s)",
		},
		{
			name:     "unknown column type",
			contents: `{"version":1,"tables":{"t":{"name":"t","columns":[{"name":"payload","type":"BLOB"}],"rows":[]}}}`,
			wantIn:   `unsupported type "BLOB"`,
		},
		{
			name:     "cell type does not match column",
			contents: `{"version":1,"tables":{"t":{"name":"t","columns":[{"name":"id","type":"INTEGER"}],"rows":[[{"type":"TEXT","text":"one"}]]}}}`,
			wantIn:   `stores a "TEXT" value in "INTEGER" column`,
		},
		{
			name:     "no columns",
			contents: `{"version":1,"tables":{"t":{"name":"t","columns":[],"rows":[]}}}`,
			wantIn:   "has no columns",
		},
		{
			name:     "duplicate column",
			contents: `{"version":1,"tables":{"t":{"name":"t","columns":[{"name":"id","type":"INTEGER"},{"name":"id","type":"TEXT"}],"rows":[]}}}`,
			wantIn:   "defined more than once",
		},
		{
			name:     "unnamed column",
			contents: `{"version":1,"tables":{"t":{"name":"t","columns":[{"name":"","type":"INTEGER"}],"rows":[]}}}`,
			wantIn:   "unnamed column",
		},
		{
			name:     "upper-case column name is unaddressable",
			contents: `{"version":1,"tables":{"t":{"name":"t","columns":[{"name":"ID","type":"INTEGER"}],"rows":[]}}}`,
			wantIn:   "is not lower-case",
		},
		{
			name:     "null in a not-null column",
			contents: `{"version":1,"tables":{"t":{"name":"t","columns":[{"name":"id","type":"INTEGER","not_null":true}],"rows":[[{"type":"INTEGER","null":true}]]}}}`,
			wantIn:   "holds NULL in non-nullable column",
		},
		{
			name:     "table stored under a mismatched key",
			contents: `{"version":1,"tables":{"other":{"name":"t","columns":[{"name":"id","type":"INTEGER"}],"rows":[]}}}`,
			wantIn:   `is stored under key "other"`,
		},
		{
			name:     "two primary keys",
			contents: `{"version":1,"tables":{"t":{"name":"t","columns":[{"name":"a","type":"INTEGER","primary_key":true},{"name":"b","type":"INTEGER","primary_key":true}],"rows":[]}}}`,
			wantIn:   "declares 2 PRIMARY KEY columns",
		},
		{
			name:     "negative varchar length",
			contents: `{"version":1,"tables":{"t":{"name":"t","columns":[{"name":"a","type":"TEXT","max_length":-1}],"rows":[]}}}`,
			wantIn:   "negative length",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			engine, err := Open(writeSnapshot(t, testCase.contents))
			if err == nil {
				t.Fatalf("Open() succeeded, want a storage error (engine = %v)", engine)
			}
			var databaseError *Error
			if !errors.As(err, &databaseError) {
				t.Fatalf("Open() error = %T (%v), want *database.Error", err, err)
			}
			if databaseError.Code != CodeStorage {
				t.Errorf("error code = %q, want %q", databaseError.Code, CodeStorage)
			}
			if !strings.Contains(databaseError.Message, testCase.wantIn) {
				t.Errorf("error message = %q, want it to contain %q", databaseError.Message, testCase.wantIn)
			}
		})
	}
}

// A corrupt snapshot used to panic here rather than fail at Open, because the
// executor addresses cells positionally.
func TestQueryingAShortRowDoesNotPanic(t *testing.T) {
	path := writeSnapshot(t, `{"version":1,"tables":{"t":{"name":"t","columns":[{"name":"id","type":"INTEGER"},{"name":"note","type":"TEXT"}],"rows":[[{"type":"INTEGER","integer":1}]]}}}`)
	engine, err := Open(path)
	if err == nil {
		if _, err := engine.Execute("SELECT * FROM t"); err == nil {
			t.Fatal("SELECT against a short row succeeded, want a failure at Open or query time")
		}
	}
}

func TestOpenAcceptsAValidSnapshot(t *testing.T) {
	path := writeSnapshot(t, `{"version":1,"tables":{"t":{"name":"t","columns":[{"name":"id","type":"INTEGER","primary_key":true},{"name":"note","type":"TEXT"}],"rows":[[{"type":"INTEGER","integer":1},{"type":"TEXT","null":true}]]}}}`)
	engine, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	result, err := engine.Execute("SELECT id, note FROM t")
	if err != nil {
		t.Fatalf("SELECT error = %v", err)
	}
	if got, want := len(result.Rows), 1; got != want {
		t.Fatalf("row count = %d, want %d", got, want)
	}
	if got, want := result.Rows[0][1].String(), "NULL"; got != want {
		t.Errorf("note = %q, want %q", got, want)
	}
}

// A NULL written without a type must adopt its column's type on load so that
// it still renders and compares correctly.
func TestOpenNormalizesUntypedNulls(t *testing.T) {
	path := writeSnapshot(t, `{"version":1,"tables":{"t":{"name":"t","columns":[{"name":"note","type":"TEXT"}],"rows":[[{"null":true}]]}}}`)
	engine, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	result, err := engine.Execute("SELECT note FROM t")
	if err != nil {
		t.Fatalf("SELECT error = %v", err)
	}
	if got, want := result.Rows[0][0].Type, TypeText; got != want {
		t.Errorf("normalized type = %q, want %q", got, want)
	}
}
