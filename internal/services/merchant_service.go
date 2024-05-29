package services

import (
	"context"

	merchant_entity "github.com/danzBraham/beli-mang/internal/entities/merchant"
	"github.com/danzBraham/beli-mang/internal/repositories"
	"github.com/oklog/ulid/v2"
)

type MerchantService interface {
	CreateMerchant(ctx context.Context, userId string, payload *merchant_entity.AddMerchantRequest) (*merchant_entity.AddMerchantResponse, error)
	GetMerchants(ctx context.Context, params *merchant_entity.MerchantQueryParams) (*merchant_entity.GetMerchantResponse, error)
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

func (s *MerchantServiceImpl) GetMerchants(ctx context.Context, params *merchant_entity.MerchantQueryParams) (*merchant_entity.GetMerchantResponse, error) {
	merchants, err := s.Repository.GetMerchants(ctx, params)
	if err != nil {
		return nil, err
	}

	getMerchants := []*merchant_entity.GetMerchant{}
	for _, merchant := range merchants {
		getMerchants = append(getMerchants, &merchant_entity.GetMerchant{
			Id:       merchant.Id,
			Name:     merchant.Name,
			Category: merchant.Category,
			ImageURL: merchant.ImageURL,
			Location: merchant_entity.Location{
				Lat:  merchant.Location.Lat,
				Long: merchant.Location.Long,
			},
			CreatedAt: merchant.CreatedAt,
		})
	}

	countMerchants, err := s.Repository.CountMerhcants(ctx)
	if err != nil {
		return nil, err
	}

	return &merchant_entity.GetMerchantResponse{
		Data: getMerchants,
		Meta: merchant_entity.Meta{
			Limit:  params.Limit,
			Offset: params.Offset,
			Total:  countMerchants,
		},
	}, nil
}
