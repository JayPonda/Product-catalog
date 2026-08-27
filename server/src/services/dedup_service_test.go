package services

import (
	"errors"
	"testing"
	"time"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	v1 "github.com/JayPonda/Product-catalog/server/src/structs/v1"
	"github.com/JayPonda/Product-catalog/server/testdb"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

// ---- helpers ----

func mkOrder(t *testing.T, createdAt time.Time) models.Order {
	t.Helper()
	id, err := uuid.NewUUID()
	if err != nil {
		t.Fatalf("new uuid: %v", err)
	}
	return models.Order{ID: id, CustomerID: id, TotalBill: 100, CreatedAt: createdAt}
}

func contains(ids []uuid.UUID, target uuid.UUID) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func chunkContaining(chunks [][2]time.Time, at time.Time) *[2]time.Time {
	for i := range chunks {
		if !at.Before(chunks[i][0]) && !at.After(chunks[i][1]) {
			return &chunks[i]
		}
	}
	return nil
}

func chunkContainingBoth(chunks [][2]time.Time, a, b time.Time) *[2]time.Time {
	for i := range chunks {
		if chunkContaining(chunks[i:i+1], a) != nil && chunkContaining(chunks[i:i+1], b) != nil {
			return &chunks[i]
		}
	}
	return nil
}

// ---- ClusterKeepLatest (pure) ----

func TestClusterKeepLatest_TwoOrdersKeepsLatest(t *testing.T) {
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	first := mkOrder(t, base)
	second := mkOrder(t, base.Add(time.Minute))

	remove, keep := ClusterKeepLatest([]models.Order{first, second}, 5*time.Minute)

	if len(remove) != 1 || remove[0] != first.ID {
		t.Errorf("expected earlier order %s removed, got %v", first.ID, remove)
	}
	if len(keep) != 1 || keep[0] != second.ID {
		t.Errorf("expected latest order %s kept, got %v", second.ID, keep)
	}
}

func TestClusterKeepLatest_ChainOfThreeRemovesFirstTwo(t *testing.T) {
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	o1 := mkOrder(t, base)
	o2 := mkOrder(t, base.Add(time.Minute))
	o3 := mkOrder(t, base.Add(2*time.Minute))

	remove, keep := ClusterKeepLatest([]models.Order{o3, o1, o2}, 5*time.Minute) // unsorted input

	if len(keep) != 1 || keep[0] != o3.ID {
		t.Errorf("expected only latest %s kept, got %v", o3.ID, keep)
	}
	if len(remove) != 2 || !contains(remove, o1.ID) || !contains(remove, o2.ID) {
		t.Errorf("expected o1 and o2 removed, got %v", remove)
	}
}

func TestClusterKeepLatest_GapSplitsClusters(t *testing.T) {
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	o1 := mkOrder(t, base)
	o2 := mkOrder(t, base.Add(time.Minute))
	// 9 minute gap (> nearby of 5m) starts a brand-new cluster.
	o3 := mkOrder(t, base.Add(10*time.Minute))
	o4 := mkOrder(t, base.Add(11*time.Minute))

	remove, keep := ClusterKeepLatest([]models.Order{o1, o2, o3, o4}, 5*time.Minute)

	// Each cluster keeps its own latest; both cluster-latests survive.
	if len(keep) != 2 || !contains(keep, o2.ID) || !contains(keep, o4.ID) {
		t.Errorf("expected o2 and o4 kept, got %v", keep)
	}
	if len(remove) != 2 || !contains(remove, o1.ID) || !contains(remove, o3.ID) {
		t.Errorf("expected o1 and o3 removed, got %v", remove)
	}
}

func TestClusterKeepLatest_EdgeCases(t *testing.T) {
	if remove, keep := ClusterKeepLatest(nil, 5*time.Minute); remove != nil || keep != nil {
		t.Error("nil slice should yield nil outputs")
	}
	solo := mkOrder(t, time.Now())
	if remove, keep := ClusterKeepLatest([]models.Order{solo}, 5*time.Minute); remove != nil || keep != nil {
		t.Error("single-element slice should yield nil outputs (nothing to deduplicate)")
	}
}

// ---- SplitRange (pure) ----

func TestSplitRange_SingleBatch(t *testing.T) {
	start := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)

	chunks := SplitRange(start, end, 150*time.Minute, 5*time.Minute)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0][0] != start || chunks[0][1] != end {
		t.Errorf("expected chunk %v..%v, got %v..%v", start, end, chunks[0][0], chunks[0][1])
	}
}

