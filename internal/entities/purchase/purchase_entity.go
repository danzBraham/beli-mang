package purchase_entity

import (
	item_entity "github.com/danzBraham/beli-mang/internal/entities/item"
	merchant_entity "github.com/danzBraham/beli-mang/internal/entities/merchant"
)

type MerchantNearbyQueryParams struct {
	Id       string
	Limit    int
	Offset   int
	Name     string
	Category string
}

type GetMerchantsNearby struct {
	Merchant *merchant_entity.GetMerchant `json:"merchant"`
	Items    []*item_entity.GetItem       `json:"items"`
}

type GetMerchantsNearbyResponse struct {
	Data []*GetMerchantsNearby `json:"data"`
}

type Location struct {
	Lat  float64 `json:"lat" validate:"required"`
	Long float64 `json:"long" validate:"required"`
}

type Item struct {
	Id       string `json:"itemId" validate:"required"`
	Quantity int    `json:"quantity" validate:"required"`
}

type Order struct {
	MerchantId      string `json:"merchantId" validate:"required"`
	IsStartingPoint bool   `json:"isStartingPoint" validate:"required"`
	Items           []Item `json:"items"`
}

type UserEstimateRequest struct {
	UserLocation Location `json:"userLocation"`
	Orders       []Order  `json:"orders" validate:"onestartingpoint"`
}

type UserEstimateResponse struct {
	TotalPrice      int    `json:"totalPrice"`
	DeliveryTime    int    `json:"estimatedDeliveryTimeInMinutes"`
	EstimateOrderId string `json:"calculatedEstimateId"`
}

type EstimateOrder struct {
	Id                    string
	UserLocation          Location
	TotalPrice            int
	EstimatedDeliveryTime int
	CreatedAt             string
	UpdatedAt             string
}

type OrderMerchant struct {
	Id                 string
	MerchantId         string
	TotalMerchantPrice int
	IsStartingPoint    bool
	EstimateId         string
	CreatedAt          string
	UpdatedAt          string
}

type OrderItem struct {
	Id              string
	ItemId          string
	Quantity        int
	TotalItemPrice  int
	OrderMerchantId string
	CreatedAt       string
	UpdatedAt       string
}
