package repositories

import (
	"context"
	"fmt"

	merchant_entity "github.com/danzBraham/beli-mang/internal/entities/merchant"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MerchantRepository interface {
	CreateMerchant(ctx context.Context, merchant *merchant_entity.Merchant) error
}

type MerchantRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewMerchantRepository(db *pgxpool.Pool) MerchantRepository {
	return &MerchantRepositoryImpl{DB: db}
}

func (r *MerchantRepositoryImpl) CreateMerchant(ctx context.Context, merchant *merchant_entity.Merchant) error {
	location := fmt.Sprintf("SRID=4326;POINT(%v %v)", merchant.Location.Long, merchant.Location.Lat)
	query := `INSERT INTO merchants (id, name, category, image_url, location, user_id)
						VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.DB.Exec(ctx, query, &merchant.Id, &merchant.Name, &merchant.Category, &merchant.ImageURL, location, &merchant.UserId)
	if err != nil {
		return err
	}
	return nil
}
