package database

import (
	"bytes"
	"fmt"
	"text/tabwriter"
)

// Result is the transport-independent outcome of one SQL statement.
type Result struct {
	Columns      []string
	Rows         [][]Value
	AffectedRows int
	Message      string
}

// Format returns a stable, human-readable representation for the v0.1 wire
// protocol. The protocol can later carry Result structurally without changing
// the engine.
func (r Result) Format() string {
	if len(r.Columns) == 0 {
		if r.Message != "" {
			return r.Message
		}
		return fmt.Sprintf("%d row(s) affected", r.AffectedRows)
	}

	var output bytes.Buffer
	w := tabwriter.NewWriter(&output, 0, 4, 2, ' ', 0)
	for index, column := range r.Columns {
		if index > 0 {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, column)
	}
	fmt.Fprintln(w)
	for index := range r.Columns {
		if index > 0 {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, "---")
	}
	fmt.Fprintln(w)
	for _, row := range r.Rows {
		for index, value := range row {
			if index > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, value.String())
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintf(w, "(%d row(s))\n", len(r.Rows))
	_ = w.Flush()
	return output.String()
}
