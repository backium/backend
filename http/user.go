package http

import (
	"net/http"

	"github.com/backium/backend/core"
	"github.com/backium/backend/errors"
	"github.com/labstack/echo/v4"
)

func (h *Handler) HandleRegisterOwner(c echo.Context) error {
	const op = errors.Op("http/Handler.RegisterOwner")

	type request struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,password"`
	}

	ctx := c.Request().Context()

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

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

func (h *Handler) HandleRegisterEmployee(c echo.Context) error {
	const op = errors.Op("http/Handler.RegisterEmployee")

	type request struct {
		Email      string  `json:"email" validate:"required,email"`
		Password   string  `json:"password" validate:"required,password"`
		EmployeeID core.ID `json:"employee_id" validate:"required"`
	}

	ctx := c.Request().Context()

	merchant := core.MerchantFromContext(ctx)
	if merchant == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	user := core.NewUserEmployee(merchant.ID, req.EmployeeID)
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
		EmployeeID:   user.EmployeeID,
		ExistingUser: false,
	})
}

func (h *Handler) HandleLogin(c echo.Context) error {
	const op = errors.Op("authHandler.Login")

	type request struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	ctx := c.Request().Context()

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return errors.E(op, err)
	}

	user, err := h.UserService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return errors.E(op, err)
	}

	if err := h.setSession(c, user); err != nil {
		return errors.E(op, err)
	}

	return c.JSON(http.StatusOK, NewUser(user))
}

func (h *Handler) HandleUniversalLogin(c echo.Context) error {
	const op = errors.Op("http/Handler.UniversalSignin")

	type request struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	ctx := c.Request().Context()

	req := request{}
	if err := bindAndValidate(c, &req); err != nil {
		return errors.E(op, err)
	}

	user, err := h.UserService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return errors.E(op, err)
	}

	s := core.NewSession(user)
	if err := h.SessionRepository.Set(ctx, s); err != nil {
		return errors.E(op, err)
	}
	c.Response().Header().Set("session-id", string(s.ID))

	return c.NoContent(http.StatusOK)
}

func (h *Handler) HandleUniversalGetSession(c echo.Context) error {
	const op = errors.Op("http/Handler.UniversalSignin")

	ctx := c.Request().Context()

	sid := core.ID(c.QueryParam("sid"))
	s, err := h.SessionRepository.Get(ctx, sid)
	if err != nil {
		return errors.E(op, errors.KindInvalidCredentials, err)
	}
	token, err := s.Encode([]byte("backium"))

	c.SetCookie(&http.Cookie{
		Name:  "web_session",
		Value: token,
		Path:  "/api/v1",
	})

	return c.NoContent(http.StatusOK)
}

func (h *Handler) HandleLogout(c echo.Context) error {
	const op = errors.Op("http/Handler.Logout")

	ctx := c.Request().Context()

	session := core.SessionFromContext(ctx)
	if session == nil {
		return errors.E(op, errors.KindUnexpected, "invalid echo.Context")
	}

	h.SessionRepository.Delete(ctx, session.ID)

	return c.NoContent(http.StatusOK)
}

func (h *Handler) setSession(c echo.Context, u core.User) error {
	const op = errors.Op("authHandler.setSession")

	ctx := c.Request().Context()

	session := core.NewSession(u)
	token, err := session.Encode([]byte("backium"))
	if err != nil {
		return errors.E(op, err)
	}
	if err := h.SessionRepository.Set(ctx, session); err != nil {
		return errors.E(op, err)
	}

	c.SetCookie(&http.Cookie{
		Name:  "web_session",
		Value: token,
	})

	return nil
}

type User struct {
	ID         core.ID `json:"id"`
	Email      string  `json:"email"`
	IsOwner    bool    `json:"is_owner"`
	EmployeeID core.ID `json:"employee_id"`
	MerchantID core.ID `json:"merchant_id"`
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

type RegisterResponse struct {
	UserID       core.ID `json:"user_id,omitempty"`
	EmployeeID   core.ID `json:"employee_id,omitempty"`
	MerchantID   core.ID `json:"merchant_id,omitempty"`
	ExistingUser bool    `json:"existing_user"`
}
