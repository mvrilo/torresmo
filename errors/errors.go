package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var (
	ErrBadRequest = New(http.StatusText(http.StatusBadRequest)).WithCode(http.StatusBadRequest)
	ErrInternal   = New(http.StatusText(http.StatusInternalServerError)).WithCode(http.StatusInternalServerError)
)

type Error struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"message"`
	err  error
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) Wrap(err error) *Error {
	if err != nil {
		e.err = err
	}
	return e
}

func (e *Error) Error() string {
	var msg string

	switch err := e.err.(type) {
	case validator.ValidationErrors:
		msg = fmt.Sprintf("%s is %s", err[0].Field(), err[0].ActualTag())
	default:
		msg = err.Error()
	}

	return fmt.Sprintf("%s (%d): %s", e.Msg, e.Code, msg)
}

func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(e)
}

func (e *Error) WithCode(code int) *Error {
	e.Code = code
	return e
}

func New(msg string) *Error {
	return &Error{
		Msg: msg,
	}
}
