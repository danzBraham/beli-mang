package controllers

import (
	"errors"
	"net/http"
	"time"

	user_entity "github.com/danzBraham/beli-mang/internal/entities/user"
	user_exception "github.com/danzBraham/beli-mang/internal/exceptions/user"
	http_helper "github.com/danzBraham/beli-mang/internal/helpers/http"
	validator_helper "github.com/danzBraham/beli-mang/internal/helpers/validator"
	"github.com/danzBraham/beli-mang/internal/services"
	"github.com/go-chi/chi/v5"
)

type AdminController struct {
	Service services.UserService
}

func NewAdminController(service services.UserService) *AdminController {
	return &AdminController{Service: service}
}

func (c *AdminController) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/register", c.handleRegisterAdminUser)
	r.Post("/login", c.handleLoginAdminUser)

	return r
}

func (c *AdminController) handleRegisterAdminUser(w http.ResponseWriter, r *http.Request) {
	payload := &user_entity.RegisterUserRequest{}

	err := http_helper.DecodeJSON(r, payload)
	if err != nil {
		http_helper.ResponseError(w, http.StatusBadRequest, err.Error(), "Failed to decode JSON")
		return
	}

	err = validator_helper.ValidatePayload(payload)
	if err != nil {
		http_helper.ResponseError(w, http.StatusBadRequest, err.Error(), "Request doesn't pass validation")
		return
	}

	userRepsonse, err := c.Service.RegisterAdminUser(r.Context(), payload)
	if errors.Is(err, user_exception.ErrUsernameAlreadyExists) {
		http_helper.ResponseError(w, http.StatusConflict, "Conflict error", err.Error())
		return
	}
	if errors.Is(err, user_exception.ErrAdminEmailAlreadyExists) {
		http_helper.ResponseError(w, http.StatusConflict, "Conflict error", err.Error())
		return
	}
	if err != nil {
		http_helper.ResponseError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	cookie := http.Cookie{
		Name:    "Authorization",
		Value:   userRepsonse.Token,
		Expires: time.Now().Add(2 * time.Hour),
	}
	http.SetCookie(w, &cookie)

	http_helper.EncodeJSON(w, http.StatusCreated, &userRepsonse)
}

func (c *AdminController) handleLoginAdminUser(w http.ResponseWriter, r *http.Request) {
	payload := &user_entity.LoginUserRequest{}

	err := http_helper.DecodeJSON(r, payload)
	if err != nil {
		http_helper.ResponseError(w, http.StatusBadRequest, err.Error(), "Failed to decode JSON")
		return
	}

	err = validator_helper.ValidatePayload(payload)
	if err != nil {
		http_helper.ResponseError(w, http.StatusBadRequest, err.Error(), "Request doesn't pass validation")
		return
	}

	userRepsonse, err := c.Service.LoginAdminUser(r.Context(), payload)
	if errors.Is(err, user_exception.ErrUserNotFound) {
		http_helper.ResponseError(w, http.StatusNotFound, "Not found error", err.Error())
		return
	}
	if errors.Is(err, user_exception.ErrInvalidPassword) {
		http_helper.ResponseError(w, http.StatusBadRequest, "Bad request error", err.Error())
		return
	}
	if err != nil {
		http_helper.ResponseError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	cookie := http.Cookie{
		Name:    "Authorization",
		Value:   userRepsonse.Token,
		Expires: time.Now().Add(2 * time.Hour),
	}
	http.SetCookie(w, &cookie)

	http_helper.EncodeJSON(w, http.StatusOK, &userRepsonse)
}
