package services

import (
	"context"

	item_entity "github.com/danzBraham/beli-mang/internal/entities/item"
	purchase_entity "github.com/danzBraham/beli-mang/internal/entities/purchase"
	item_exception "github.com/danzBraham/beli-mang/internal/exceptions/item"
	merchant_exception "github.com/danzBraham/beli-mang/internal/exceptions/merchant"
	purchase_exception "github.com/danzBraham/beli-mang/internal/exceptions/purchase"
	"github.com/danzBraham/beli-mang/internal/repositories"
	"github.com/oklog/ulid/v2"
)

type PurchaseService interface {
	GetMerchantsNearby(ctx context.Context, location *purchase_entity.Location, params *purchase_entity.MerchantNearbyQueryParams) (*purchase_entity.GetMerchantsNearbyResponse, error)
	EstimateOrder(ctx context.Context, userId string, payload *purchase_entity.UserEstimateRequest) (*purchase_entity.UserEstimateResponse, error)
	CreateOrder(ctx context.Context, userId string, payload *purchase_entity.UserOrderRequest) (*purchase_entity.UserOrderResponse, error)
	GetUserOrders(ctx context.Context, userId string, params *purchase_entity.OrderQueryParams) ([]*purchase_entity.GetUserOrder, error)
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

func (s *PurchaseServiceImpl) GetMerchantsNearby(ctx context.Context, location *purchase_entity.Location, params *purchase_entity.MerchantNearbyQueryParams) (*purchase_entity.GetMerchantsNearbyResponse, error) {
	merchantsNearby, err := s.PurchaseRepository.GetMerchantsNearby(ctx, location, params)
	if err != nil {
		return nil, err
	}

	getMerchants := []*purchase_entity.GetMerchantsNearby{}
	for _, merchant := range merchantsNearby {
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
			Merchant: merchant,
			Items:    getItems,
		})
	}

	countMerchants, err := s.MerchantRepository.CountMerhcants(ctx)
	if err != nil {
		return nil, err
	}

	return &purchase_entity.GetMerchantsNearbyResponse{
		Data: getMerchants,
		Meta: &purchase_entity.Meta{
			Limit:  params.Limit,
			Offset: params.Offset,
			Total:  countMerchants,
		},
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

func (s *PurchaseServiceImpl) CreateOrder(ctx context.Context, userId string, payload *purchase_entity.UserOrderRequest) (*purchase_entity.UserOrderResponse, error) {
	isEstimateIdExists, err := s.PurchaseRepository.VerifyEstimateId(ctx, payload.EstimateId)
	if err != nil {
		return nil, err
	}
	if !isEstimateIdExists {
		return nil, purchase_exception.ErrEstimateIdNotFound
	}

	userOrder := &purchase_entity.UserOrder{
		Id:         ulid.Make().String(),
		EstimateId: payload.EstimateId,
		UserId:     userId,
	}

	err = s.PurchaseRepository.CreateOrder(ctx, userOrder)
	if err != nil {
		return nil, err
	}

	return &purchase_entity.UserOrderResponse{
		OrderId: userOrder.Id,
	}, nil
}

func (s *PurchaseServiceImpl) GetUserOrders(ctx context.Context, userId string, params *purchase_entity.OrderQueryParams) ([]*purchase_entity.GetUserOrder, error) {
	getOrders, err := s.PurchaseRepository.GetOrders(ctx, userId, params)
	if err != nil {
		return nil, err
	}

	return getOrders, nil
}
