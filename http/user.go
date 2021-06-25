package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

func (h *Handler) Signup(c echo.Context) error {
	const op = errors.Op("authHandler.Signup")
	req := UserCreateRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	u, err := h.UserService.Create(c.Request().Context(), core.UserCreateRequest{
		Email:    req.Email,
		Password: req.Password,
		IsOwner:  true,
	})
	if errors.Is(err, errors.KindUserExist) {
		return c.JSON(http.StatusOK, SignupResponse{
			ExistingUser: true,
		})
	}
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, SignupResponse{
		UserID:       u.ID,
		MerchantID:   u.MerchantID,
		ExistingUser: false,
	})
}

func (h *Handler) Login(c echo.Context) error {
	const op = errors.Op("authHandler.Login")
	req := UserLoginRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return errors.E(op, err)
	}
	u, err := h.UserService.Login(c.Request().Context(), core.UserLoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return errors.E(op, err)
	}
	if err := h.setSession(c, u); err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewUser(u))
}

func (h *Handler) UniversalSignin(c echo.Context) error {
	const op = errors.Op("http/Handler.UniversalSignin")
	req := UserLoginRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return errors.E(op, err)
	}
	ctx := c.Request().Context()
	u, err := h.UserService.Login(ctx, core.UserLoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return errors.E(op, err)
	}
	s := newSession(u)
	if err := h.SessionRepository.Set(ctx, s); err != nil {
		return errors.E(op, err)
	}
	c.Response().Header().Set("session-id", s.ID)
	return c.NoContent(http.StatusOK)
}

func (h *Handler) UniversalGetSession(c echo.Context) error {
	const op = errors.Op("http/Handler.UniversalSignin")
	sid := c.QueryParam("sid")
	ctx := c.Request().Context()
	s, err := h.SessionRepository.Get(ctx, sid)
	if err != nil {
		return errors.E(op, errors.KindInvalidCredentials, err)
	}
	stoken, err := s.encode([]byte("backium"))
	c.SetCookie(&http.Cookie{
		Name:  "web_session",
		Value: stoken,
		Path:  "/api/v1",
	})
	return c.NoContent(http.StatusOK)
}

func (h *Handler) Signout(c echo.Context) error {
	ac := c.(*AuthContext)
	h.SessionRepository.Delete(c.Request().Context(), ac.Session.ID)
	return c.JSONBlob(http.StatusOK, []byte("{}"))
}

func (h *Handler) setSession(c echo.Context, u core.User) error {
	const op = errors.Op("authHandler.setSession")
	s := newSession(u)
	stoken, err := s.encode([]byte("backium"))
	if err != nil {
		return errors.E(op, err)
	}
	if err := h.SessionRepository.Set(c.Request().Context(), s); err != nil {
		return errors.E(op, err)
	}
	c.SetCookie(&http.Cookie{
		Name:  "web_session",
		Value: stoken,
	})
	return nil
}

type User struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	IsOwner    bool   `json:"is_owner"`
	MerchantID string `json:"merchant_id"`
}

func NewUser(u core.User) User {
	return User{
		ID:         u.ID,
		Email:      u.Email,
		IsOwner:    u.Kind == core.UserKindOwner,
		MerchantID: u.MerchantID,
	}
}

type UserCreateRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type SignupResponse struct {
	UserID       string `json:"user_id,omitempty"`
	MerchantID   string `json:"merchant_id,omitempty"`
	ExistingUser bool   `json:"existing_user"`
}
