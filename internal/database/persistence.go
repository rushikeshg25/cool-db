package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	return state, nil
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
