package repositories

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	merchant_entity "github.com/danzBraham/beli-mang/internal/entities/merchant"
	purchase_entity "github.com/danzBraham/beli-mang/internal/entities/purchase"
	purchase_exception "github.com/danzBraham/beli-mang/internal/exceptions/purchase"
	formula_helper "github.com/danzBraham/beli-mang/internal/helpers/formula"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PurchaseRepository interface {
	GetMerchantsNearby(ctx context.Context, location *purchase_entity.Location, params *purchase_entity.MerchantNearbyQueryParams) ([]*merchant_entity.GetMerchant, error)
	CreateEstimateOrder(ctx context.Context, estimateOrder *purchase_entity.EstimateOrder, orderMerchants []*purchase_entity.OrderMerchant, orderItems []*purchase_entity.OrderItem) (*purchase_entity.EstimateOrder, error)
	CreateOrder(ctx context.Context, userOrder *purchase_entity.UserOrder) error
	VerifyEstimateId(ctx context.Context, estimateId string) (bool, error)
	GetOrders(ctx context.Context, userId string, params *purchase_entity.OrderQueryParams) ([]*purchase_entity.GetUserOrder, error)
}

type PurchaseRepositoryImpl struct {
	DB *pgxpool.Pool
}

func NewPurchaseRepository(db *pgxpool.Pool) PurchaseRepository {
	return &PurchaseRepositoryImpl{DB: db}
}

