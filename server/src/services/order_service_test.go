package services_test

import (
	"testing"
	"time"

	"github.com/JayPonda/Product-catalog/server/src/repositories"
	"github.com/JayPonda/Product-catalog/server/src/services"
	"github.com/JayPonda/Product-catalog/server/testdb"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

// newOrderService spins up a real in-memory SQLite database and wires the
// production repository + service on top of it (zero mocking).
func newOrderService(t *testing.T) *services.OrderService {
	t.Helper()
	db := testdb.OpenSQLite(t)
	repo, err := repositories.InitOrderRepository(db, utils.NewStructuredLogger())
	if err != nil {
		t.Fatalf("init order repo: %v", err)
	}
	svc, err := services.InitOrderService(db, utils.NewStructuredLogger(), repo)
	if err != nil {
		t.Fatalf("init order service: %v", err)
	}
	return svc
}

func seedPair(t *testing.T, svc *services.OrderService, base time.Time) (string, string) {
	t.Helper()
	id1 := uuid.NewString()
	id2 := uuid.NewString()
	testdb.SeedOrder(t, svc.Db, id1, uuid.NewString(), 323.32, base)
	testdb.SeedOrder(t, svc.Db, id2, uuid.NewString(), 323.32, base.Add(time.Minute))
	return id1, id2
}

func TestOrderService_ListOrdersInRange_E2E(t *testing.T) {
	svc := newOrderService(t)
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	seedPair(t, svc, base)

	resp, err := svc.ListOrdersInRange(nil, base.Add(-time.Hour), base.Add(time.Hour), 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 2 || len(resp.Orders) != 2 {
		t.Errorf("expected 2 orders, got total=%d len=%d", resp.Total, len(resp.Orders))
	}
	if !resp.Orders[0].CreatedAt.Before(resp.Orders[1].CreatedAt) {
		t.Error("expected ASC ordering")
	}
}

func TestOrderService_RemoveOrders_E2E(t *testing.T) {
	svc := newOrderService(t)
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	id1, id2 := seedPair(t, svc, base)

	affected, err := svc.RemoveOrders(nil, []uuid.UUID{mustParse(t, id1), mustParse(t, id2)})
	if err != nil {
		t.Fatal(err)
	}
	if affected != 2 {
		t.Errorf("affected = %d, want 2", affected)
	}

	resp, err := svc.ListOrdersInRange(nil, base.Add(-time.Hour), base.Add(time.Hour), 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 0 {
		t.Errorf("expected 0 orders after removal, got %d", resp.Total)
	}
}

func TestOrderService_RemoveOrders_Empty_E2E(t *testing.T) {
	svc := newOrderService(t)
	affected, err := svc.RemoveOrders(nil, nil)
	if err != nil {
		t.Fatalf("empty ids should not error: %v", err)
	}
	if affected != 0 {
		t.Errorf("affected = %d, want 0", affected)
	}
}

func TestOrderService_RemoveOrder_TransactionCommit_E2E(t *testing.T) {
	svc := newOrderService(t)
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	id1, _ := seedPair(t, svc, base)

	if err := svc.RemoveOrder(nil, mustParse(t, id1)); err != nil {
		t.Fatalf("RemoveOrder: %v", err)
	}

	// Committed transaction: exactly one order remains visible.
	resp, err := svc.ListOrdersInRange(nil, base.Add(-time.Hour), base.Add(time.Hour), 100, 0)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 1 || len(resp.Orders) != 1 {
		t.Fatalf("expected 1 remaining order, got total=%d", resp.Total)
	}
	if resp.Orders[0].ID.String() == id1 {
		t.Error("removed order still visible")
	}
}

func TestOrderService_ListOrders_E2E(t *testing.T) {
	svc := newOrderService(t)
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	seedPair(t, svc, base)

	resp, err := svc.ListOrders(nil, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 2 || len(resp.Orders) != 2 {
		t.Fatalf("total=%d len=%d, want 2/2", resp.Total, len(resp.Orders))
	}
	if !resp.Orders[0].CreatedAt.After(resp.Orders[1].CreatedAt) {
		t.Error("ListOrders must be newest-first")
	}
}

func mustParse(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", s, err)
	}
	return id
}

func TestOrderService_ListOrdersInRangeTx_E2E(t *testing.T) {
	svc := newOrderService(t)
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	seedPair(t, svc, base)

	err := svc.InTx(nil, func(tx *goqu.TxDatabase) error {
		resp, err := svc.ListOrdersInRangeTx(nil, tx, base.Add(-time.Hour), base.Add(time.Hour), 100, 0)
		if err != nil {
			return err
		}
		if resp.Total != 2 || len(resp.Orders) != 2 {
			t.Errorf("expected 2 orders in tx, got total=%d len=%d", resp.Total, len(resp.Orders))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("transaction failed: %v", err)
	}
}

func TestOrderService_RemoveOrdersTx_Empty_E2E(t *testing.T) {
	svc := newOrderService(t)
	err := svc.InTx(nil, func(tx *goqu.TxDatabase) error {
		affected, err := svc.RemoveOrdersTx(nil, tx, nil)
		if err != nil {
			return err
		}
		if affected != 0 {
			t.Errorf("affected = %d, want 0", affected)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("transaction failed: %v", err)
	}
}
