package cmd

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/JayPonda/Product-catalog/server/src/models"
	"github.com/JayPonda/Product-catalog/server/src/repositories"
	"github.com/JayPonda/Product-catalog/server/src/services"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// dedupOrdersRemoveCmd is a placeholder. The duplicate-order removal logic will
// be implemented later; it is expected to use OrderService (ListOrders /
// RemoveOrder) to find and delete duplicate orders.
// considering this is used as a command in any job/cronjob
// core philology is not to clean up full data, rather then this is used as scheduled job.
// why? so for this use case, if orders are duplicate and this order is placed by the humans, that means it might chance that
// this is happening because of double click or similar condition. so that means order is multiplied in very short period of time.
// this can happen with any technical reasons. so here assumption is we will run this job each 4 hours.
// main philosophy was there is no guaranteed that order can duplicate in 4 hour of window, that can be happen or left due to long processing hour
// so we will run this at 4 hours but we will take data of 5 hour, one hour prior of the job, to currant time.
// main assumption will be this command can complete at any time, but it will always starts in time.
// this this assumption can't possible we need to store the last cron time, so we can take according ly
// from that time to currant time kind of thing. and then we need to apply batch processing for 4 hour + 5 minute window.
// as continuative batch check no need for longer window.
// this process divided into four parts,
// 0. define the meaning of deduplication
// 1. get the the orders of 5 hours, this will run relative to the current time. so (5 hour prior to current time)
// 2. find the duplicated entries among them
// 3. fix those orders/remove those orders

// dedup command flags.
var (
	dedupDryRun bool
	dedupNearby time.Duration
	dedupWindow time.Duration
	dedupStart  string
	dedupEnd    string
	dedupBatch  int
)

// dedupOrdersRemoveCmd builds the duplicate-order removal command. It receives
// the shared config and logger as arguments so the command pulls envs/context
// from configuration rather than reaching for globals.
func dedupOrdersRemoveCmd(cfg AppConfig, logger *utils.StructuredLogger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dedup-orders-remove",
		Short: "Remove duplicate orders",
		RunE: func(cmd *cobra.Command, args []string) error {
			// part 2: define start and end time for this session.
			var start, end time.Time

			if dedupStart != "" && dedupEnd != "" {
				// manual mode: both --starttime and --endtime supplied (RFC3339).
				var err error
				if start, err = time.Parse(time.RFC3339, dedupStart); err != nil {
					logger.Error(nil, "dedup_orders_remove", "config", "invalid --starttime", utils.LoggerMeta{"value": dedupStart}, err.Error())
					return fmt.Errorf("invalid --starttime %q: %w", dedupStart, err)
				}
				if end, err = time.Parse(time.RFC3339, dedupEnd); err != nil {
					logger.Error(nil, "dedup_orders_remove", "config", "invalid --endtime", utils.LoggerMeta{"value": dedupEnd}, err.Error())
					return fmt.Errorf("invalid --endtime %q: %w", dedupEnd, err)
				}
				if !end.After(start) {
					logger.Error(nil, "dedup_orders_remove", "config", "end must be after start", utils.LoggerMeta{"start": dedupStart, "end": dedupEnd}, "")
					return fmt.Errorf("--endtime %q must be after --starttime %q", dedupEnd, dedupStart)
				}
			} else {
				// scheduled mode: end = now, start = now - window (2.5h default,
				// 2h interval + 0.5h overlap so boundary-straddling clusters are caught).
				now := time.Now()
				end = now
				start = now.Add(-dedupWindow)
			}

			return dedupOrdersRemove(start, end, cfg, logger)
		},
	}

	cmd.Flags().BoolVar(&dedupDryRun, "dry-run", false, "report duplicates without deleting")
	cmd.Flags().DurationVar(&dedupNearby, "nearby", 5*time.Minute, "max gap between orders to treat as duplicate")
	cmd.Flags().DurationVar(&dedupWindow, "window", 150*time.Minute, "auto-mode lookback window (default 2.5h)")
	cmd.Flags().StringVar(&dedupStart, "starttime", "", "manual mode: RFC3339 start (requires --endtime)")
	cmd.Flags().StringVar(&dedupEnd, "endtime", "", "manual mode: RFC3339 end (requires --starttime)")
	cmd.Flags().IntVar(&dedupBatch, "batch", 200, "fetch batch size")

	return cmd
}

