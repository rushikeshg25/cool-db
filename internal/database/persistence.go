package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func loadState(path string) (databaseState, error) {
	empty := databaseState{Version: currentFormatVersion, Tables: make(map[string]*table)}
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return empty, nil
	}
	if err != nil {
		return databaseState{}, wrapError(CodeStorage, err, "could not open database %q", path)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var state databaseState
	if err := decoder.Decode(&state); err != nil {
		return databaseState{}, wrapError(CodeStorage, err, "could not decode database %q", path)
	}
	if err := ensureEOF(decoder); err != nil {
		return databaseState{}, wrapError(CodeStorage, err, "database %q contains trailing data", path)
	}
	if state.Version != currentFormatVersion {
		return databaseState{}, newError(CodeStorage, "database %q uses unsupported format version %d", path, state.Version)
	}
	if state.Tables == nil {
		state.Tables = make(map[string]*table)
	}
	if err := validateState(path, state); err != nil {
		return databaseState{}, err
	}
	return state, nil
}

// validateState rejects a snapshot the engine cannot execute against. The
// executor addresses cells positionally, so a row that is narrower than its
// table would panic on the first SELECT, UPDATE, or DELETE that reaches it.
func validateState(path string, state databaseState) error {
	for key, tbl := range state.Tables {
		if tbl == nil {
			return newError(CodeStorage, "database %q: table %q is missing its definition", path, key)
		}
		if tbl.Name != key {
			return newError(CodeStorage, "database %q: table %q is stored under key %q", path, tbl.Name, key)
		}
		if err := validateTable(path, tbl); err != nil {
			return err
		}
	}
	return nil
}

func validateTable(path string, tbl *table) error {
	if len(tbl.Columns) == 0 {
		return newError(CodeStorage, "database %q: table %q has no columns", path, tbl.Name)
	}
	seen := make(map[string]struct{}, len(tbl.Columns))
	primaryKeys := 0
	for _, column := range tbl.Columns {
		if column.Name == "" {
			return newError(CodeStorage, "database %q: table %q has an unnamed column", path, tbl.Name)
		}
		// findColumn compares against a lower-cased name, so a column stored
		// with any upper-case character can never be addressed by a query.
		if column.Name != strings.ToLower(column.Name) {
			return newError(CodeStorage, "database %q: column %q in table %q is not lower-case", path, column.Name, tbl.Name)
		}
		if _, exists := seen[column.Name]; exists {
			return newError(CodeStorage, "database %q: column %q is defined more than once in table %q", path, column.Name, tbl.Name)
		}
		seen[column.Name] = struct{}{}
		if !column.Type.valid() {
			return newError(CodeStorage, "database %q: column %q in table %q has unsupported type %q", path, column.Name, tbl.Name, column.Type)
		}
		if column.MaxLength < 0 {
			return newError(CodeStorage, "database %q: column %q in table %q has a negative length", path, column.Name, tbl.Name)
		}
		if column.PrimaryKey {
			primaryKeys++
		}
	}
	if primaryKeys > 1 {
		return newError(CodeStorage, "database %q: table %q declares %d PRIMARY KEY columns", path, tbl.Name, primaryKeys)
	}

	for rowIndex, row := range tbl.Rows {
		if len(row) != len(tbl.Columns) {
			return newError(CodeStorage, "database %q: row %d of table %q has %d value(s) but the table has %d column(s)", path, rowIndex, tbl.Name, len(row), len(tbl.Columns))
		}
		for columnIndex, column := range tbl.Columns {
			value := row[columnIndex]
			if value.Null {
				if column.NotNull || column.PrimaryKey {
					return newError(CodeStorage, "database %q: row %d of table %q holds NULL in non-nullable column %q", path, rowIndex, tbl.Name, column.Name)
				}
				// A NULL written by an older build may carry no type at all.
				row[columnIndex].Type = column.Type
				continue
			}
			if value.Type != column.Type {
				return newError(CodeStorage, "database %q: row %d of table %q stores a %q value in %q column %q", path, rowIndex, tbl.Name, value.Type, column.Type, column.Name)
			}
		}
	}
	return nil
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

func saveState(path string, state databaseState) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0700); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}
	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode database: %w", err)
	}
	payload = append(payload, '\n')

	temporary, err := os.CreateTemp(directory, ".cooldb-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary database: %w", err)
	}
	temporaryPath := temporary.Name()
	cleanup := func() {
		_ = temporary.Close()
		_ = os.Remove(temporaryPath)
	}
	if err := temporary.Chmod(0600); err != nil {
		cleanup()
		return fmt.Errorf("set database permissions: %w", err)
	}
	if _, err := temporary.Write(payload); err != nil {
		cleanup()
		return fmt.Errorf("write database: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync database: %w", err)
	}
	if err := temporary.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close database: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		cleanup()
		return fmt.Errorf("replace database: %w", err)
	}

	directoryFile, err := os.Open(directory)
	if err != nil {
		return fmt.Errorf("open database directory: %w", err)
	}
	defer directoryFile.Close()
	if err := directoryFile.Sync(); err != nil {
		return fmt.Errorf("sync database directory: %w", err)
	}
	return nil
}
