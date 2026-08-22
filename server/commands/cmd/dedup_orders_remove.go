package cmd

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
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

	// part 2: fetch all orders in the window, batched (default 200).
	type dupKey struct {
		customer uuid.UUID
		cents    int64
	}
	groups := make(map[dupKey][]models.Order)

	offset := 0
	var scanned int
	for {
		resp, err := orderSvc.ListOrdersInRange(start, end, dedupBatch, offset)
		if err != nil {
			logger.Error(nil, "dedup_orders_remove", "fetch", "failed to list orders in range", utils.LoggerMeta{"offset": offset}, err.Error())
			return err
		}
		for _, o := range resp.Orders {
			key := dupKey{customer: o.CustomerID, cents: int64(math.Round(o.TotalBill * 100))}
			groups[key] = append(groups[key], o)
		}
		scanned += len(resp.Orders)
		logger.Debug(nil, "dedup_orders_remove", "fetch", "fetched batch", utils.LoggerMeta{
			"offset":        offset,
			"fetched":       len(resp.Orders),
			"scanned_total": scanned,
		})
		if len(resp.Orders) < dedupBatch {
			break
		}
		offset += len(resp.Orders)
	}

	logger.Info(nil, "dedup_orders_remove", "scan", "orders scanned", utils.LoggerMeta{
		"scanned": scanned,
		"groups":  len(groups),
	})

	// part 0 + 3: define duplicates (same customer + value, nearby created_at)
	// and keep the earliest of each time-proximity cluster.
	var toRemove []uuid.UUID
	dupGroups := 0
	for _, orders := range groups {
		if len(orders) < 2 {
			continue
		}
		sort.Slice(orders, func(i, j int) bool {
			return orders[i].CreatedAt.Before(orders[j].CreatedAt)
		})
		dupGroups++
		lastKeep := orders[0].CreatedAt
		for i := 1; i < len(orders); i++ {
			if orders[i].CreatedAt.Sub(lastKeep) <= dedupNearby {
				toRemove = append(toRemove, orders[i].ID)
			} else {
				lastKeep = orders[i].CreatedAt
			}
		}
	}

	logger.Info(nil, "dedup_orders_remove", "cluster", "duplicates identified", utils.LoggerMeta{
		"duplicate_groups": dupGroups,
		"to_remove":        len(toRemove),
	})

	// report or apply.
	if dedupDryRun {
		logger.Info(nil, "dedup_orders_remove", "dry_run", "no changes applied (dry-run)", utils.LoggerMeta{
			"start":            start.Format(time.RFC3339),
			"end":              end.Format(time.RFC3339),
			"scanned":          scanned,
			"duplicate_groups": dupGroups,
			"would_remove":     len(toRemove),
		})
		return nil
	}

	removed := 0
	for _, id := range toRemove {
		if err := orderSvc.RemoveOrder(id); err != nil {
			if err == sql.ErrNoRows {
				logger.Debug(nil, "dedup_orders_remove", "remove", "order already removed, skipping", utils.LoggerMeta{"order_id": id.String()})
				continue
			}
			logger.Error(nil, "dedup_orders_remove", "remove", "failed to remove order", utils.LoggerMeta{"order_id": id.String()}, err.Error())
			return fmt.Errorf("failed to remove order %s: %w", id, err)
		}
		logger.Debug(nil, "dedup_orders_remove", "remove", "soft-deleted order", utils.LoggerMeta{"order_id": id.String()})
		removed++
	}

	logger.Info(nil, "dedup_orders_remove", "done", "dedup complete", utils.LoggerMeta{
		"start":            start.Format(time.RFC3339),
		"end":              end.Format(time.RFC3339),
		"scanned":          scanned,
		"duplicate_groups": dupGroups,
		"removed":          removed,
	})
	return nil
}
