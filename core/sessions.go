package core

import (
	"context"

	"github.com/backium/backend/errors"
	"github.com/dgrijalva/jwt-go"
)

type SessionStorage interface {
	Set(context.Context, Session) error
	Get(context.Context, ID) (Session, error)
	Delete(context.Context, ID) error
}

type Session struct {
	ID         ID
	UserID     ID
	MerchantID ID
	Kind       UserKind
}

func NewSession(u User) Session {
	id := NewID("sess")
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
		ID:         ID(id),
		UserID:     ID(userID),
		MerchantID: ID(merchantID),
		Kind:       UserKind(kind),
	}, nil
}

func (s *Session) Encode(key []byte) (string, error) {
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