func TestSplitRange_AlignsExactlyOnWindowBoundaries(t *testing.T) {
	start := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	end := start.Add(4 * time.Hour) // 240 minutes

	window := 120 * time.Minute
	overlap := 10 * time.Minute

	chunks := SplitRange(start, end, window, overlap)

	// Range splits into:
	// Chunk 1: T0 .. T0+120m
	// Chunk 2: T0+110m (T0+120m - 10m) .. T0+230m
	// Chunk 3: T0+220m .. T0+240m (clamped to end)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d: %v", len(chunks), chunks)
	}

	if chunks[0][0] != start || chunks[0][1] != start.Add(120*time.Minute) {
		t.Errorf("chunk 1 mismatch: %v", chunks[0])
	}
	if chunks[1][0] != start.Add(110*time.Minute) || chunks[1][1] != start.Add(230*time.Minute) {
		t.Errorf("chunk 2 mismatch: %v", chunks[1])
	}
	if chunks[2][0] != start.Add(220*time.Minute) || chunks[2][1] != end {
		t.Errorf("chunk 3 mismatch: %v", chunks[2])
	}

	// Boundary order placed at T0+115m (inside the overlap) must belong to both chunk 1 and chunk 2.
	boundaryTime := start.Add(115 * time.Minute)
	if chunkContainingBoth(chunks, boundaryTime, boundaryTime) == nil {
		t.Errorf("boundary order at %s should belong to overlapping chunks 1 and 2", boundaryTime)
	}
}

func TestSplitRange_InvalidArguments(t *testing.T) {
	start := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)

	// 0 window fallback
	chunks := SplitRange(start, end, 0, 5*time.Minute)
	if len(chunks) == 0 {
		t.Fatal("expected chunks, got empty list")
	}

	// overlap >= window resets overlap to 0
	noOverlap := SplitRange(start, end, 150*time.Minute, 200*time.Minute)
	if len(noOverlap) != 1 {
		t.Errorf("expected 1 chunk, got %d", len(noOverlap))
	}
}

// ---- atomic chunk pipeline (real SQLite + fault-injecting stub) ----

func getDedupConfig(dryRun bool, nearby, window time.Duration, batch int) DedupConfig {
	return DedupConfig{
		DryRun: dryRun,
		Nearby: nearby,
		Window: window,
		Batch:  batch,
	}
}

// newDedupService wires a real in-memory SQLite database with the production
// repository + service (zero mocking) for driving the actual dedup pipeline.
func newDedupService(t *testing.T) *OrderService {
	t.Helper()
	db := testdb.OpenSQLite(t)
	repo, err := repositories.InitOrderRepository(db, utils.NewStructuredLogger())
	if err != nil {
		t.Fatalf("init order repo: %v", err)
	}
	svc, err := InitOrderService(db, utils.NewStructuredLogger(), repo)
	if err != nil {
		t.Fatalf("init order service: %v", err)
	}
	return svc
}

func seedOrderRow(t *testing.T, db *goqu.Database, customer string, bill float64, createdAt time.Time) uuid.UUID {
	t.Helper()
	id := uuid.New()
	testdb.SeedOrder(t, db, id.String(), customer, bill, createdAt)
	return id
}

func aliveOrderIDs(t *testing.T, db *goqu.Database) map[string]bool {
	t.Helper()
	var ids []string
	err := db.From("orders").
		Where(goqu.C("deleted_at").IsNull()).
		Select("id").
		ScanVals(&ids)
	if err != nil {
		t.Fatalf("query alive orders: %v", err)
	}
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}

// stubStore implements DedupStore in memory.
type stubStore struct {
	orders       []models.Order
	scanErrAfter int // pages served successfully before scan fails (0 = never)
	removeErr    error
	pagesServed  int
	removeCalls  int
	removedIDs   []uuid.UUID
	txs          int
	committed    bool
	rolledBack   bool
}

func (s *stubStore) InTx(ctx utils.RequestContext, fn func(tx *goqu.TxDatabase) error) error {
	s.txs++
	if err := fn(nil); err != nil { // tx handle is unused by this stub
		s.rolledBack = true
		return err
	}
	s.committed = true
	return nil
}

