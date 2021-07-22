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

func ContainsOneID(given []ID, target []ID) bool {
	for _, id := range target {
		if ContainsID(given, id) {
			return true
		}
	}
	return false
}

func ContainsAllID(given []ID, target []ID) bool {
	for _, id := range target {
		if !ContainsID(given, id) {
			return false
		}
	}
	return true
}

func ContainsID(given []ID, target ID) bool {
	for _, id := range given {
		if id == target {
			return true
		}
	}
	return false
}

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
