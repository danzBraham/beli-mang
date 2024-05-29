package services

import (
	"context"

	item_entity "github.com/danzBraham/beli-mang/internal/entities/item"
	"github.com/danzBraham/beli-mang/internal/repositories"
	"github.com/oklog/ulid/v2"
)

type ItemService interface {
	CreateItem(ctx context.Context, merchantId string, payload *item_entity.AddItemRequest) (*item_entity.AddItemResponse, error)
}

type ItemServiceImpl struct {
	Repository repositories.ItemRepository
}

func NewItemService(repostiory repositories.ItemRepository) ItemService {
	return &ItemServiceImpl{Repository: repostiory}
}

func (s *ItemServiceImpl) CreateItem(ctx context.Context, merchantId string, payload *item_entity.AddItemRequest) (*item_entity.AddItemResponse, error) {
	item := &item_entity.Item{
		Id:         ulid.Make().String(),
		Name:       payload.Name,
		Category:   payload.Category,
		Price:      payload.Price,
		ImageURL:   payload.ImageURL,
		MerchantId: merchantId,
	}

	err := s.Repository.CreateItem(ctx, item)
	if err != nil {
		return nil, err
	}

	return &item_entity.AddItemResponse{
		Id: item.Id,
	}, nil
}
