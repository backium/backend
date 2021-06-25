package http

import (
	"context"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type AuthContext struct {
	echo.Context
	Session    Session
	MerchantID string
}

func (h *Handler) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		const op = errors.Op("handler.Auth.Authenticate")
		cookie, err := c.Cookie("web_session")
		if err != nil {
			return errors.E(op, errors.KindInvalidSession, err)
		}
		ds, err := DecodeSession(cookie.Value)
		if err != nil {
			return errors.E(op, errors.KindInvalidSession, err)
		}
		rs, err := h.SessionRepository.Get(c.Request().Context(), ds.ID)
		if err != nil {
			return errors.E(op, errors.KindInvalidSession, err)
		}
		c.Logger().Infof("session found: %+v", rs)

		return next(&AuthContext{
			Context:    c,
			Session:    rs,
			MerchantID: rs.MerchantID,
		})
	}
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
	Kind       core.UserKind
}

func newSession(u core.User) Session {
	id, err := gonanoid.New()
	if err != nil {
		panic(err)
	}
	return Session{
		ID:         id,
		UserID:     u.ID,
		MerchantID: u.MerchantID,
		Kind:       u.Kind,
	}
}

func DecodeSession(encodedSession string) (Session, error) {
	const op = "handler.DecodeSession"
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(encodedSession, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("backium"), nil
	})
	if err != nil {
		return Session{}, errors.E(op, errors.KindInvalidSession, err)
	}
	id, ok := claims["session_id"].(string)
	if !ok {
		return Session{}, errors.E(op, errors.KindInvalidSession, "missing or invalid session_id")
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		return Session{}, errors.E(op, errors.KindInvalidSession, "missing or invalid user_id")
	}
	merchantID, ok := claims["merchant_id"].(string)
	if !ok {
		return Session{}, errors.E(op, errors.KindInvalidSession, "missing or invalid merchant_id")
	}
	kind, ok := claims["kind"].(string)
	if !ok {
		return Session{}, errors.E(op, errors.KindInvalidSession, "missing or invalid kind")
	}
	return Session{
		ID:         id,
		UserID:     userID,
		MerchantID: merchantID,
		Kind:       core.UserKind(kind),
	}, nil
}

func (s *Session) encode(key []byte) (string, error) {
	const op = errors.Op("Session.encode")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"session_id":  s.ID,
		"user_id":     s.UserID,
		"merchant_id": s.MerchantID,
		"kind":        s.Kind,
	})
	sig, err := token.SignedString(key)
	if err != nil {
		return "", errors.E(op, errors.KindInvalidSession, err)
	}
	return sig, nil

}
