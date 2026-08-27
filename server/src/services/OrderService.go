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
func (orderServicePtr *OrderService) ListOrders(ctx utils.RequestContext, limit int, offset int) (v1.ListOrdersResponse, error) {
	var response v1.ListOrdersResponse

	orders, total, err := orderServicePtr.OrderManager.ListOrders(ctx, limit, offset)
	if err != nil {
		orderServicePtr.Logger.Error(ctx, "OrderService.go", "ListOrders", "failed to list orders", utils.LoggerMeta{"limit": limit, "offset": offset}, err.Error())
		return response, err
	}

	response.Orders = orders
	response.Total = total
	response.Limit = limit
	response.Offset = offset

	orderServicePtr.Logger.Debug(ctx, "OrderService.go", "ListOrders", "orders listed", utils.LoggerMeta{"count": len(orders), "total": total})
	return response, nil
}

// ListOrdersInRange returns paginated orders within a time range, oldest first.
func (orderServicePtr *OrderService) ListOrdersInRange(ctx utils.RequestContext, start time.Time, end time.Time, limit int, offset int) (v1.ListOrdersResponse, error) {
	var response v1.ListOrdersResponse

	orders, total, err := orderServicePtr.OrderManager.ListOrdersInRange(ctx, start, end, limit, offset)
	if err != nil {
		orderServicePtr.Logger.Error(ctx, "OrderService.go", "ListOrdersInRange", "failed to list orders in range", utils.LoggerMeta{"limit": limit, "offset": offset}, err.Error())
		return response, err
	}

	response.Orders = orders
	response.Total = total
	response.Limit = limit
	response.Offset = offset

	orderServicePtr.Logger.Debug(ctx, "OrderService.go", "ListOrdersInRange", "orders in range listed", utils.LoggerMeta{"count": len(orders), "total": total})
	return response, nil
}

// RemoveOrder deletes an order and its line items within a single transaction.
func (orderServicePtr *OrderService) RemoveOrder(ctx utils.RequestContext, id uuid.UUID) error {
	tx, err := orderServicePtr.Db.Begin()
	if err != nil {
		orderServicePtr.Logger.Error(ctx, "OrderService.go", "RemoveOrder", "failed to begin transaction", utils.LoggerMeta{"id": id.String()}, err.Error())
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if err := orderServicePtr.OrderManager.DeleteOrder(ctx, id, tx); err != nil {
		if err == sql.ErrNoRows {
			orderServicePtr.Logger.Warn(ctx, "OrderService.go", "RemoveOrder", "order not found", utils.LoggerMeta{"id": id.String()})
		} else {
			orderServicePtr.Logger.Error(ctx, "OrderService.go", "RemoveOrder", "failed to delete order", utils.LoggerMeta{"id": id.String()}, err.Error())
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		orderServicePtr.Logger.Error(ctx, "OrderService.go", "RemoveOrder", "failed to commit transaction", utils.LoggerMeta{"id": id.String()}, err.Error())
		return err
	}

	orderServicePtr.Logger.Debug(ctx, "OrderService.go", "RemoveOrder", "order removed", utils.LoggerMeta{"id": id.String()})
	return nil
}

// RemoveOrders soft-deletes the given orders in a single batched statement.
// It returns the number of rows actually affected.
func (orderServicePtr *OrderService) RemoveOrders(ctx utils.RequestContext, ids []uuid.UUID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result, err := orderServicePtr.OrderManager.DeleteOrders(ctx, ids)
	if err != nil {
		orderServicePtr.Logger.Error(ctx, "OrderService.go", "RemoveOrders", "failed to delete orders", utils.LoggerMeta{"count": len(ids)}, err.Error())
		return 0, err
	}
	orderServicePtr.Logger.Debug(ctx, "OrderService.go", "RemoveOrders", "orders removed", utils.LoggerMeta{"count": len(ids), "rows_affected": result})
	return result, nil
}

// InTx runs fn inside a single database transaction: fn receives the tx and
// all reads/writes it performs through it commit atomically, or nothing is
// applied if fn returns an error.
func (orderServicePtr *OrderService) InTx(ctx utils.RequestContext, fn func(tx *goqu.TxDatabase) error) error {
	tx, err := orderServicePtr.Db.Begin()
	if err != nil {
		orderServicePtr.Logger.Error(ctx, "OrderService.go", "InTx", "failed to begin transaction", utils.LoggerMeta{}, err.Error())
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if err := fn(tx); err != nil {
		orderServicePtr.Logger.Error(ctx, "OrderService.go", "InTx", "transaction function failed", utils.LoggerMeta{}, err.Error())
		return err
	}

	if err := tx.Commit(); err != nil {
		orderServicePtr.Logger.Error(ctx, "OrderService.go", "InTx", "failed to commit transaction", utils.LoggerMeta{}, err.Error())
		return err
	}

	orderServicePtr.Logger.Debug(ctx, "OrderService.go", "InTx", "transaction committed", utils.LoggerMeta{})
	return nil
}

// ListOrdersInRangeTx is the transaction-scoped variant of ListOrdersInRange:
// reads run against the supplied tx so they share one consistent snapshot with
// other statements executed on it (used by the dedup job's atomic chunk).
func (orderServicePtr *OrderService) ListOrdersInRangeTx(ctx utils.RequestContext, tx *goqu.TxDatabase, start time.Time, end time.Time, limit int, offset int) (v1.ListOrdersResponse, error) {
	var response v1.ListOrdersResponse

	orders, total, err := orderServicePtr.OrderManager.ListOrdersInRange(ctx, start, end, limit, offset, tx)
	if err != nil {
		orderServicePtr.Logger.Error(ctx, "OrderService.go", "ListOrdersInRangeTx", "failed to list orders in range", utils.LoggerMeta{"limit": limit, "offset": offset}, err.Error())
		return response, err
	}

	response.Orders = orders
	response.Total = total
	response.Limit = limit
	response.Offset = offset

	orderServicePtr.Logger.Debug(ctx, "OrderService.go", "ListOrdersInRangeTx", "orders in range listed", utils.LoggerMeta{"count": len(orders), "total": total})
	return response, nil
}

// RemoveOrdersTx is the transaction-scoped variant of RemoveOrders: the
// soft-delete joins the caller's transaction instead of auto-committing, so it
// can be made atomic with the scan that selected the ids (dedup job).
func (orderServicePtr *OrderService) RemoveOrdersTx(ctx utils.RequestContext, tx *goqu.TxDatabase, ids []uuid.UUID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result, err := orderServicePtr.OrderManager.DeleteOrders(ctx, ids, tx)
	if err != nil {
		orderServicePtr.Logger.Error(ctx, "OrderService.go", "RemoveOrdersTx", "failed to delete orders in transaction", utils.LoggerMeta{"count": len(ids)}, err.Error())
		return 0, err
	}
	orderServicePtr.Logger.Debug(ctx, "OrderService.go", "RemoveOrdersTx", "orders removed in transaction", utils.LoggerMeta{"count": len(ids), "rows_affected": result})
	return result, nil
}
