package services

import (
	"context"

	item_entity "github.com/danzBraham/beli-mang/internal/entities/item"
	merchant_entity "github.com/danzBraham/beli-mang/internal/entities/merchant"
	purchase_entity "github.com/danzBraham/beli-mang/internal/entities/purchase"
	item_exception "github.com/danzBraham/beli-mang/internal/exceptions/item"
	merchant_exception "github.com/danzBraham/beli-mang/internal/exceptions/merchant"
	"github.com/danzBraham/beli-mang/internal/repositories"
	"github.com/oklog/ulid/v2"
)

type PurchaseService interface {
	GetMerchantsNearby(ctx context.Context, location *merchant_entity.Location, params *purchase_entity.MerchantNearbyQueryParams) (*purchase_entity.GetMerchantsNearbyResponse, error)
	EstimateOrder(ctx context.Context, userId string, payload *purchase_entity.UserEstimateRequest) (*purchase_entity.UserEstimateResponse, error)
}

type PurchaseServiceImpl struct {
	PurchaseRepository repositories.PurchaseRepository
	MerchantRepository repositories.MerchantRepository
	ItemRepository     repositories.ItemRepository
}

func NewPurchaseService(
	purchaseRepository repositories.PurchaseRepository,
	merchantRepository repositories.MerchantRepository,
	itemRepository repositories.ItemRepository,
) PurchaseService {
	return &PurchaseServiceImpl{
		PurchaseRepository: purchaseRepository,
		MerchantRepository: merchantRepository,
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

func (s *PurchaseServiceImpl) EstimateOrder(ctx context.Context, userId string, payload *purchase_entity.UserEstimateRequest) (*purchase_entity.UserEstimateResponse, error) {
	estimateOrder := &purchase_entity.EstimateOrder{
		Id:           ulid.Make().String(),
		UserLocation: payload.UserLocation,
	}
	orderMerchants := []*purchase_entity.OrderMerchant{}
	orderItems := []*purchase_entity.OrderItem{}

	for _, order := range payload.Orders {
		isMerchantIdExists, err := s.MerchantRepository.VerifyId(ctx, order.MerchantId)
		if err != nil {
			return nil, err
		}
		if !isMerchantIdExists {
			return nil, merchant_exception.ErrMerchantIdNotFound
		}

		orderMerchant := &purchase_entity.OrderMerchant{
			Id:              ulid.Make().String(),
			MerchantId:      order.MerchantId,
			IsStartingPoint: order.IsStartingPoint,
			EstimateId:      estimateOrder.Id,
		}

		orderMerchants = append(orderMerchants, orderMerchant)

		for _, item := range order.Items {
			isItemIdExists, err := s.ItemRepository.VerifyId(ctx, item.Id)
			if err != nil {
				return nil, err
			}
			if !isItemIdExists {
				return nil, item_exception.ErrItemIdNotFound
			}

			orderItems = append(orderItems, &purchase_entity.OrderItem{
				Id:              ulid.Make().String(),
				ItemId:          item.Id,
				Quantity:        item.Quantity,
				OrderMerchantId: orderMerchant.Id,
			})
		}
	}

	estimateOrder, err := s.PurchaseRepository.CreateEstimateOrder(ctx, estimateOrder, orderMerchants, orderItems)
	if err != nil {
		return nil, err
	}

	return &purchase_entity.UserEstimateResponse{
		TotalPrice:      estimateOrder.TotalPrice,
		DeliveryTime:    estimateOrder.EstimatedDeliveryTime,
		EstimateOrderId: estimateOrder.Id,
	}, nil
}
