package services

import (
	"context"

	item_entity "github.com/danzBraham/beli-mang/internal/entities/item"
	merchant_entity "github.com/danzBraham/beli-mang/internal/entities/merchant"
	purchase_entity "github.com/danzBraham/beli-mang/internal/entities/purchase"
	"github.com/danzBraham/beli-mang/internal/repositories"
)

type PurchaseService interface {
	GetMerchantsNearby(ctx context.Context, location *merchant_entity.Location, params *purchase_entity.MerchantNearbyQueryParams) (*purchase_entity.GetMerchantsNearbyResponse, error)
}

type PurchaseServiceImpl struct {
	PurchaseRepository repositories.PurchaseRepository
	ItemRepository     repositories.ItemRepository
}

func NewPurchaseService(
	purchaseRepository repositories.PurchaseRepository,
	itemRepository repositories.ItemRepository,
) PurchaseService {
	return &PurchaseServiceImpl{
		PurchaseRepository: purchaseRepository,
		ItemRepository:     itemRepository,
	}
}

func (s *PurchaseServiceImpl) GetMerchantsNearby(ctx context.Context, location *merchant_entity.Location, params *purchase_entity.MerchantNearbyQueryParams) (*purchase_entity.GetMerchantsNearbyResponse, error) {
	merchants, err := s.PurchaseRepository.GetMerchantsNearby(ctx, location, params)
	if err != nil {
		return nil, err
	}

	getMerchants := []*purchase_entity.GetMerchantsNearby{}
	for _, merchant := range merchants {
		items, err := s.ItemRepository.GetItemsByMerchantId(ctx, merchant.Id)
		if err != nil {
			return nil, err
		}

		getItems := []*item_entity.GetItem{}
		for _, item := range items {
			getItems = append(getItems, &item_entity.GetItem{
				Id:        item.Id,
				Name:      item.Name,
				Category:  item.Category,
				Price:     item.Price,
				ImageURL:  item.ImageURL,
				CreatedAt: item.CreatedAt,
			})
		}

		getMerchants = append(getMerchants, &purchase_entity.GetMerchantsNearby{
			Merchant: &merchant_entity.GetMerchant{
				Id:       merchant.Id,
				Name:     merchant.Name,
				Category: merchant.Category,
				ImageURL: merchant.ImageURL,
				Location: merchant_entity.Location{
					Lat:  merchant.Location.Lat,
					Long: merchant.Location.Long,
				},
				CreatedAt: merchant.CreatedAt,
			},
			Items: getItems,
		})
	}

	return &purchase_entity.GetMerchantsNearbyResponse{
		Data: getMerchants,
	}, nil
}
