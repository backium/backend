package http

import (
	"fmt"
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

func (h *Handler) RegisterOwner(c echo.Context) error {
	const op = errors.Op("http/Handler.RegisterOwner")
	req := RegisterRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	user := core.NewUserOwner()
	user.Email = req.Email

	user, err := h.UserService.Create(ctx, user, req.Password)
	if errors.Is(err, errors.KindUserExist) {
		return c.JSON(http.StatusOK, RegisterResponse{
			ExistingUser: true,
		})
	}
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, RegisterResponse{
		UserID:       user.ID,
		MerchantID:   user.MerchantID,
		ExistingUser: false,
	})
}

func (h *Handler) RegisterEmployee(c echo.Context) error {
	const op = errors.Op("http/Handler.RegisterEmployee")
	ac, ok := c.(*AuthContext)
	if !ok {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}
	req := RegisterEmployeeRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	user := core.NewUserEmployee(ac.MerchantID, req.EmployeeID)
	user.Email = req.Email

	fmt.Println(user, req.Password)
	user, err := h.UserService.Create(ctx, user, req.Password)
	if errors.Is(err, errors.KindUserExist) {
		return c.JSON(http.StatusOK, RegisterResponse{
			ExistingUser: true,
		})
	}
	if err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, RegisterResponse{
		UserID:       user.ID,
		MerchantID:   user.MerchantID,
		EmployeeID:   user.EmployeeID,
		ExistingUser: false,
	})
}

func (h *Handler) Login(c echo.Context) error {
	const op = errors.Op("authHandler.Login")
	req := LoginRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return errors.E(op, err)
	}
	ctx := c.Request().Context()
	user, err := h.UserService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return errors.E(op, err)
	}
	if err := h.setSession(c, user); err != nil {
		return errors.E(op, err)
	}
	return c.JSON(http.StatusOK, NewUser(user))
}

func (h *Handler) UniversalLogin(c echo.Context) error {
	const op = errors.Op("http/Handler.UniversalSignin")
	req := LoginRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return errors.E(op, err)
	}
	ctx := c.Request().Context()
	user, err := h.UserService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return errors.E(op, err)
	}
	s := newSession(user)
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

func (h *Handler) Logout(c echo.Context) error {
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
	EmployeeID string `json:"employee_id"`
	MerchantID string `json:"merchant_id"`
}

func NewUser(user core.User) User {
	return User{
		ID:         user.ID,
		Email:      user.Email,
		IsOwner:    user.Kind == core.UserKindOwner,
		EmployeeID: user.EmployeeID,
		MerchantID: user.MerchantID,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

type RegisterEmployeeRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,password"`
	EmployeeID string `json:"employee_id" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterResponse struct {
	UserID       string `json:"user_id,omitempty"`
	EmployeeID   string `json:"employee_id,omitempty"`
	MerchantID   string `json:"merchant_id,omitempty"`
	ExistingUser bool   `json:"existing_user"`
}