func (s *stubStore) ListOrdersInRangeTx(ctx utils.RequestContext, tx *goqu.TxDatabase, start time.Time, end time.Time, limit int, offset int) (v1.ListOrdersResponse, error) {
	var resp v1.ListOrdersResponse
	if s.scanErrAfter > 0 && s.pagesServed >= s.scanErrAfter {
		s.pagesServed++
		return resp, errors.New("boom mid-scan")
	}
	s.pagesServed++
	endIdx := offset + limit
	if endIdx > len(s.orders) {
		endIdx = len(s.orders)
	}
	if offset > len(s.orders) {
		offset = len(s.orders)
	}
	resp.Orders = s.orders[offset:endIdx]
	resp.Total = int64(len(s.orders))
	return resp, nil
}

func (s *stubStore) RemoveOrdersTx(ctx utils.RequestContext, tx *goqu.TxDatabase, ids []uuid.UUID) (int64, error) {
	s.removeCalls++
	if s.removeErr != nil {
		return 0, s.removeErr
	}
	s.removedIDs = append(s.removedIDs, ids...)
	return int64(len(ids)), nil
}

func TestProcessDedupChunk_RemovesEarlierDuplicates_E2E(t *testing.T) {
	svc := newDedupService(t)
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)

	custA := uuid.NewString()
	o1 := seedOrderRow(t, svc.Db, custA, 42.00, base)
	o2 := seedOrderRow(t, svc.Db, custA, 42.00, base.Add(time.Minute))
	solo := seedOrderRow(t, svc.Db, uuid.NewString(), 7.00, base.Add(30*time.Minute))

	res := ProcessDedupChunk(nil, base.Add(-time.Hour), base.Add(time.Hour), svc, getDedupConfig(false, 5*time.Minute, time.Hour, 100), utils.NewStructuredLogger())

	if res.err != nil {
		t.Fatalf("ProcessDedupChunk returned error: %v", res.err)
	}
	if res.scanned != 3 || res.dupGroups != 1 || res.removed != 1 {
		t.Errorf("result = %+v, want scanned=3 dupGroups=1 removed=1", res)
	}
	if len(res.removedIDs) != 1 || res.removedIDs[0] != o1 {
		t.Errorf("removedIDs = %v, want [%s]", res.removedIDs, o1)
	}

	alive := aliveOrderIDs(t, svc.Db)
	if !alive[o2.String()] || !alive[solo.String()] {
		t.Errorf("expected o2 and solo alive, got %v", alive)
	}
	if alive[o1.String()] {
		t.Error("expected earlier duplicate o1 soft-deleted")
	}
	if len(alive) != 2 {
		t.Errorf("alive count = %d, want 2", len(alive))
	}
}

func TestProcessDedupChunk_DryRunLeavesRowsIntact_E2E(t *testing.T) {
	svc := newDedupService(t)
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)

	custA := uuid.NewString()
	o1 := seedOrderRow(t, svc.Db, custA, 42.00, base)
	seedOrderRow(t, svc.Db, custA, 42.00, base.Add(time.Minute))

	res := ProcessDedupChunk(nil, base.Add(-time.Hour), base.Add(time.Hour), svc, getDedupConfig(true, 5*time.Minute, time.Hour, 100), utils.NewStructuredLogger())

	if res.err != nil {
		t.Fatalf("ProcessDedupChunk returned error: %v", res.err)
	}
	if res.wouldRemove != 1 || res.removed != 0 {
		t.Errorf("result = %+v, want wouldRemove=1 removed=0", res)
	}
	if len(res.removedIDs) != 1 || res.removedIDs[0] != o1 {
		t.Errorf("removedIDs should flag only o1 (%s), got %v", o1, res.removedIDs)
	}
	if got := aliveOrderIDs(t, svc.Db); len(got) != 2 {
		t.Errorf("dry-run must not touch rows, alive = %v", got)
	}
}

