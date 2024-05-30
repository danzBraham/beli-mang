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
