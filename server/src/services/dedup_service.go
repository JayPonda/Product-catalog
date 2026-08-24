package services

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/JayPonda/Product-catalog/server/src/models"
	v1 "github.com/JayPonda/Product-catalog/server/src/structs/v1"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

// DedupStore is the persistence surface the dedup job needs.
type DedupStore interface {
	InTx(fn func(tx *goqu.TxDatabase) error) error
	ListOrdersInRangeTx(tx *goqu.TxDatabase, start time.Time, end time.Time, limit int, offset int) (v1.ListOrdersResponse, error)
	RemoveOrdersTx(tx *goqu.TxDatabase, ids []uuid.UUID) (int64, error)
}

// DedupConfig holds the options for the deduplication run.
type DedupConfig struct {
	DryRun bool
	Nearby time.Duration
	Window time.Duration
	Start  string
	End    string
	Batch  int
}

// RunDedupOrdersRemove runs the duplicate-order removal for the given time window.
func RunDedupOrdersRemove(start, end time.Time, store DedupStore, dCfg DedupConfig, logger *utils.StructuredLogger) error {
	mode := "scheduled"
	if dCfg.Start != "" && dCfg.End != "" {
		mode = "manual"
	}

	logger.Info(nil, "dedup_orders_remove", "start", "dedup run starting", utils.LoggerMeta{
		"mode":    mode,
		"start":   start.Format(time.RFC3339),
		"end":     end.Format(time.RFC3339),
		"window":  dCfg.Window.String(),
		"nearby":  dCfg.Nearby.String(),
		"batch":   dCfg.Batch,
		"dry_run": dCfg.DryRun,
	})

	// Split the range into window-sized batches.
	chunks := SplitRange(start, end, dCfg.Window, dCfg.Nearby)
	logger.Debug(nil, "dedup_orders_remove", "plan", "range split into sub-windows", utils.LoggerMeta{
		"chunks": len(chunks),
	})

	results := make([]dedupChunkResult, len(chunks))
	if len(chunks) == 1 {
		// Single batch: process sequentially (e.g. the normal scheduled 2.5h run).
		results[0] = ProcessDedupChunk(chunks[0][0], chunks[0][1], store, dCfg, logger)
	} else {
		// Multiple batches: process concurrently via a fixed worker pool, so the
		// number of goroutines stays bounded (maxConcurrency) regardless of how
		// many batches a large range is split into.
		const maxConcurrency = 5
		jobs := make(chan int, len(chunks))
		for i := range chunks {
			jobs <- i
		}
		close(jobs)

		var wg sync.WaitGroup
		for w := 0; w < maxConcurrency; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range jobs {
					s, e := chunks[i][0], chunks[i][1]
					logger.Debug(nil, "dedup_orders_remove", "chunk", "processing sub-window", utils.LoggerMeta{
						"index": i,
						"start": s.Format(time.RFC3339),
						"end":   e.Format(time.RFC3339),
					})
					results[i] = ProcessDedupChunk(s, e, store, dCfg, logger)
				}
			}()
		}
		wg.Wait()
	}

	// Aggregate results across all batches.
	var totalScanned, totalDup, totalRemoved int
	seen := make(map[uuid.UUID]struct{})
	var errs []error
	for i, r := range results {
		if r.err != nil {
			errs = append(errs, fmt.Errorf(
				"chunk %d (%s..%s): %w",
				i, chunks[i][0].Format(time.RFC3339), chunks[i][1].Format(time.RFC3339), r.err))
			continue
		}
		totalScanned += r.scanned
		totalDup += r.dupGroups
		totalRemoved += r.removed
		for _, id := range r.removedIDs {
			seen[id] = struct{}{}
		}
	}
	totalWould := len(seen)

	if dCfg.DryRun {
		logger.Info(nil, "dedup_orders_remove", "dry_run", "no changes applied (dry-run)", utils.LoggerMeta{
			"start":            start.Format(time.RFC3339),
			"end":              end.Format(time.RFC3339),
			"chunks":           len(chunks),
			"scanned":          totalScanned,
			"duplicate_groups": totalDup,
			"would_remove":     totalWould,
		})
		if len(errs) > 0 {
			return errors.Join(errs...)
		}
		return nil
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	logger.Info(nil, "dedup_orders_remove", "done", "dedup complete", utils.LoggerMeta{
		"start":            start.Format(time.RFC3339),
		"end":              end.Format(time.RFC3339),
		"chunks":           len(chunks),
		"scanned":          totalScanned,
		"duplicate_groups": totalDup,
		"removed":          totalRemoved,
	})
	return nil
}

// dedupChunkResult holds the outcome of processing a single sub-window.
type dedupChunkResult struct {
	scanned     int
	dupGroups   int
	removed     int
	wouldRemove int
	removedIDs  []uuid.UUID
	err         error
}

