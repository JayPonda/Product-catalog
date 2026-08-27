package services_test

import (
	"testing"
	"time"

	"github.com/JayPonda/Product-catalog/server/src/services"
	v1 "github.com/JayPonda/Product-catalog/server/src/structs/v1"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

// errDedupStore makes ListOrdersInRangeTx always fail so ProcessDedupChunk
// surfaces an error that RunDedupOrdersRemove must aggregate and return.
type errDedupStore struct{}

func (errDedupStore) InTx(ctx utils.RequestContext, fn func(tx *goqu.TxDatabase) error) error {
	return fn(nil)
}

func (errDedupStore) ListOrdersInRangeTx(ctx utils.RequestContext, tx *goqu.TxDatabase, start, end time.Time, limit, offset int) (v1.ListOrdersResponse, error) {
	return v1.ListOrdersResponse{}, errBoom
}

func (errDedupStore) RemoveOrdersTx(ctx utils.RequestContext, tx *goqu.TxDatabase, ids []uuid.UUID) (int64, error) {
	return 0, nil
}

func TestRunDedupOrdersRemove_ChunkErrorPropagates(t *testing.T) {
	start := time.Now().Add(-2 * time.Hour)
	end := time.Now()
	err := services.RunDedupOrdersRemove(nil, start, end, errDedupStore{},
		services.DedupConfig{Window: time.Hour}, utils.NewStructuredLogger())
	if err == nil {
		t.Error("expected chunk error to propagate")
	}
}

func TestRunDedupOrdersRemove_DryRunChunkError(t *testing.T) {
	start := time.Now().Add(-2 * time.Hour)
	end := time.Now()
	err := services.RunDedupOrdersRemove(nil, start, end, errDedupStore{},
		services.DedupConfig{Window: time.Hour, DryRun: true}, utils.NewStructuredLogger())
	if err == nil {
		t.Error("expected dry-run chunk error to propagate")
	}
}
