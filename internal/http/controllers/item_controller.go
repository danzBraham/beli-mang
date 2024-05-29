package controllers

import (
	"net/http"

	item_entity "github.com/danzBraham/beli-mang/internal/entities/item"
	http_helper "github.com/danzBraham/beli-mang/internal/helpers/http"
	validator_helper "github.com/danzBraham/beli-mang/internal/helpers/validator"
	"github.com/danzBraham/beli-mang/internal/http/middlewares"
	"github.com/danzBraham/beli-mang/internal/services"
	"github.com/go-chi/chi/v5"
)

type ItemController struct {
	Service services.ItemService
}

func NewItemController(service services.ItemService) *ItemController {
	return &ItemController{Service: service}
}

func (c *ItemController) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middlewares.Authenticate)
	r.Post("/", c.handleAddItem)

	return r
}

func (c *ItemController) handleAddItem(w http.ResponseWriter, r *http.Request) {
	isAdmin, ok := r.Context().Value(middlewares.ContextIsAdminKey).(bool)
	if !ok {
		http_helper.ResponseError(w, http.StatusUnauthorized, "IsAdmin type assertion failed", "IsAdmin not found in the context")
		return
	}
	if !isAdmin {
		http_helper.ResponseError(w, http.StatusUnauthorized, "Unauthorized error", "you're not admin")
		return
	}

	merchantId := chi.URLParam(r, "merchantId")
	paylaod := &item_entity.AddItemRequest{}

	err := http_helper.DecodeJSON(r, paylaod)
	if err != nil {
		http_helper.ResponseError(w, http.StatusBadRequest, err.Error(), "Failed to decode JSON")
		return
	}

	err = validator_helper.ValidatePayload(paylaod)
	if err != nil {
		http_helper.ResponseError(w, http.StatusBadRequest, err.Error(), "Request doesn't pass validation")
		return
	}

	itemResponse, err := c.Service.CreateItem(r.Context(), merchantId, paylaod)
	if err != nil {
		http_helper.ResponseError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	http_helper.EncodeJSON(w, http.StatusCreated, &itemResponse)
}