func TestProcessDedupChunk_ScanFailureAbortsBeforeAnyDelete(t *testing.T) {
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)

	store := &stubStore{scanErrAfter: 1} // first page OK, second page fails
	for i := 0; i < 4; i++ {             // 4 orders => needs a second page at batch=2
		store.orders = append(store.orders, models.Order{
			ID:         uuid.New(),
			CustomerID: uuid.New(),
			TotalBill:  10,
			CreatedAt:  base.Add(time.Duration(i) * time.Minute),
		})
	}

	res := ProcessDedupChunk(nil, base.Add(-time.Hour), base.Add(time.Hour), store, getDedupConfig(false, 5*time.Minute, time.Hour, 2), utils.NewStructuredLogger())

	if res.err == nil {
		t.Fatal("expected mid-scan failure to surface as chunk error")
	}
	if store.removeCalls != 0 {
		t.Errorf("no delete may run after a failed scan, got %d calls", store.removeCalls)
	}
	if store.committed || !store.rolledBack {
		t.Errorf("transaction must roll back on scan failure (committed=%v rolledBack=%v)", store.committed, store.rolledBack)
	}
}

func TestProcessDedupChunk_RemoveFailureRollsBackChunk(t *testing.T) {
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)

	store := &stubStore{removeErr: errors.New("boom on delete")}
	cust := uuid.New()
	store.orders = append(store.orders,
		models.Order{ID: uuid.New(), CustomerID: cust, TotalBill: 10, CreatedAt: base},
		models.Order{ID: uuid.New(), CustomerID: cust, TotalBill: 10, CreatedAt: base.Add(time.Minute)},
	)

	res := ProcessDedupChunk(nil, base.Add(-time.Hour), base.Add(time.Hour), store, getDedupConfig(false, 5*time.Minute, time.Hour, 100), utils.NewStructuredLogger())

	if res.err == nil {
		t.Fatal("expected delete failure to surface as chunk error")
	}
	if store.removeCalls != 1 {
		t.Errorf("delete attempted once, got %d", store.removeCalls)
	}
	if store.committed || !store.rolledBack {
		t.Errorf("transaction must roll back so scan+scan-derived delete stay all-or-nothing (committed=%v rolledBack=%v)", store.committed, store.rolledBack)
	}
}

func TestRunDedupOrdersRemove_MultiChunkIdempotentRerun_E2E(t *testing.T) {
	const window = 30 * time.Minute
	const nearby = 5 * time.Minute

	svc := newDedupService(t)
	base := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	start, end := base, base.Add(70*time.Minute)

	chunks := SplitRange(start, end, window, nearby)
	if len(chunks) != 3 {
		t.Fatalf("precondition: expected 3 chunks, got %d (%v)", len(chunks), chunks)
	}

	custA := uuid.NewString()
	bFirst := seedOrderRow(t, svc.Db, custA, 42.00, base.Add(29*time.Minute))
	bSecond := seedOrderRow(t, svc.Db, custA, 42.00, base.Add(32*time.Minute))

	custB := uuid.NewString()
	d1 := seedOrderRow(t, svc.Db, custB, 10.00, base.Add(40*time.Minute))
	d2 := seedOrderRow(t, svc.Db, custB, 10.00, base.Add(41*time.Minute))

	s1 := seedOrderRow(t, svc.Db, uuid.NewString(), 5.00, base.Add(10*time.Minute))
	s2 := seedOrderRow(t, svc.Db, uuid.NewString(), 7.00, base.Add(60*time.Minute))

	if err := RunDedupOrdersRemove(nil, start, end, svc, getDedupConfig(false, nearby, window, 100), utils.NewStructuredLogger()); err != nil {
		t.Fatalf("run 1: %v", err)
	}

	wantAlive := map[string]bool{
		bSecond.String(): true,
		d2.String():      true,
		s1.String():      true,
		s2.String():      true,
	}
	gotAlive := aliveOrderIDs(t, svc.Db)
	if len(gotAlive) != len(wantAlive) {
		t.Fatalf("after run 1 alive = %v, want exactly %v", gotAlive, wantAlive)
	}
	for id := range wantAlive {
		if !gotAlive[id] {
			t.Errorf("expected %s to survive, alive = %v", id, gotAlive)
		}
	}
	if gotAlive[bFirst.String()] || gotAlive[d1.String()] {
		t.Errorf("earlier duplicates bFirst(%s)/d1(%s) must be gone", bFirst, d1)
	}

	if err := RunDedupOrdersRemove(nil, start, end, svc, getDedupConfig(false, nearby, window, 100), utils.NewStructuredLogger()); err != nil {
		t.Fatalf("run 2: %v", err)
	}
	if again := aliveOrderIDs(t, svc.Db); len(again) != len(wantAlive) {
		t.Errorf("rerun changed state: alive = %v", again)
	}
}
