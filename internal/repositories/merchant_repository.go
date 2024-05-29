package repositories

import (
	"context"
	"fmt"
	"strconv"
	"time"

	merchant_entity "github.com/danzBraham/beli-mang/internal/entities/merchant"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MerchantRepository interface {
	CreateMerchant(ctx context.Context, merchant *merchant_entity.Merchant) error
	GetMerchants(ctx context.Context, params *merchant_entity.MerchantQueryParams) ([]*merchant_entity.Merchant, error)
	CountMerhcants(ctx context.Context) (count int, err error)
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

func (r *MerchantRepositoryImpl) GetMerchants(ctx context.Context, params *merchant_entity.MerchantQueryParams) ([]*merchant_entity.Merchant, error) {
	query := `SELECT id, name, category, image_url, 
							ST_Y(location::geometry) AS latitude,
							ST_X(location::geometry) AS longitude,
							user_id, created_at
						FROM merchants WHERE 1 = 1`
	args := []interface{}{}
	argId := 1

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

	switch params.Category {
	case merchant_entity.SmallRestaurant:
		query += ` AND category = '` + merchant_entity.SmallRestaurant + `'`
	case merchant_entity.MediumRestaurant:
		query += ` AND category = '` + merchant_entity.MediumRestaurant + `'`
	case merchant_entity.LargeRestaurant:
		query += ` AND category = '` + merchant_entity.LargeRestaurant + `'`
	case merchant_entity.MerchandiseRestaurant:
		query += ` AND category = '` + merchant_entity.MerchandiseRestaurant + `'`
	case merchant_entity.BoothKiosk:
		query += ` AND category = '` + merchant_entity.BoothKiosk + `'`
	case merchant_entity.ConvenienceStore:
		query += ` AND category = '` + merchant_entity.ConvenienceStore + `'`
	default:
		query += " AND category = ''"
	}

	switch params.CreatedAt {
	case "asc":
		query += ` ORDER BY created_at ASC`
	case "desc":
		query += ` ORDER BY created_at DESC`
	default:
		query += ` ORDER BY created_at DESC`
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
		var timeCreated time.Time
		err := rows.Scan(
			&merchant.Id,
			&merchant.Name,
			&merchant.Category,
			&merchant.ImageURL,
			&merchant.Location.Lat,
			&merchant.Location.Long,
			&merchant.UserId,
			&timeCreated,
		)
		if err != nil {
			return nil, err
		}
		merchant.CreatedAt = timeCreated.Format(time.RFC3339)
		merchants = append(merchants, &merchant)
	}

	return merchants, nil
}

func (r *MerchantRepositoryImpl) CountMerhcants(ctx context.Context) (count int, err error) {
	query := `SELECT COUNT(1) FROM merchants`
	err = r.DB.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
