package handler

import (
	"net/http"

	"github.com/backium/backend/controller"
	"github.com/backium/backend/entity"
	"github.com/labstack/echo/v4"
)

type userResource struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	IsOwner    bool   `json:"is_owner"`
	MerchantID string `json:"merchant_id"`
}

type createUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

type loginUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Auth struct {
	Controller controller.User
	SessionRepository
}

func (h *Auth) Signup(c echo.Context) error {
	req := createUserRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	u, err := h.Controller.Create(c.Request().Context(), controller.CreateUserRequest{
		Email:    req.Email,
		Password: req.Password,
		IsOwner:  true,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, userResourceFrom(u))
}

func (h *Auth) Login(c echo.Context) error {
	req := loginUserRequest{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	u, err := h.Controller.Login(c.Request().Context(), controller.LoginUserRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.setSession(c, u); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, userResourceFrom(u))
}

func (h *Auth) Signout(c echo.Context) error {
	ac := c.(*AuthContext)
	h.Delete(c.Request().Context(), ac.ID)
	return c.JSONBlob(http.StatusOK, []byte("{}"))
}

func (h *Auth) setSession(c echo.Context, u entity.User) error {
	s := newSession(u)
	stoken, err := s.encode([]byte("backium"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if err := h.Set(c.Request().Context(), s); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	c.SetCookie(&http.Cookie{
		Name:  "web_session",
		Value: stoken,
	})
	return nil
}

func (h *Auth) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("web_session")
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: session cookie missing")
		}
		ds, err := DecodeSession(cookie.Value)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: error parsing session")
		}
		rs, err := h.Get(c.Request().Context(), ds.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized: session not found")
		}
		c.Logger().Infof("session found: %+v", rs)

		return next(&AuthContext{
			Context: c,
			Session: rs,
		})
	}
}

func userResourceFrom(u entity.User) userResource {
	return userResource{
		ID:         u.ID,
		Email:      u.Email,
		IsOwner:    u.IsOwner,
		MerchantID: u.MerchantID,
	}
}
