package handler

import (
	"context"

	"github.com/backium/backend/entity"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type AuthContext struct {
	echo.Context
	Session
}

type SessionRepository interface {
	Set(context.Context, Session) error
	Get(context.Context, string) (Session, error)
	Delete(context.Context, string) error
}

type Session struct {
	ID         string
	UserID     string
	MerchantID string
	IsSuper    bool
	IsOwner    bool
}

func newSession(u entity.User) Session {
	id, err := gonanoid.New()
	if err != nil {
		panic(err)
	}
	return Session{
		ID:         id,
		UserID:     u.ID,
		MerchantID: u.MerchantID,
		IsSuper:    u.IsSuper,
		IsOwner:    u.IsOwner,
	}
}

func DecodeSession(encodedSession string) (Session, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(encodedSession, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("backium"), nil
	})
	if err != nil {
		return Session{}, err
	}
	return Session{
		ID:         claims["session_id"].(string),
		UserID:     claims["user_id"].(string),
		MerchantID: claims["merchant_id"].(string),
		IsSuper:    claims["is_super"].(bool),
	}, nil
}

func (s *Session) encode(key []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"session_id":  s.ID,
		"user_id":     s.UserID,
		"merchant_id": s.MerchantID,
		"is_super":    s.IsSuper,
	})
	return token.SignedString(key)
}
