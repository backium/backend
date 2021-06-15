package handler

import (
	"context"

	"github.com/dgrijalva/jwt-go"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type SessionRepository interface {
	Set(context.Context, Session) error
	Get(context.Context, string) (Session, error)
}

type Session struct {
	ID     string
	UserID string
}

func newSession(userID string) Session {
	id, err := gonanoid.New()
	if err != nil {
		panic(err)
	}
	return Session{
		ID:     id,
		UserID: userID,
	}
}

func (s *Session) encode(key []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"session_id": s.ID,
		"user_id":    s.UserID,
	})
	return token.SignedString(key)
}
