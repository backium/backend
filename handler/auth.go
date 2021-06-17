package handler

import (
	"net/http"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

type userResource struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	IsOwner    bool   `json:"is_owner"`
	MerchantID string `json:"merchant_id"`
}

func userResourceFrom(u entity.User) userResource {
	return userResource{
		ID:         u.ID,
		Email:      u.Email,
		IsOwner:    u.Kind == entity.UserKindOwner,
		MerchantID: u.MerchantID,
	}
}

type createUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

type loginUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type signupResponse struct {
	UserID       string `json:"user_id,omitempty"`
	MerchantID   string `json:"merchant_id,omitempty"`
	ExistingUser bool   `json:"existing_user"`
}

type Auth struct {
	Controller controller.User
	SessionRepository
}

func (h *Auth) Signup(c echo.Context) error {
	const op = errors.Op("authHandler.Signup")
	req := createUserRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	u, err := h.Controller.Create(c.Request().Context(), controller.CreateUserRequest{
		Email:    req.Email,
		Password: req.Password,
		IsOwner:  true,
	})
	if errors.Is(err, errors.KindUserExist) {
		return c.JSON(http.StatusOK, signupResponse{
			ExistingUser: true,
		})
	}
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, signupResponse{
		UserID:       u.ID,
		MerchantID:   u.MerchantID,
		ExistingUser: false,
	})
}

func (h *Auth) Login(c echo.Context) error {
	const op = errors.Op("authHandler.Login")
	req := loginUserRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return errors.E(op, err)
	}
	u, err := h.Controller.Login(c.Request().Context(), controller.LoginUserRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return errors.E(op, err)
	}
	if err := h.setSession(c, u); err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, userResourceFrom(u))
}

func (h *Auth) Signout(c echo.Context) error {
	ac := c.(*AuthContext)
	h.Delete(c.Request().Context(), ac.ID)
	return c.JSONBlob(http.StatusOK, []byte("{}"))
}

func (h *Auth) setSession(c echo.Context, u entity.User) error {
	const op = errors.Op("authHandler.setSession")
	s := newSession(u)
	stoken, err := s.encode([]byte("backium"))
	if err != nil {
		return errors.E(op, err)
	}
	if err := h.Set(c.Request().Context(), s); err != nil {
		return errors.E(op, err)
	}
	c.SetCookie(&http.Cookie{
		Name:  "web_session",
		Value: stoken,
	})
	return nil
}

func (h *Auth) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
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
		rs, err := h.Get(c.Request().Context(), ds.ID)
		if err != nil {
			return errors.E(op, errors.KindInvalidSession, err)
		}
		c.Logger().Infof("session found: %+v", rs)

		return next(&AuthContext{
			Context: c,
			Session: rs,
		})
	}
}
