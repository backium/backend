package repository

import gonanoid "github.com/matoous/go-nanoid/v2"

const (
	alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	idSize   = 14
)

func generateID(prefix string) string {
	id, err := gonanoid.Generate(alphabet, idSize)
	if err != nil {
		panic(err)
	}
	return prefix + "_" + id
}
