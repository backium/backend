package core

import (
	"strings"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	idSize   = 14
)

type ID string

func NewID(prefix string) ID {
	return NewIDWithSize(prefix, idSize)
}

func NewIDWithSize(prefix string, size int) ID {
	id, err := gonanoid.Generate(alphabet, size)
	if err != nil {
		panic(err)
	}
	return ID(prefix + "_" + id)
}

func (id *ID) Validate() bool {
	p := strings.Split(string(*id), "_")
	if len(p) != 2 {
		return false
	}
	if len(p[1]) != idSize {
		return false
	}
	return true
}
