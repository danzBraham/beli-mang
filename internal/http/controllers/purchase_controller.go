package controllers

import (
	"net/http"
	"strconv"

	merchant_entity "github.com/danzBraham/beli-mang/internal/entities/merchant"
	purchase_entity "github.com/danzBraham/beli-mang/internal/entities/purchase"
	http_helper "github.com/danzBraham/beli-mang/internal/helpers/http"
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
	location := &merchant_entity.Location{
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
