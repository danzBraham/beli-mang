package repositories

import (
	"context"
	"strconv"
	"time"

	merchant_entity "github.com/danzBraham/beli-mang/internal/entities/merchant"
	purchase_entity "github.com/danzBraham/beli-mang/internal/entities/purchase"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PurchaseRepository interface {
	GetMerchantsNearby(ctx context.Context, location *merchant_entity.Location, params *purchase_entity.MerchantNearbyQueryParams) ([]*merchant_entity.Merchant, error)
}

type PurchaseRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewPurchaseRepository(db *pgxpool.Pool) PurchaseRepository {
	return &PurchaseRepositoryImpl{DB: db}
}

func (r *PurchaseRepositoryImpl) GetMerchantsNearby(ctx context.Context, location *merchant_entity.Location, params *purchase_entity.MerchantNearbyQueryParams) ([]*merchant_entity.Merchant, error) {
	query := `SELECT id, name, category, image_url, 
							ST_Y(location::geometry) AS latitude,
							ST_X(location::geometry) AS longitude,
							user_id, created_at, updated_at
						FROM merchants
						WHERE 1 = 1
						ORDER BY location <-> ST_SetSRID(ST_MakePoint($1, $2), 4326)`
	args := []interface{}{}
	args = append(args, location.Long, location.Lat)
	argId := 3

	if params.Id != "" {
		query += ` AND id = $` + strconv.Itoa(argId)
		args = append(args, params.Id)
		argId++
	}

	if params.Name != "" {
		query += ` AND name ILIKE $` + strconv.Itoa(argId)
		args = append(args, "%"+params.Name+"%")
		argId++
	}

	validCategories := map[string]bool{
		merchant_entity.SmallRestaurant:       true,
		merchant_entity.MediumRestaurant:      true,
		merchant_entity.LargeRestaurant:       true,
		merchant_entity.MerchandiseRestaurant: true,
		merchant_entity.BoothKiosk:            true,
		merchant_entity.ConvenienceStore:      true,
	}

	if params.Category != "" {
		if !validCategories[params.Category] {
			return []*merchant_entity.Merchant{}, nil
		}
		query += ` AND category = $` + strconv.Itoa(argId)
		args = append(args, params.Category)
		argId++
	}

	query += ` LIMIT $` + strconv.Itoa(argId) + ` OFFSET $` + strconv.Itoa(argId+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	merchants := []*merchant_entity.Merchant{}
	for rows.Next() {
		var merchant merchant_entity.Merchant
		var timeCreated, timeUpdated time.Time
		err := rows.Scan(
			&merchant.Id,
			&merchant.Name,
			&merchant.Category,
			&merchant.ImageURL,
			&merchant.Location.Lat,
			&merchant.Location.Long,
			&merchant.UserId,
			&timeCreated,
			&timeUpdated,
		)
		if err != nil {
			return nil, err
		}
		merchant.CreatedAt = timeCreated.Format(time.RFC3339)
		merchant.UpdatedAt = timeUpdated.Format(time.RFC3339)
		merchants = append(merchants, &merchant)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return merchants, nil
}
