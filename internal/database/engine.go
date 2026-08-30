package database

import (
	"fmt"
	"strings"
	"sync"
)

const currentFormatVersion = 1

type table struct {
	Name    string    `json:"name"`
	Columns []Column  `json:"columns"`
	Rows    [][]Value `json:"rows"`
}

type databaseState struct {
	Version int               `json:"version"`
	Tables  map[string]*table `json:"tables"`
}

// Engine owns a single persistent database and executes the supported v0.1
// SQL subset against it.
type Engine struct {
	mu      sync.RWMutex
	path    string
	state   databaseState
	persist func(databaseState) error
}

// Open loads a database file or initializes an empty database when it does not
// exist. The first successful mutation creates the file.
func Open(path string) (*Engine, error) {
	state, err := loadState(path)
	if err != nil {
		return nil, err
	}
	engine := &Engine{path: path, state: state}
	engine.persist = func(state databaseState) error { return saveState(path, state) }
	return engine, nil
}

// Execute parses and runs exactly one SQL statement.
func (e *Engine) Execute(query string) (Result, error) {
	stmt, err := parseSQL(strings.TrimSpace(query))
	if err != nil {
		return Result{}, err
	}

	if selectStmt, ok := stmt.(selectStatement); ok {
		e.mu.RLock()
		defer e.mu.RUnlock()
		return executeSelect(e.state, selectStmt)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	next := cloneState(e.state)
	result, err := executeMutation(&next, stmt)
	if err != nil {
		return Result{}, err
	}
	if err := e.persist(next); err != nil {
		return Result{}, wrapError(CodeStorage, err, "could not persist database")
	}
	e.state = next
	return result, nil
}

func executeMutation(state *databaseState, stmt statement) (Result, error) {
	switch typed := stmt.(type) {
	case createTableStatement:
		return executeCreateTable(state, typed)
	case dropTableStatement:
		return executeDropTable(state, typed)
	case insertStatement:
		return executeInsert(state, typed)
	case updateStatement:
		return executeUpdate(state, typed)
	case deleteStatement:
		return executeDelete(state, typed)
	default:
		return Result{}, newError(CodeSyntax, "statement cannot mutate the database")
	}
}

func executeCreateTable(state *databaseState, stmt createTableStatement) (Result, error) {
	if _, exists := state.Tables[stmt.table]; exists {
		return Result{}, newError(CodeAlreadyExists, "table %q already exists", stmt.table)
	}
	seen := make(map[string]struct{}, len(stmt.columns))
	primaryKeys := 0
	for _, column := range stmt.columns {
		if _, exists := seen[column.Name]; exists {
			return Result{}, newError(CodeAlreadyExists, "column %q is defined more than once", column.Name)
		}
		seen[column.Name] = struct{}{}
		if column.PrimaryKey {
			primaryKeys++
		}
	}
	if primaryKeys > 1 {
		return Result{}, newError(CodeConstraint, "v0.1 supports one PRIMARY KEY column per table")
	}

	state.Tables[stmt.table] = &table{Name: stmt.table, Columns: stmt.columns, Rows: make([][]Value, 0)}
	return Result{Message: fmt.Sprintf("table %q created", stmt.table)}, nil
}

func executeDropTable(state *databaseState, stmt dropTableStatement) (Result, error) {
	if _, exists := state.Tables[stmt.table]; !exists {
		return Result{}, newError(CodeNotFound, "table %q does not exist", stmt.table)
	}
	delete(state.Tables, stmt.table)
	return Result{Message: fmt.Sprintf("table %q dropped", stmt.table)}, nil
}

func executeInsert(state *databaseState, stmt insertStatement) (Result, error) {
	tbl, err := findTable(*state, stmt.table)
	if err != nil {
		return Result{}, err
	}
	columns := stmt.columns
	if len(columns) == 0 {
		columns = make([]string, len(tbl.Columns))
		for index, column := range tbl.Columns {
			columns[index] = column.Name
		}
	}
	if len(columns) != len(stmt.values) {
		return Result{}, newError(CodeConstraint, "INSERT lists %d column(s) but provides %d value(s)", len(columns), len(stmt.values))
	}

	row := make([]Value, len(tbl.Columns))
	for index, column := range tbl.Columns {
		row[index] = Value{Type: column.Type, Null: true}
	}
	assigned := make(map[int]struct{}, len(columns))
	for index, name := range columns {
		columnIndex, column, err := findColumn(tbl, name)
		if err != nil {
			return Result{}, err
		}
		if _, exists := assigned[columnIndex]; exists {
			return Result{}, newError(CodeAlreadyExists, "column %q is assigned more than once", name)
		}
		value, err := coerceLiteral(stmt.values[index], column)
		if err != nil {
			return Result{}, err
		}
		row[columnIndex] = value
		assigned[columnIndex] = struct{}{}
	}
	if err := validateRow(tbl, row, -1); err != nil {
		return Result{}, err
	}
	tbl.Rows = append(tbl.Rows, row)
	return Result{AffectedRows: 1}, nil
}

func executeSelect(state databaseState, stmt selectStatement) (Result, error) {
	tbl, err := findTable(state, stmt.table)
	if err != nil {
		return Result{}, err
	}
	projection := make([]int, 0, len(tbl.Columns))
	columns := stmt.columns
	if stmt.allColumns {
		columns = make([]string, len(tbl.Columns))
		for index, column := range tbl.Columns {
			columns[index] = column.Name
		}
	}
	for _, name := range columns {
		index, _, err := findColumn(tbl, name)
		if err != nil {
			return Result{}, err
		}
		projection = append(projection, index)
	}

	matcher, err := buildMatcher(tbl, stmt.where)
	if err != nil {
		return Result{}, err
	}
	rows := make([][]Value, 0)
	for _, storedRow := range tbl.Rows {
		if !matcher(storedRow) {
			continue
		}
		row := make([]Value, len(projection))
		for resultIndex, storedIndex := range projection {
			row[resultIndex] = storedRow[storedIndex]
		}
		rows = append(rows, row)
	}
	return Result{Columns: columns, Rows: rows}, nil
}

func executeUpdate(state *databaseState, stmt updateStatement) (Result, error) {
	tbl, err := findTable(*state, stmt.table)
	if err != nil {
		return Result{}, err
	}
	type resolvedAssignment struct {
		index int
		value Value
	}
	resolved := make([]resolvedAssignment, 0, len(stmt.assignments))
	seen := make(map[int]struct{}, len(stmt.assignments))
	for _, assignment := range stmt.assignments {
		index, column, err := findColumn(tbl, assignment.column)
		if err != nil {
			return Result{}, err
		}
		if _, exists := seen[index]; exists {
			return Result{}, newError(CodeAlreadyExists, "column %q is assigned more than once", assignment.column)
		}
		value, err := coerceLiteral(assignment.value, column)
		if err != nil {
			return Result{}, err
		}
		resolved = append(resolved, resolvedAssignment{index: index, value: value})
		seen[index] = struct{}{}
	}
	matcher, err := buildMatcher(tbl, stmt.where)
	if err != nil {
		return Result{}, err
	}

	affected := 0
	for rowIndex, storedRow := range tbl.Rows {
		if !matcher(storedRow) {
			continue
		}
		updated := append([]Value(nil), storedRow...)
		for _, assignment := range resolved {
			updated[assignment.index] = assignment.value
		}
		if err := validateRow(tbl, updated, rowIndex); err != nil {
			return Result{}, err
		}
		tbl.Rows[rowIndex] = updated
		affected++
	}
	return Result{AffectedRows: affected}, nil
}

func executeDelete(state *databaseState, stmt deleteStatement) (Result, error) {
	tbl, err := findTable(*state, stmt.table)
	if err != nil {
		return Result{}, err
	}
	matcher, err := buildMatcher(tbl, stmt.where)
	if err != nil {
		return Result{}, err
	}
	kept := make([][]Value, 0, len(tbl.Rows))
	affected := 0
	for _, row := range tbl.Rows {
		if matcher(row) {
			affected++
			continue
		}
		kept = append(kept, row)
	}
	tbl.Rows = kept
	return Result{AffectedRows: affected}, nil
}

func findTable(state databaseState, name string) (*table, error) {
	tbl, exists := state.Tables[strings.ToLower(name)]
	if !exists {
		return nil, newError(CodeNotFound, "table %q does not exist", name)
	}
	return tbl, nil
}

func findColumn(tbl *table, name string) (int, Column, error) {
	for index, column := range tbl.Columns {
		if column.Name == strings.ToLower(name) {
			return index, column, nil
		}
	}
	return 0, Column{}, newError(CodeNotFound, "column %q does not exist in table %q", name, tbl.Name)
}

func buildMatcher(tbl *table, where *predicate) (func([]Value) bool, error) {
	if where == nil {
		return func([]Value) bool { return true }, nil
	}
	index, column, err := findColumn(tbl, where.column)
	if err != nil {
		return nil, err
	}
	switch where.op {
	case predicateIsNull:
		return func(row []Value) bool { return row[index].Null }, nil
	case predicateIsNotNull:
		return func(row []Value) bool { return !row[index].Null }, nil
	}
	// coerceLiteral would reject NULL against a NOT NULL column, so it is only
	// reached for an equality predicate.
	expected, err := coerceLiteral(where.value, column)
	if err != nil {
		return nil, err
	}
	return func(row []Value) bool { return valuesEqual(row[index], expected) }, nil
}

func validateRow(tbl *table, row []Value, excludedRow int) error {
	for columnIndex, column := range tbl.Columns {
		value := row[columnIndex]
		if value.Null && (column.NotNull || column.PrimaryKey) {
			return newError(CodeConstraint, "column %q cannot be NULL", column.Name)
		}
		if !column.Unique && !column.PrimaryKey || value.Null {
			continue
		}
		for rowIndex, existing := range tbl.Rows {
			if rowIndex != excludedRow && valuesEqual(existing[columnIndex], value) {
				return newError(CodeConstraint, "duplicate value for unique column %q", column.Name)
			}
		}
	}
	return nil
}

func cloneState(state databaseState) databaseState {
	clone := databaseState{Version: state.Version, Tables: make(map[string]*table, len(state.Tables))}
	for name, source := range state.Tables {
		tableClone := &table{Name: source.Name, Columns: append([]Column(nil), source.Columns...), Rows: make([][]Value, len(source.Rows))}
		for index, row := range source.Rows {
			tableClone.Rows[index] = append([]Value(nil), row...)
		}
		clone.Tables[name] = tableClone
	}
	return clone
}
