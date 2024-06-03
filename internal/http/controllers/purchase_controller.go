package controllers

import (
	"errors"
	"net/http"
	"strconv"

	purchase_entity "github.com/danzBraham/beli-mang/internal/entities/purchase"
	item_exception "github.com/danzBraham/beli-mang/internal/exceptions/item"
	merchant_exception "github.com/danzBraham/beli-mang/internal/exceptions/merchant"
	purchase_exception "github.com/danzBraham/beli-mang/internal/exceptions/purchase"
	http_helper "github.com/danzBraham/beli-mang/internal/helpers/http"
	validator_helper "github.com/danzBraham/beli-mang/internal/helpers/validator"
	"github.com/danzBraham/beli-mang/internal/http/middlewares"
	"github.com/danzBraham/beli-mang/internal/services"
	"github.com/go-chi/chi/v5"
)

type PurchaseController struct {
	Service services.PurchaseService
}

func NewPurchaseController(service services.PurchaseService) *PurchaseController {
	return &PurchaseController{Service: service}
}

func (c *PurchaseController) MerchantRoutes() chi.Router {
	r := chi.NewRouter()

	r.Use(middlewares.Authenticate)
	r.Get("/nearby/{lat},{long}", c.handleGetMerchantsNearby)

	return r
}

func (c *PurchaseController) handleGetMerchantsNearby(w http.ResponseWriter, r *http.Request) {
	isAdmin, ok := r.Context().Value(middlewares.ContextIsAdminKey).(bool)
	if !ok {
		http_helper.ResponseError(w, http.StatusUnauthorized, "IsAdmin type assertion failed", "IsAdmin not found in the context")
		return
	}
	if isAdmin {
		http_helper.ResponseError(w, http.StatusUnauthorized, "Unauthorized error", "you're not a user")
		return
	}

	lat, err := strconv.ParseFloat(chi.URLParam(r, "lat"), 64)
	if err != nil {
		http_helper.ResponseError(w, http.StatusBadRequest, "Bad request error error", "lat is not valid")
		return
	}
	long, err := strconv.ParseFloat(chi.URLParam(r, "long"), 64)
	if err != nil {
		http_helper.ResponseError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}
	location := &purchase_entity.Location{
		Lat:  lat,
		Long: long,
	}

	query := r.URL.Query()

	params := &purchase_entity.MerchantNearbyQueryParams{
		Id:       query.Get("merchantId"),
		Limit:    5,
		Offset:   0,
		Name:     query.Get("name"),
		Category: query.Get("merchantCategory"),
	}

	if limit := query.Get("limit"); limit != "" {
		params.Limit, _ = strconv.Atoi(limit)
	}

	if offset := query.Get("offset"); offset != "" {
		params.Offset, _ = strconv.Atoi(offset)
	}

	merchantsNearbyResponse, err := c.Service.GetMerchantsNearby(r.Context(), location, params)
	if err != nil {
		http_helper.ResponseError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	http_helper.EncodeJSON(w, http.StatusOK, merchantsNearbyResponse)
}

func (c *PurchaseController) HandleUserEstimateOrder(w http.ResponseWriter, r *http.Request) {
	isAdmin, ok := r.Context().Value(middlewares.ContextIsAdminKey).(bool)
	if !ok {
		http_helper.ResponseError(w, http.StatusUnauthorized, "IsAdmin type assertion failed", "IsAdmin not found in the context")
		return
	}
	if isAdmin {
		http_helper.ResponseError(w, http.StatusUnauthorized, "Unauthorized error", "you're not a user")
		return
	}

	userId, ok := r.Context().Value(middlewares.ContextUserIdKey).(string)
	if !ok {
		http_helper.ResponseError(w, http.StatusUnauthorized, "UserId type assertion failed", "UserId not found in the context")
		return
	}

	payload := &purchase_entity.UserEstimateRequest{}

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

	userEstimateResponse, err := c.Service.EstimateOrder(r.Context(), userId, payload)
	if errors.Is(err, purchase_exception.ErrDistanceTooFar) {
		http_helper.ResponseError(w, http.StatusBadRequest, "Bad request error", err.Error())
		return
	}
	if errors.Is(err, merchant_exception.ErrMerchantIdNotFound) {
		http_helper.ResponseError(w, http.StatusNotFound, "Not found error", err.Error())
		return
	}
	if errors.Is(err, item_exception.ErrItemIdNotFound) {
		http_helper.ResponseError(w, http.StatusNotFound, "Not found error", err.Error())
		return
	}
	if err != nil {
		http_helper.ResponseError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	http_helper.EncodeJSON(w, http.StatusOK, userEstimateResponse)
}

func (c *PurchaseController) HandleUserOrder(w http.ResponseWriter, r *http.Request) {
	isAdmin, ok := r.Context().Value(middlewares.ContextIsAdminKey).(bool)
	if !ok {
		http_helper.ResponseError(w, http.StatusUnauthorized, "IsAdmin type assertion failed", "IsAdmin not found in the context")
		return
	}
	if isAdmin {
		http_helper.ResponseError(w, http.StatusUnauthorized, "Unauthorized error", "you're not a user")
		return
	}

	userId, ok := r.Context().Value(middlewares.ContextUserIdKey).(string)
	if !ok {
		http_helper.ResponseError(w, http.StatusUnauthorized, "UserId type assertion failed", "UserId not found in the context")
		return
	}

	payload := &purchase_entity.UserOrderRequest{}

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

	userOrderResponse, err := c.Service.CreateOrder(r.Context(), userId, payload)
	if errors.Is(err, purchase_exception.ErrEstimateIdNotFound) {
		http_helper.ResponseError(w, http.StatusNotFound, "Not found error", err.Error())
		return
	}
	if err != nil {
		http_helper.ResponseError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	http_helper.EncodeJSON(w, http.StatusCreated, userOrderResponse)
}

func (c *PurchaseController) HandleGetUserOrders(w http.ResponseWriter, r *http.Request) {
	isAdmin, ok := r.Context().Value(middlewares.ContextIsAdminKey).(bool)
	if !ok {
		http_helper.ResponseError(w, http.StatusUnauthorized, "IsAdmin type assertion failed", "IsAdmin not found in the context")
		return
	}
	if isAdmin {
		http_helper.ResponseError(w, http.StatusUnauthorized, "Unauthorized error", "you're not a user")
		return
	}

	userId, ok := r.Context().Value(middlewares.ContextUserIdKey).(string)
	if !ok {
		http_helper.ResponseError(w, http.StatusUnauthorized, "UserId type assertion failed", "UserId not found in the context")
		return
	}

	query := r.URL.Query()

	params := &purchase_entity.OrderQueryParams{
		MerchantId: query.Get("merchantId"),
		Limit:      5,
		Offset:     0,
		Name:       query.Get("name"),
		Category:   query.Get("merchantCategory"),
	}

	if limit := query.Get("limit"); limit != "" {
		params.Limit, _ = strconv.Atoi(limit)
	}

	if offset := query.Get("offset"); offset != "" {
		params.Offset, _ = strconv.Atoi(offset)
	}

	userOrdersResponse, err := c.Service.GetUserOrders(r.Context(), userId, params)
	if err != nil {
		http_helper.ResponseError(w, http.StatusInternalServerError, "Internal server error", err.Error())
		return
	}

	http_helper.EncodeJSON(w, http.StatusOK, userOrdersResponse)
}
