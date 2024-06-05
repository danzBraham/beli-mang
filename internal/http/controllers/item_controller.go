package controllers

import (
	"errors"
	"net/http"
	"strconv"

	item_entity "github.com/danzBraham/beli-mang/internal/entities/item"
	merchant_exception "github.com/danzBraham/beli-mang/internal/exceptions/merchant"
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
	r.Get("/", c.handleGetItems)

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
	payload := &item_entity.AddItemRequest{}

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

	itemResponse, err := c.Service.CreateItem(r.Context(), merchantId, payload)
	if errors.Is(err, merchant_exception.ErrMerchantIdNotFound) {
		http_helper.ResponseError(w, http.StatusNotFound, "Not found error", err.Error())
		return
	}
	if err != nil {
		http_helper.ResponseError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	http_helper.EncodeJSON(w, http.StatusCreated, &itemResponse)
}

func (c *ItemController) handleGetItems(w http.ResponseWriter, r *http.Request) {
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
	query := r.URL.Query()

	params := &item_entity.ItemQueryParams{
		Id:        query.Get("itemId"),
		Limit:     5,
		Offset:    0,
		Name:      query.Get("name"),
		Category:  query.Get("productCategory"),
		CreatedAt: query.Get("createdAt"),
	}

	if limit := query.Get("limit"); limit != "" {
		params.Limit, _ = strconv.Atoi(limit)
	}

	if offset := query.Get("offset"); offset != "" {
		params.Offset, _ = strconv.Atoi(offset)
	}

	itemsResponse, err := c.Service.GetItems(r.Context(), merchantId, params)
	if errors.Is(err, merchant_exception.ErrMerchantIdNotFound) {
		http_helper.ResponseError(w, http.StatusNotFound, "Not found error", err.Error())
		return
	}
	if err != nil {
		http_helper.ResponseError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	http_helper.EncodeJSON(w, http.StatusOK, &itemsResponse)
}
