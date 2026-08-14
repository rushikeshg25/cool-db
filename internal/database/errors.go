package database

import "fmt"

// ErrorCode categorizes failures so transports can map them without parsing
// human-readable messages.
type ErrorCode string

const (
	CodeSyntax        ErrorCode = "SYNTAX_ERROR"
	CodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	CodeNotFound      ErrorCode = "NOT_FOUND"
	CodeConstraint    ErrorCode = "CONSTRAINT_VIOLATION"
	CodeType          ErrorCode = "TYPE_ERROR"
	CodeStorage       ErrorCode = "STORAGE_ERROR"
)

// Error is a query or storage error returned by Engine.
type Error struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func newError(code ErrorCode, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

func wrapError(code ErrorCode, err error, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...), Cause: err}
}
