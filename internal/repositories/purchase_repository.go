package repositories

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	merchant_entity "github.com/danzBraham/beli-mang/internal/entities/merchant"
	purchase_entity "github.com/danzBraham/beli-mang/internal/entities/purchase"
	purchase_exception "github.com/danzBraham/beli-mang/internal/exceptions/purchase"
	formula_helper "github.com/danzBraham/beli-mang/internal/helpers/formula"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PurchaseRepository interface {
	GetMerchantsNearby(ctx context.Context, location *merchant_entity.Location, params *purchase_entity.MerchantNearbyQueryParams) ([]*merchant_entity.Merchant, error)
	CreateEstimateOrder(ctx context.Context, estimateOrder *purchase_entity.EstimateOrder, orderMerchants []*purchase_entity.OrderMerchant, orderItems []*purchase_entity.OrderItem) (*purchase_entity.EstimateOrder, error)
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

func (r *PurchaseRepositoryImpl) CreateEstimateOrder(ctx context.Context, estimateOrder *purchase_entity.EstimateOrder, orderMerchants []*purchase_entity.OrderMerchant, orderItems []*purchase_entity.OrderItem) (*purchase_entity.EstimateOrder, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	// Calculate cartesian distance and check if any merhcant is too far
	startLat := estimateOrder.UserLocation.Lat
	startLong := estimateOrder.UserLocation.Long
	const MaxDistance = 3.0 // Max distance in km

	getMerchantLocationQuery := `
		SELECT ST_Y(location::geometry) AS latitude, ST_X(location::geometry) AS longitude
		FROM merchants
		WHERE id = $1
	`

	for _, orderMerchant := range orderMerchants {
		var merchantLat, merchantLong float64
		err := tx.QueryRow(ctx, getMerchantLocationQuery, orderMerchant.MerchantId).Scan(&merchantLat, &merchantLong)
		if err != nil {
			return nil, err
		}

		distance := formula_helper.Haversine(startLat, merchantLat, startLong, merchantLong)
		if distance > MaxDistance {
			return nil, purchase_exception.ErrDistanceTooFar
		}
	}

	// Create order
	createEstimateQuery := `
		INSERT INTO estimates (id, user_location)
		VALUES ($1, $2)
	`
	location := fmt.Sprintf("SRID=4326;POINT(%v %v)", estimateOrder.UserLocation.Long, estimateOrder.UserLocation.Lat)
	_, err = tx.Exec(ctx, createEstimateQuery, estimateOrder.Id, location)
	if err != nil {
		return nil, err
	}

	// Create order merchants
	createOrderMerchantQuery := `
		INSERT INTO order_merchants (id, merchant_id, is_starting_point, estimate_id)
		VALUES ($1, $2, $3, $4)
	`
	for _, orderMerchant := range orderMerchants {
		_, err := tx.Exec(ctx, createOrderMerchantQuery, orderMerchant.Id, orderMerchant.MerchantId, orderMerchant.IsStartingPoint, orderMerchant.EstimateId)
		if err != nil {
			return nil, err
		}
	}

	// Create order items
	createOrderItemQuery := `
		INSERT INTO order_items (id, item_id, quantity, total_item_price, order_merchant_id)
		SELECT $1, $2, $3, $4 * price, $5 FROM items WHERE id = $6;
	`
	for _, orderItem := range orderItems {
		_, err := tx.Exec(ctx, createOrderItemQuery,
			orderItem.Id,
			orderItem.ItemId,
			orderItem.Quantity,
			orderItem.Quantity,
			orderItem.OrderMerchantId,
			orderItem.ItemId,
		)
		if err != nil {
			return nil, err
		}
	}

	// Set total merchant price
	setTotalMerchantPriceQuery := `
		UPDATE order_merchants
		SET total_merchant_price = (
			SELECT SUM(total_item_price) 
			FROM order_items
			WHERE order_merchant_id = $1
		)
		WHERE id = $1
	`
	for _, orderMerchant := range orderMerchants {
		_, err := tx.Exec(ctx, setTotalMerchantPriceQuery, orderMerchant.Id)
		if err != nil {
			return nil, err
		}
	}

	// Set total price
	var orderId string
	var totalPrice int
	setTotalPriceQuery := `
		UPDATE estimates
		SET total_price = (
			SELECT SUM(total_merchant_price) 
			FROM order_merchants
			WHERE estimate_id = $1
		)
		WHERE id = $1
		RETURNING id, total_price
	`
	err = tx.QueryRow(ctx, setTotalPriceQuery, estimateOrder.Id).Scan(&orderId, &totalPrice)
	if err != nil {
		return nil, err
	}

	// Calculate estimated delivery time
	var maxDistance float64
	for _, orderMerchant := range orderMerchants {
		var merchantLat, merchantLong float64
		err := tx.QueryRow(ctx, getMerchantLocationQuery, orderMerchant.MerchantId).Scan(&merchantLat, &merchantLong)
		if err != nil {
			return nil, err
		}

		distance := formula_helper.Haversine(startLat, merchantLat, startLong, merchantLong)
		if distance > maxDistance {
			maxDistance = distance
		}
	}

	averageSpeed := 40.0
	estimatedDeliveryTime := int(math.Round((maxDistance / averageSpeed) * 60)) // Time in minutes, rounded to nearest integer

	// Update the order with estimated delivery time
	updateEstimatedDeliveryTime := `
		UPDATE estimates
		SET estimated_delivery_time = $1
		WHERE id = $2
	`
	_, err = tx.Exec(ctx, updateEstimatedDeliveryTime, estimatedDeliveryTime, estimateOrder.Id)
	if err != nil {
		return nil, err
	}

	return &purchase_entity.EstimateOrder{
		Id:                    orderId,
		TotalPrice:            totalPrice,
		EstimatedDeliveryTime: estimatedDeliveryTime,
	}, nil
}