func (r *PurchaseRepositoryImpl) GetMerchantsNearby(ctx context.Context, location *purchase_entity.Location, params *purchase_entity.MerchantNearbyQueryParams) ([]*merchant_entity.GetMerchant, error) {
	query := `
		WITH user_location AS (
			SELECT ST_SetSRID(ST_MakePoint($1, $2), 4326) AS location
		)
		SELECT 
			m.id, m.name, m.category, m.image_url,
			ST_Y(m.location::geometry) AS latitude, ST_X(m.location::geometry) AS longitude, m.created_at
		FROM merchants m, user_location ul
		WHERE 1 = 1
	`
	args := []interface{}{location.Long, location.Lat}
	argId := len(args) + 1

	if params.Id != "" {
		query += ` AND m.id = $` + strconv.Itoa(argId)
		args = append(args, params.Id)
		argId++
	}

	if params.Name != "" {
		query += ` AND m.name ILIKE $` + strconv.Itoa(argId)
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
			return []*merchant_entity.GetMerchant{}, nil
		}
		query += ` AND m.category = $` + strconv.Itoa(argId)
		args = append(args, params.Category)
		argId++
	}

	query += ` ORDER BY m.location <-> ul.location`

	query += ` LIMIT $` + strconv.Itoa(argId) + ` OFFSET $` + strconv.Itoa(argId+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	merchants := []*merchant_entity.GetMerchant{}
	for rows.Next() {
		var merchant merchant_entity.GetMerchant
		var timeCreated time.Time
		err := rows.Scan(
			&merchant.Id,
			&merchant.Name,
			&merchant.Category,
			&merchant.ImageURL,
			&merchant.Location.Lat,
			&merchant.Location.Long,
			&timeCreated,
		)
		if err != nil {
			return nil, err
		}
		merchant.CreatedAt = timeCreated.Format(time.RFC3339)
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

	const MaxDistance = 3.0 // Max distance in km

	// Collect all points for smallest enclosing circle calculation
	points := []purchase_entity.Location{estimateOrder.UserLocation}

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
		points = append(points, purchase_entity.Location{Long: merchantLong, Lat: merchantLat})
	}

	// Check if any point exceeds the MaxDistance using smallest enclosing circle
	circle := formula_helper.SmallestEnclosingCircle(points)
	if circle.Radius > MaxDistance {
		return nil, purchase_exception.ErrDistanceTooFar
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

		distance := formula_helper.Haversine(estimateOrder.UserLocation.Lat, merchantLat, estimateOrder.UserLocation.Long, merchantLong)
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

func (r *PurchaseRepositoryImpl) CreateOrder(ctx context.Context, userOrder *purchase_entity.UserOrder) error {
	query := `
		INSERT INTO orders (id, estimate_id, user_id)
		VALUES ($1, $2, $3)
	`
	_, err := r.DB.Exec(ctx, query, &userOrder.Id, &userOrder.EstimateId, &userOrder.UserId)
	if err != nil {
		return err
	}
	return nil
}

func (r *PurchaseRepositoryImpl) VerifyEstimateId(ctx context.Context, estimateId string) (bool, error) {
	var one int
	query := `SELECT 1 FROM estimates WHERE id = $1`
	err := r.DB.QueryRow(ctx, query, estimateId).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *PurchaseRepositoryImpl) GetOrders(ctx context.Context, userId string, params *purchase_entity.OrderQueryParams) ([]*purchase_entity.GetUserOrder, error) {
	query := `
		SELECT
			o.id,
			m.id, m.name, m.category, m.image_url, ST_Y(m.location::geometry) AS latitude, ST_X(m.location::geometry) AS longitude, m.created_at,
			i.id, i.name, i.category, i.price, oi.quantity, i.image_url, i.created_at
		FROM orders o
		INNER JOIN order_merchants om ON om.estimate_id = o.estimate_id
		INNER JOIN merchants m ON m.id = om.merchant_id
		INNER JOIN order_items oi ON oi.order_merchant_id = om.id
		INNER JOIN items i ON i.id = oi.item_id
		WHERE o.user_id = $1
	`
	args := []interface{}{}
	args = append(args, userId)
	argId := 2

	if params.MerchantId != "" {
		query += ` AND m.id = $` + strconv.Itoa(argId)
		args = append(args, params.MerchantId)
		argId++
	}

	if params.Name != "" {
		query += ` AND m.name ILIKE $` + strconv.Itoa(argId) + ` OR i.name ILIKE $` + strconv.Itoa(argId)
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
			return []*purchase_entity.GetUserOrder{}, nil
		}
		query += ` AND m.category = $` + strconv.Itoa(argId)
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

	ordersMap := make(map[string]*purchase_entity.GetUserOrder)
	for rows.Next() {
		var (
			orderId, merchantId, merchantName, merchantCategory, merchantImageUrl string
			merchantLat, merchantLong                                             float64
			merchantCreatedAt, itemCreatedAt                                      time.Time
			itemId, itemName, itemCategory, itemImageUrl                          string
			itemPrice, itemQuantity                                               int
		)

		err := rows.Scan(
			&orderId,
			&merchantId, &merchantName, &merchantCategory, &merchantImageUrl, &merchantLat, &merchantLong, &merchantCreatedAt,
			&itemId, &itemName, &itemCategory, &itemPrice, &itemQuantity, &itemImageUrl, &itemCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if _, exists := ordersMap[orderId]; !exists {
			ordersMap[orderId] = &purchase_entity.GetUserOrder{
				OrderId: orderId,
				Orders:  []purchase_entity.GetOrder{},
			}
		}

		order := ordersMap[orderId]
		foundMerchant := false

		for i := range order.Orders {
			if order.Orders[i].Merchant.Id == merchantId {
				order.Orders[i].Items = append(order.Orders[i].Items, purchase_entity.GetItem{
					Id:        itemId,
					Name:      itemName,
					Category:  itemCategory,
					Price:     itemPrice,
					Quantity:  itemQuantity,
					ImageURL:  itemImageUrl,
					CreatedAt: itemCreatedAt.Format(time.RFC3339Nano),
				})
				foundMerchant = true
				break
			}
		}

		if !foundMerchant {
			order.Orders = append(order.Orders, purchase_entity.GetOrder{
				Merchant: purchase_entity.GetMerchant{
					Id:        merchantId,
					Name:      merchantName,
					Category:  merchantCategory,
					ImageURL:  merchantImageUrl,
					Location:  purchase_entity.Location{Lat: merchantLat, Long: merchantLong},
					CreatedAt: merchantCreatedAt.Format(time.RFC3339Nano),
				},
				Items: []purchase_entity.GetItem{
					{
						Id:        itemId,
						Name:      itemName,
						Category:  itemCategory,
						Price:     itemPrice,
						Quantity:  itemQuantity,
						ImageURL:  itemImageUrl,
						CreatedAt: itemCreatedAt.Format(time.RFC3339Nano),
					},
				},
			})
		}
	}

	orders := make([]*purchase_entity.GetUserOrder, 0, len(ordersMap))
	for _, order := range ordersMap {
		orders = append(orders, order)
	}

	return orders, nil
}
