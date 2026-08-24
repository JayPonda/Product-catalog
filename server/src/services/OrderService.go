package services

import (
	"database/sql"
	"time"

	"github.com/JayPonda/Product-catalog/server/src/repositories"
	v1 "github.com/JayPonda/Product-catalog/server/src/structs/v1"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

type OrderService struct {
	Db           *goqu.Database
	Logger       *utils.StructuredLogger
	OrderManager *repositories.OrderRepository
}

func InitOrderService(db *goqu.Database, logger *utils.StructuredLogger, orderManager *repositories.OrderRepository) (*OrderService, error) {
	return &OrderService{
		Db:           db,
		Logger:       logger,
		OrderManager: orderManager,
	}, nil
}

// ListOrders returns paginated orders, newest first.
func (orderServicePtr *OrderService) ListOrders(limit int, offset int) (v1.ListOrdersResponse, error) {
	var response v1.ListOrdersResponse

	orders, total, err := orderServicePtr.OrderManager.ListOrders(limit, offset)
	if err != nil {
		return response, err
	}

	response.Orders = orders
	response.Total = total
	response.Limit = limit
	response.Offset = offset

	return response, nil
}

// ListOrdersInRange returns paginated orders within a time range, oldest first.
func (orderServicePtr *OrderService) ListOrdersInRange(start time.Time, end time.Time, limit int, offset int) (v1.ListOrdersResponse, error) {
	var response v1.ListOrdersResponse

	orders, total, err := orderServicePtr.OrderManager.ListOrdersInRange(start, end, limit, offset)
	if err != nil {
		return response, err
	}

	response.Orders = orders
	response.Total = total
	response.Limit = limit
	response.Offset = offset

	return response, nil
}

// RemoveOrder deletes an order and its line items within a single transaction.
func (orderServicePtr *OrderService) RemoveOrder(id uuid.UUID) error {
	tx, err := orderServicePtr.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := orderServicePtr.OrderManager.DeleteOrder(id, tx); err != nil {
		if err == sql.ErrNoRows {
			return err
		}
		return err
	}

	return tx.Commit()
}

// RemoveOrders soft-deletes the given orders in a single batched statement.
// It returns the number of rows actually affected.
func (orderServicePtr *OrderService) RemoveOrders(ids []uuid.UUID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	return orderServicePtr.OrderManager.DeleteOrders(ids)
}

// InTx runs fn inside a single database transaction: fn receives the tx and
// all reads/writes it performs through it commit atomically, or nothing is
// applied if fn returns an error.
func (orderServicePtr *OrderService) InTx(fn func(tx *goqu.TxDatabase) error) error {
	tx, err := orderServicePtr.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

// ListOrdersInRangeTx is the transaction-scoped variant of ListOrdersInRange:
// reads run against the supplied tx so they share one consistent snapshot with
// other statements executed on it (used by the dedup job's atomic chunk).
func (orderServicePtr *OrderService) ListOrdersInRangeTx(tx *goqu.TxDatabase, start time.Time, end time.Time, limit int, offset int) (v1.ListOrdersResponse, error) {
	var response v1.ListOrdersResponse

	orders, total, err := orderServicePtr.OrderManager.ListOrdersInRange(start, end, limit, offset, tx)
	if err != nil {
		return response, err
	}

	response.Orders = orders
	response.Total = total
	response.Limit = limit
	response.Offset = offset

	return response, nil
}

// RemoveOrdersTx is the transaction-scoped variant of RemoveOrders: the
// soft-delete joins the caller's transaction instead of auto-committing, so it
// can be made atomic with the scan that selected the ids (dedup job).
func (orderServicePtr *OrderService) RemoveOrdersTx(tx *goqu.TxDatabase, ids []uuid.UUID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	return orderServicePtr.OrderManager.DeleteOrders(ids, tx)
}