// dedupOrdersRemove runs the duplicate-order removal for the given time window.
// It receives the app config and logger as arguments so the command can pull
// envs/context from configuration rather than reaching for globals inside.
//
// If the requested range is larger than a single window, it is split into
// window-sized batches that are processed concurrently (bounded worker pool),
// so a long manual range (e.g. 6h -> three 2.5h batches) runs in parallel while
// staying within DB limits.
func dedupOrdersRemove(start, end time.Time, cfg AppConfig, logger *utils.StructuredLogger) error {
	mode := "scheduled"
	if dedupStart != "" && dedupEnd != "" {
		mode = "manual"
	}

	logger.Info(nil, "dedup_orders_remove", "start", "dedup run starting", utils.LoggerMeta{
		"mode":    mode,
		"start":   start.Format(time.RFC3339),
		"end":     end.Format(time.RFC3339),
		"window":  dedupWindow.String(),
		"nearby":  dedupNearby.String(),
		"batch":   dedupBatch,
		"dry_run": dedupDryRun,
	})

	db := utils.InitDB(cfg)

	orderRepo, err := repositories.InitOrderRepository(db, logger)
	if err != nil {
		logger.Error(nil, "dedup_orders_remove", "init", "failed to init order repository", nil, err.Error())
		return err
	}
	orderSvc, err := services.InitOrderService(db, logger, orderRepo)
	if err != nil {
		logger.Error(nil, "dedup_orders_remove", "init", "failed to init order service", nil, err.Error())
		return err
	}

	// Split the range into window-sized batches.
	chunks := splitRange(start, end, dedupWindow, dedupNearby)
	logger.Debug(nil, "dedup_orders_remove", "plan", "range split into sub-windows", utils.LoggerMeta{
		"chunks": len(chunks),
	})

	results := make([]dedupChunkResult, len(chunks))
	if len(chunks) == 1 {
		// Single batch: process sequentially (e.g. the normal scheduled 2.5h run).
		results[0] = processDedupChunk(chunks[0][0], chunks[0][1], orderSvc, logger)
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
					results[i] = processDedupChunk(s, e, orderSvc, logger)
				}
			}()
		}
		wg.Wait()
	}

	// Aggregate results across all batches. removedIDs are unioned so that
	// orders appearing in overlapping batch regions are not double-counted.
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

	if dedupDryRun {
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

// processDedupChunk fetches, clusters, and (unless dry-run) removes duplicates
// for one sub-window. It is safe to run concurrently for disjoint windows.
func processDedupChunk(start, end time.Time, orderSvc *services.OrderService, logger *utils.StructuredLogger) dedupChunkResult {
	type dupKey struct {
		customer uuid.UUID
		cents    int64
	}
	groups := make(map[dupKey][]models.Order)

	offset := 0
	scanned := 0
	for {
		resp, err := orderSvc.ListOrdersInRange(start, end, dedupBatch, offset)
		if err != nil {
			return dedupChunkResult{err: err}
		}
		for _, o := range resp.Orders {
			key := dupKey{customer: o.CustomerID, cents: int64(math.Round(o.TotalBill * 100))}
			groups[key] = append(groups[key], o)
		}
		scanned += len(resp.Orders)
		if len(resp.Orders) < dedupBatch {
			break
		}
		offset += len(resp.Orders)
	}

	// define duplicates (same customer + value, nearby created_at). Within each
	// time-proximity cluster the LATEST order is kept and the earlier ones are
	// treated as redundant and removed.
	var toRemove []uuid.UUID
	var keptIDs []uuid.UUID
	dupGroups := 0
	for _, orders := range groups {
		if len(orders) < 2 {
			continue
		}
		sort.Slice(orders, func(i, j int) bool {
			return orders[i].CreatedAt.Before(orders[j].CreatedAt)
		})

		// Walk clusters: orders within --nearby of the previous order belong to
		// the same cluster; the latest order of each cluster is kept.
		dupGroups++
		clusterStart := 0
		prevTime := orders[0].CreatedAt
		for i := 1; i < len(orders); i++ {
			if orders[i].CreatedAt.Sub(prevTime) <= dedupNearby {
				prevTime = orders[i].CreatedAt
				continue
			}
			// boundary: orders[clusterStart..i-1] form a cluster; keep its last.
			for j := clusterStart; j < i-1; j++ {
				toRemove = append(toRemove, orders[j].ID)
			}
			keptIDs = append(keptIDs, orders[i-1].ID)
			clusterStart = i
			prevTime = orders[i].CreatedAt
		}
		// flush final cluster: keep the last order, remove the earlier ones.
		for j := clusterStart; j < len(orders)-1; j++ {
			toRemove = append(toRemove, orders[j].ID)
		}
		keptIDs = append(keptIDs, orders[len(orders)-1].ID)
	}

	// Diagnostic: prove the latest is kept and earlier ones are redundant by
	// surfacing the flagged IDs with their created_at.
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

	if dedupDryRun {
		return dedupChunkResult{scanned: scanned, dupGroups: dupGroups, wouldRemove: len(toRemove), removedIDs: toRemove}
	}

	removed := 0
	if len(toRemove) > 0 {
		affected, err := orderSvc.RemoveOrders(toRemove)
		if err != nil {
			return dedupChunkResult{err: err}
		}
		removed = int(affected)
	}
	return dedupChunkResult{scanned: scanned, dupGroups: dupGroups, removed: removed, removedIDs: toRemove}
}

// splitRange divides [start, end) into window-sized sub-windows that overlap by
// `overlap`, so a duplicate cluster straddling a batch boundary (within the
// nearby threshold) is still fully captured by one batch. The final batch is
// clamped to `end`. A range smaller than one window yields a single chunk.
func splitRange(start, end time.Time, window, overlap time.Duration) [][2]time.Time {
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
