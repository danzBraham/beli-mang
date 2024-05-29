package services

import (
	"context"

	merchant_entity "github.com/danzBraham/beli-mang/internal/entities/merchant"
	"github.com/danzBraham/beli-mang/internal/repositories"
	"github.com/oklog/ulid/v2"
)

type MerchantService interface {
	CreateMerchant(ctx context.Context, userId string, payload *merchant_entity.AddMerchantRequest) (*merchant_entity.AddMerchantResponse, error)
}

type MerchantServiceImpl struct {
	Repository repositories.MerchantRepository
}

func NewMerchantService(repostiory repositories.MerchantRepository) MerchantService {
	return &MerchantServiceImpl{Repository: repostiory}
}

func (s *MerchantServiceImpl) CreateMerchant(ctx context.Context, userId string, payload *merchant_entity.AddMerchantRequest) (*merchant_entity.AddMerchantResponse, error) {
	merchant := &merchant_entity.Merchant{
		Id:       ulid.Make().String(),
		Name:     payload.Name,
		Category: payload.Category,
		ImageURL: payload.ImageURL,
		Location: payload.Location,
		UserId:   userId,
	}

	err := s.Repository.CreateMerchant(ctx, merchant)
	if err != nil {
		return nil, err
	}

	return &merchant_entity.AddMerchantResponse{
		Id: merchant.Id,
	}, nil
}
