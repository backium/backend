package errors

import "errors"

type Op string
type Kind int

const (
	KindUnexpected Kind = iota + 1
	KindValidation
	KindUserExist
	KindNotFound
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