// ProcessDedupChunk fetches, clusters, and (unless dry-run) removes duplicates
// for one sub-window.
func ProcessDedupChunk(start, end time.Time, store DedupStore, dCfg DedupConfig, logger *utils.StructuredLogger) dedupChunkResult {
	type dupKey struct {
		customer uuid.UUID
		cents    int64
	}

	var result dedupChunkResult

	err := store.InTx(func(tx *goqu.TxDatabase) error {
		groups := make(map[dupKey][]models.Order)

		offset := 0
		scanned := 0
		for {
			resp, err := store.ListOrdersInRangeTx(tx, start, end, dCfg.Batch, offset)
			if err != nil {
				return err
			}
			for _, o := range resp.Orders {
				key := dupKey{customer: o.CustomerID, cents: int64(math.Round(o.TotalBill * 100))}
				groups[key] = append(groups[key], o)
			}
			scanned += len(resp.Orders)
			if len(resp.Orders) < dCfg.Batch {
				break
			}
			offset += len(resp.Orders)
		}

		var toRemove []uuid.UUID
		var keptIDs []uuid.UUID
		dupGroups := 0
		for _, orders := range groups {
			if len(orders) < 2 {
				continue
			}
			dupGroups++
			remove, keep := ClusterKeepLatest(orders, dCfg.Nearby)
			toRemove = append(toRemove, remove...)
			keptIDs = append(keptIDs, keep...)
		}

		byID := make(map[uuid.UUID]time.Time, len(toRemove)+len(keptIDs))
		for _, os := range groups {
			for _, o := range os {
				byID[o.ID] = o.CreatedAt
			}
		}
		redundant := make([]string, 0, len(toRemove))
		for _, id := range toRemove {
			redundant = append(redundant, fmt.Sprintf("%s@%s", id.String(), byID[id].Format(time.RFC3339)))
		}
		kept := make([]string, 0, len(keptIDs))
		for _, id := range keptIDs {
			kept = append(kept, fmt.Sprintf("%s@%s", id.String(), byID[id].Format(time.RFC3339)))
		}
		logger.Debug(nil, "dedup_orders_remove", "redundant", "orders flagged redundant (earlier ones, would remove)", utils.LoggerMeta{
			"count":     len(redundant),
			"order_ids": redundant,
		})
		logger.Debug(nil, "dedup_orders_remove", "kept", "orders kept (latest of each cluster)", utils.LoggerMeta{
			"count":     len(kept),
			"order_ids": kept,
		})

		result = dedupChunkResult{scanned: scanned, dupGroups: dupGroups, removedIDs: toRemove}

		if dCfg.DryRun {
			result.wouldRemove = len(toRemove)
			return nil
		}

		removed := int64(0)
		if len(toRemove) > 0 {
			affected, err := store.RemoveOrdersTx(tx, toRemove)
			if err != nil {
				return err
			}
			removed = affected
		}
		result.removed = int(removed)
		return nil
	})
	if err != nil {
		return dedupChunkResult{err: err}
	}

	return result
}

// SplitRange divides [start, end) into window-sized sub-windows that overlap by
// `overlap`.
func SplitRange(start, end time.Time, window, overlap time.Duration) [][2]time.Time {
	if window <= 0 {
		window = time.Hour
	}
	if overlap >= window {
		overlap = 0
	}
	var chunks [][2]time.Time
	t := start
	for t.Before(end) {
		cEnd := t.Add(window)
		if cEnd.After(end) {
			cEnd = end
		}
		chunks = append(chunks, [2]time.Time{t, cEnd})
		if cEnd == end {
			break
		}
		t = t.Add(window).Add(-overlap)
		if !t.After(chunks[len(chunks)-1][0]) {
			t = chunks[len(chunks)-1][1]
		}
	}
	if len(chunks) == 0 {
		chunks = append(chunks, [2]time.Time{start, end})
	}
	return chunks
}

// ClusterKeepLatest groups proximity orders.
func ClusterKeepLatest(orders []models.Order, nearby time.Duration) (remove, keep []uuid.UUID) {
	if len(orders) < 2 {
		return nil, nil
	}

	sorted := make([]models.Order, len(orders))
	copy(sorted, orders)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
	})

	clusterStart := 0
	prevTime := sorted[0].CreatedAt
	for i := 1; i < len(sorted); i++ {
		if sorted[i].CreatedAt.Sub(prevTime) <= nearby {
			prevTime = sorted[i].CreatedAt
			continue
		}
		// boundary: sorted[clusterStart..i-1] form a cluster; keep its last.
		for j := clusterStart; j < i-1; j++ {
			remove = append(remove, sorted[j].ID)
		}
		keep = append(keep, sorted[i-1].ID)
		clusterStart = i
		prevTime = sorted[i].CreatedAt
	}
	// flush final cluster: keep the last order, remove the earlier ones.
	for j := clusterStart; j < len(sorted)-1; j++ {
		remove = append(remove, sorted[j].ID)
	}
	keep = append(keep, sorted[len(sorted)-1].ID)
	return remove, keep
}
