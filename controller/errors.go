package controller

import "errors"

// Repository errors
var (
	ErrRepositoryInternal   = errors.New("internal repository error")
	ErrRepositoryNoDocument = errors.New("document not found")
)
