package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	ErrBadRequest          = New(http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	ErrInternalServerError = New(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	ErrServerClosed        = ErrInternalServerError.Wrap(http.ErrServerClosed)
)

type Error struct {
	Status int    `json:"status,omitempty"`
	Detail string `json:"detail,omitempty"`
	Type   string `json:"type,omitempty"`
	Title  string `json:"title"`
	extra  []string
	err    error
}

func (e Error) Unwrap() error {
	return e.err
}

func (e Error) Wrap(err error) Error {
	if err != nil {
		e.err = err
	}
	return e
}

func (e Error) Error() string {
	msg := fmt.Sprintf("%s: %v", e.Title, e.err)

	if e.Status > 0 && e.Detail != "" {
		msg = fmt.Sprintf("%s (%d), %s: %v", e.Title, e.Status, e.Detail, e.err)
	} else if e.Status > 0 {
		msg = fmt.Sprintf("%s (%d): %v", e.Title, e.Status, e.err)
	}

	return msg
}

func (e Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(e)
}

func New(msg string, code int) Error {
	return Error{
		Title:  msg,
		Status: code,
	}
}
