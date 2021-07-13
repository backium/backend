package errors

import (
	"errors"
	"fmt"
)

type Op string
type Kind int

const (
	KindUnexpected Kind = iota + 1
	KindValidation
	KindUserExist
	KindNotFound
	KindNoPermission
	KindInvalidCredentials
	KindInvalidSession
)

type Error struct {
	Op   Op
	Kind Kind
	Err  error
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func E(args ...interface{}) error {
	e := &Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case Op:
			e.Op = arg
		case Kind:
			e.Kind = arg
		case error:
			e.Err = arg
		case string:
			e.Err = errors.New(arg)
		default:
			panic("bad call to errors.E")
		}
	}
	return e
}

func Is(err error, kind Kind) bool {
	e, ok := err.(*Error)
	if !ok {
		return KindUnexpected == kind
	}
	if e.Kind != 0 {
		return e.Kind == kind
	}
	return Is(e.Err, kind)
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

// Errorf is equivalent to fmt.Errorf, but allows clients to import only this
// package for all error handling.
func Errorf(format string, args ...interface{}) error {
	return &errorString{fmt.Sprintf(format, args...)}
}
