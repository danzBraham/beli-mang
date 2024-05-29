package controllers

import (
	"net/http"
	"strconv"

	merchant_entity "github.com/danzBraham/beli-mang/internal/entities/merchant"
	http_helper "github.com/danzBraham/beli-mang/internal/helpers/http"
	validator_helper "github.com/danzBraham/beli-mang/internal/helpers/validator"
	"github.com/danzBraham/beli-mang/internal/http/middlewares"
	"github.com/danzBraham/beli-mang/internal/services"
	"github.com/go-chi/chi/v5"
)

type MerchantController struct {
	Service services.MerchantService
}

func NewMerchantController(service services.MerchantService) *MerchantController {
	return &MerchantController{Service: service}
}

func (c *MerchantController) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middlewares.Authenticate)
	r.Post("/", c.handleAddMerchant)
	r.Get("/", c.handleGetMerchants)

	return r
}

func (c *MerchantController) handleAddMerchant(w http.ResponseWriter, r *http.Request) {
	isAdmin, ok := r.Context().Value(middlewares.ContextIsAdminKey).(bool)
	if !ok {
		http_helper.ResponseError(w, http.StatusUnauthorized, "IsAdmin type assertion failed", "IsAdmin not found in the context")
		return
	}
	if !isAdmin {
		http_helper.ResponseError(w, http.StatusUnauthorized, "Unauthorized error", "you're not admin")
		return
	}

	userId, ok := r.Context().Value(middlewares.ContextUserIdKey).(string)
	if !ok {
		http_helper.ResponseError(w, http.StatusUnauthorized, "IsAdmin type assertion failed", "IsAdmin not found in the context")
		return
	}
	paylaod := &merchant_entity.AddMerchantRequest{}

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

	merchantResponse, err := c.Service.CreateMerchant(r.Context(), userId, paylaod)
	if err != nil {
		http_helper.ResponseError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	http_helper.EncodeJSON(w, http.StatusCreated, &merchantResponse)
}

func (c *MerchantController) handleGetMerchants(w http.ResponseWriter, r *http.Request) {
	isAdmin, ok := r.Context().Value(middlewares.ContextIsAdminKey).(bool)
	if !ok {
		http_helper.ResponseError(w, http.StatusUnauthorized, "IsAdmin type assertion failed", "IsAdmin not found in the context")
		return
	}
	if !isAdmin {
		http_helper.ResponseError(w, http.StatusUnauthorized, "Unauthorized error", "you're not admin")
		return
	}

	query := r.URL.Query()

	params := &merchant_entity.MerchantQueryParams{
		Id:        query.Get("merchantId"),
		Limit:     5,
		Offset:    0,
		Name:      query.Get("name"),
		Category:  query.Get("merchantCategory"),
		CreatedAt: query.Get("createdAt"),
	}

	if limit := query.Get("limit"); limit != "" {
		params.Limit, _ = strconv.Atoi(limit)
	}

	if offset := query.Get("offset"); offset != "" {
		params.Offset, _ = strconv.Atoi(offset)
	}

	merchantsResponse, err := c.Service.GetMerchants(r.Context(), params)
	if err != nil {
		http_helper.ResponseError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	http_helper.EncodeJSON(w, http.StatusOK, &merchantsResponse)
}
