package http

import (
	"context"
	"strings"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

type AuthContext struct {
	echo.Context
	Session    core.Session
	MerchantID string
}

func RequireAPIKey(merchantStorage core.MerchantStorage, sessionStorage core.SessionRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			const op = errors.Op("http/RequireAPIKey")

			header := c.Request().Header
			bearer := header.Get("Authorization")
			apiKey := strings.TrimPrefix(bearer, "Bearer ")

			if strings.HasPrefix(apiKey, "sk_") {
				merch, err := merchantStorage.GetByKey(context.TODO(), apiKey)
				if err != nil {
					return errors.E(op, errors.KindInvalidSession, err)
				}

				return next(&AuthContext{
					Context:    c,
					Session:    core.Session{},
					MerchantID: merch.ID,
				})
			}
			return nil
		}
	}
}

func RequireSession(merchantStorage core.MerchantStorage, sessionStorage core.SessionRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			const op = errors.Op("handler.Auth.Authenticate")
			req := c.Request()
			ctx := req.Context()

			cookie, err := c.Cookie("web_session")
			if err != nil {
				return errors.E(op, errors.KindInvalidSession, err)
			}
			session, err := core.DecodeSession(cookie.Value)
			if err != nil {
				return errors.E(op, errors.KindInvalidSession, err)
			}
			session, err = sessionStorage.Get(ctx, session.ID)
			if err != nil {
				return errors.E(op, errors.KindInvalidSession, err)
			}
			merchant, err := merchantStorage.Get(ctx, session.MerchantID)
			if err != nil {
				return errors.E(op, errors.KindInvalidSession, err)
			}
			c.Logger().Infof("session found: %+v", session)

			ctx = core.ContextWithMerchant(ctx, &merchant)
			ctx = core.ContextWithSession(ctx, &session)
			c.SetRequest(req.Clone(ctx))

			return next(c)
		}
	}
}
