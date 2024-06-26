package repositories

import (
	"context"
	"errors"
	"strconv"
	"time"

	item_entity "github.com/danzBraham/beli-mang/internal/entities/item"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ItemRepository interface {
	VerifyId(ctx context.Context, itemId string) (bool, error)
	CreateItem(ctx context.Context, item *item_entity.Item) error
	GetItems(ctx context.Context, params *item_entity.ItemQueryParams) ([]*item_entity.Item, error)
	GetItemsByMerchantId(ctx context.Context, merchantId string) ([]*item_entity.Item, error)
	CountItems(ctx context.Context) (count int, err error)
}

type ItemRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewItemRepository(db *pgxpool.Pool) ItemRepository {
	return &ItemRepositoryImpl{DB: db}
}

func (r *ItemRepositoryImpl) VerifyId(ctx context.Context, itemId string) (bool, error) {
	var one int
	query := `SELECT 1 FROM items WHERE id = $1`
	err := r.DB.QueryRow(ctx, query, itemId).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *ItemRepositoryImpl) CreateItem(ctx context.Context, item *item_entity.Item) error {
	query := `INSERT INTO items (id, name, category, price, image_url, merchant_id)
						VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.DB.Exec(ctx, query, &item.Id, &item.Name, &item.Category, &item.Price, &item.ImageURL, &item.MerchantId)
	if err != nil {
		return err
	}
	return nil
}

func (r *ItemRepositoryImpl) GetItems(ctx context.Context, params *item_entity.ItemQueryParams) ([]*item_entity.Item, error) {
	query := `SELECT id, name, category, price, image_url, merchant_id, created_at, updated_at
						FROM items 
						WHERE 1 = 1`
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

	validCategories := map[string]bool{
		item_entity.Beverage:   true,
		item_entity.Food:       true,
		item_entity.Snack:      true,
		item_entity.Condiments: true,
		item_entity.Additions:  true,
	}

	if params.Category != "" {
		if !validCategories[params.Category] {
			return []*item_entity.Item{}, nil
		}
		query += ` AND category = $` + strconv.Itoa(argId)
		args = append(args, params.Category)
		argId++
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

	items := []*item_entity.Item{}
	for rows.Next() {
		var item item_entity.Item
		var timeCreated, timeUpdated time.Time
		err := rows.Scan(
			&item.Id,
			&item.Name,
			&item.Category,
			&item.Price,
			&item.ImageURL,
			&item.MerchantId,
			&timeCreated,
			&timeUpdated,
		)
		if err != nil {
			return nil, err
		}
		item.CreatedAt = timeCreated.Format(time.RFC3339)
		item.UpdatedAt = timeUpdated.Format(time.RFC3339)
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *ItemRepositoryImpl) GetItemsByMerchantId(ctx context.Context, merchantId string) ([]*item_entity.Item, error) {
	query := `SELECT id, name, category, price, image_url, merchant_id, created_at, updated_at
						FROM items 
						WHERE merchant_id = $1`
	rows, err := r.DB.Query(ctx, query, merchantId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*item_entity.Item{}
	for rows.Next() {
		var item item_entity.Item
		var timeCreated, timeUpdated time.Time
		err := rows.Scan(
			&item.Id,
			&item.Name,
			&item.Category,
			&item.Price,
			&item.ImageURL,
			&item.MerchantId,
			&timeCreated,
			&timeUpdated,
		)
		if err != nil {
			return nil, err
		}
		item.CreatedAt = timeCreated.Format(time.RFC3339)
		item.UpdatedAt = timeUpdated.Format(time.RFC3339)
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *ItemRepositoryImpl) CountItems(ctx context.Context) (count int, err error) {
	query := `SELECT COUNT(1) FROM items`
	err = r.DB.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
