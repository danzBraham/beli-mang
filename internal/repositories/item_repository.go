package repositories

import (
	"context"

	item_entity "github.com/danzBraham/beli-mang/internal/entities/item"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ItemRepository interface {
	CreateItem(ctx context.Context, item *item_entity.Item) error
}

type ItemRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewItemRepository(db *pgxpool.Pool) ItemRepository {
	return &ItemRepositoryImpl{DB: db}
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
