package repositories

import (
	"database/sql"
	"time"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

type OrderRepository struct {
	Db     *goqu.Database
	Logger *utils.StructuredLogger
}

const ORDER_DB = "orders"
const ORDER_PRODUCT_DB = "order_products"

func InitOrderRepository(db *goqu.Database, logger *utils.StructuredLogger) (*OrderRepository, error) {
	return &OrderRepository{
		Db:     db,
		Logger: logger,
	}, nil
}

// ListOrders returns orders ordered newest-first with pagination. Soft-deleted
// orders (deleted_at IS NOT NULL) are excluded.
func (orderRepositoryPtr *OrderRepository) ListOrders(limit int, offset int, exec ...utils.Executor) ([]models.Order, int64, error) {
	db := utils.ResolveExecutor(orderRepositoryPtr.Db, exec)

	var orders []models.Order

	err := db.
		From(ORDER_DB).
		Where(goqu.C("deleted_at").IsNull()).
		Order(goqu.I("created_at").Desc()).
		Limit(uint(limit)).
		Offset(uint(offset)).
		Select(
			"id",
			"customer_id",
			"total_bill",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStructs(&orders)

	if err != nil {
		return nil, 0, err
	}

	var total int64

	_, err = db.
		From(ORDER_DB).
		Where(goqu.C("deleted_at").IsNull()).
		Select(goqu.COUNT("*")).
		ScanVal(&total)

	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// ListOrdersInRange returns orders created within [start, end] (inclusive),
// ordered by created_at ASC, with pagination. Soft-deleted orders
// (deleted_at IS NOT NULL) are excluded. ASC ordering keeps time-proximity
// clustering simple for the dedup job.
func (orderRepositoryPtr *OrderRepository) ListOrdersInRange(start time.Time, end time.Time, limit int, offset int, exec ...utils.Executor) ([]models.Order, int64, error) {
	db := utils.ResolveExecutor(orderRepositoryPtr.Db, exec)

	var orders []models.Order

	err := db.
		From(ORDER_DB).
		Where(
			goqu.C("deleted_at").IsNull(),
			goqu.C("created_at").Gte(start),
			goqu.C("created_at").Lte(end),
		).
		Order(goqu.I("created_at").Asc()).
		Limit(uint(limit)).
		Offset(uint(offset)).
		Select(
			"id",
			"customer_id",
			"total_bill",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		ScanStructs(&orders)

	if err != nil {
		return nil, 0, err
	}

	var total int64

	_, err = db.
		From(ORDER_DB).
		Where(
			goqu.C("deleted_at").IsNull(),
			goqu.C("created_at").Gte(start),
			goqu.C("created_at").Lte(end),
		).
		Select(goqu.COUNT("*")).
		ScanVal(&total)

	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// DeleteOrder soft-deletes an order by setting deleted_at.
func (orderRepositoryPtr *OrderRepository) DeleteOrder(id uuid.UUID, exec ...utils.Executor) error {
	db := utils.ResolveExecutor(orderRepositoryPtr.Db, exec)

	res, err := db.Update(ORDER_DB).Set(
		goqu.Record{
			"deleted_at": time.Now(),
		},
	).Where(
		goqu.C("id").Eq(id),
		goqu.C("deleted_at").IsNull(),
	).Executor().Exec()

	if err != nil {
		return err
	}

	if affected, _ := res.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteOrders soft-deletes the given orders in a single statement by setting
// deleted_at. Orders already soft-deleted (deleted_at IS NOT NULL) are skipped.
// It returns the number of rows actually affected, so callers can report how
// many duplicates were really removed (idempotent across runs).
func (orderRepositoryPtr *OrderRepository) DeleteOrders(ids []uuid.UUID, exec ...utils.Executor) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	db := utils.ResolveExecutor(orderRepositoryPtr.Db, exec)

	vals := make([]interface{}, len(ids))
	for i, id := range ids {
		vals[i] = id
	}

	res, err := db.Update(ORDER_DB).
		Set(goqu.Record{"deleted_at": time.Now()}).
		Where(
			goqu.C("id").In(vals...),
			goqu.C("deleted_at").IsNull(),
		).
		Executor().
		Exec()
	if err != nil {
		return 0, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}
