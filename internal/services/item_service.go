package services

import (
	"context"

	item_entity "github.com/danzBraham/beli-mang/internal/entities/item"
	merchant_exception "github.com/danzBraham/beli-mang/internal/exceptions/merchant"
	"github.com/danzBraham/beli-mang/internal/repositories"
	"github.com/oklog/ulid/v2"
)

type ItemService interface {
	CreateItem(ctx context.Context, merchantId string, payload *item_entity.AddItemRequest) (*item_entity.AddItemResponse, error)
	GetItems(ctx context.Context, merchantId string, params *item_entity.ItemQueryParams) (*item_entity.GetItemResponse, error)
}

type ItemServiceImpl struct {
	ItemRepository     repositories.ItemRepository
	MerchantRepository repositories.MerchantRepository
}

func NewItemService(itemRepostiory repositories.ItemRepository, merchantRepository repositories.MerchantRepository) ItemService {
	return &ItemServiceImpl{
		ItemRepository:     itemRepostiory,
		MerchantRepository: merchantRepository,
	}
}

func (s *ItemServiceImpl) CreateItem(ctx context.Context, merchantId string, payload *item_entity.AddItemRequest) (*item_entity.AddItemResponse, error) {
	isMerchantIdExists, err := s.MerchantRepository.VerifyId(ctx, merchantId)
	if err != nil {
		return nil, err
	}
	if !isMerchantIdExists {
		return nil, merchant_exception.ErrMerchantIdNotFound
	}

	item := &item_entity.Item{
		Id:         ulid.Make().String(),
		Name:       payload.Name,
		Category:   payload.Category,
		Price:      payload.Price,
		ImageURL:   payload.ImageURL,
		MerchantId: merchantId,
	}

	err = s.ItemRepository.CreateItem(ctx, item)
	if err != nil {
		return nil, err
	}

	return &item_entity.AddItemResponse{
		Id: item.Id,
	}, nil
}

func (s *ItemServiceImpl) GetItems(ctx context.Context, merchantId string, params *item_entity.ItemQueryParams) (*item_entity.GetItemResponse, error) {
	isMerchantIdExists, err := s.MerchantRepository.VerifyId(ctx, merchantId)
	if err != nil {
		return nil, err
	}
	if !isMerchantIdExists {
		return nil, merchant_exception.ErrMerchantIdNotFound
	}

	items, err := s.ItemRepository.GetItems(ctx, params)
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

	totalItems, err := s.ItemRepository.CountItems(ctx)
	if err != nil {
		return nil, err
	}

	return &item_entity.GetItemResponse{
		Data: getItems,
		Meta: item_entity.Meta{
			Limit:  params.Limit,
			Offset: params.Offset,
			Total:  totalItems,
		},
	}, nil
}
