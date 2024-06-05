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

type UserController struct {
	Service services.UserService
}

func NewUserController(service services.UserService) *UserController {
	return &UserController{Service: service}
}

func (c *UserController) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/register", c.handleRegisterUser)
	r.Post("/login", c.handleLoginUser)

	return r
}

func (c *UserController) handleRegisterUser(w http.ResponseWriter, r *http.Request) {
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

	userRepsonse, err := c.Service.RegisterUser(r.Context(), payload)
	if errors.Is(err, user_exception.ErrUsernameAlreadyExists) {
		http_helper.ResponseError(w, http.StatusConflict, "Conflict error", err.Error())
		return
	}
	if errors.Is(err, user_exception.ErrUserEmailAlreadyExists) {
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

func (c *UserController) handleLoginUser(w http.ResponseWriter, r *http.Request) {
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

	userRepsonse, err := c.Service.LoginUser(r.Context(), payload)
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
