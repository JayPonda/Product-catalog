package cmd

import (
	"fmt"
	"time"

	"github.com/JayPonda/Product-catalog/server/src/repositories"
	"github.com/JayPonda/Product-catalog/server/src/services"
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/spf13/cobra"
)

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
			ctx := utils.NewCLIContext()

			// part 2: define start and end time for this session.
			var start, end time.Time

			if dedupStart != "" && dedupEnd != "" {
				// manual mode: both --starttime and --endtime supplied (RFC3339).
				var err error
				if start, err = time.Parse(time.RFC3339, dedupStart); err != nil {
					logger.Error(ctx, "dedup_orders_remove.go", "RunE", "invalid --starttime", utils.LoggerMeta{"value": dedupStart}, err.Error())
					return fmt.Errorf("invalid --starttime %q: %w", dedupStart, err)
				}
				if end, err = time.Parse(time.RFC3339, dedupEnd); err != nil {
					logger.Error(ctx, "dedup_orders_remove.go", "RunE", "invalid --endtime", utils.LoggerMeta{"value": dedupEnd}, err.Error())
					return fmt.Errorf("invalid --endtime %q: %w", dedupEnd, err)
				}
				if !end.After(start) {
					logger.Error(ctx, "dedup_orders_remove.go", "RunE", "end must be after start", utils.LoggerMeta{"start": dedupStart, "end": dedupEnd}, "")
					return fmt.Errorf("--endtime %q must be after --starttime %q", dedupEnd, dedupStart)
				}
			} else {
				// scheduled mode: end = now, start = now - window (2.5h default,
				// 2h interval + 0.5h overlap so boundary-straddling clusters are caught).
				now := time.Now()
				end = now
				start = now.Add(-dedupWindow)
			}

			db := utils.InitDB(cfg)

			orderRepo, err := repositories.InitOrderRepository(db, logger)
			if err != nil {
				logger.Error(ctx, "dedup_orders_remove.go", "RunE", "failed to init order repository", nil, err.Error())
				return err
			}
			orderSvc, err := services.InitOrderService(db, logger, orderRepo)
			if err != nil {
				logger.Error(ctx, "dedup_orders_remove.go", "RunE", "failed to init order service", nil, err.Error())
				return err
			}

			dCfg := services.DedupConfig{
				DryRun: dedupDryRun,
				Nearby: dedupNearby,
				Window: dedupWindow,
				Start:  dedupStart,
				End:    dedupEnd,
				Batch:  dedupBatch,
			}

			return services.RunDedupOrdersRemove(ctx, start, end, orderSvc, dCfg, logger)
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
