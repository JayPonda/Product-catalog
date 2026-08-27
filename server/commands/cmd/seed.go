package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

// seedNamespace is a fixed UUID used to derive deterministic v5 UUIDs, so the
// seed is idempotent (same input -> same UUID -> ON CONFLICT DO NOTHING).
var seedNamespace = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479")

// seedDefaultPassword is the bcrypt-hashed password assigned to every seeded
// user. Login is not the goal of this seed; use this only for local dev.
const seedDefaultPassword = "Password123!"

var seedCSV string

// seedCmd builds the seed command. It receives the shared config and logger as
// arguments so the command pulls envs/context from configuration rather than
// reaching for globals. It reads orders.csv, ensures one user per customer code,
// then seeds the orders table referencing those users. It also writes enriched
// CSVs (users_seed.csv, orders_enriched.csv) containing the resolved UUIDs.
func seedCmd(cfg AppConfig, logger *utils.StructuredLogger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seed",
		Short: "Seed users and orders from a CSV file",
		RunE: func(cmd *cobra.Command, args []string) error {
			db := utils.InitDB(cfg)

			logger.Info("seed.go", "RunE", "seed started", utils.LoggerMeta{"csv": seedCSV})

			// --- 1. read the source CSV ---
			f, err := os.Open(seedCSV)
			if err != nil {
				logger.Error("seed.go", "RunE", "failed to open csv", utils.LoggerMeta{"path": seedCSV}, err.Error())
				return fmt.Errorf("failed to open csv %q: %w", seedCSV, err)
			}
			defer f.Close()

			reader := csv.NewReader(f)
			reader.FieldsPerRecord = -1
			rows, err := reader.ReadAll()
			if err != nil {
				logger.Error("seed.go", "RunE", "failed to parse csv", nil, err.Error())
				return fmt.Errorf("failed to parse csv: %w", err)
			}
			if len(rows) < 2 {
				logger.Warn("seed.go", "RunE", "csv has no data rows", nil)
				return fmt.Errorf("csv has no data rows")
			}
			header := rows[0]
			if len(header) < 3 {
				logger.Warn("seed.go", "RunE", "csv missing required columns", nil)
				return fmt.Errorf("csv must have at least: customer_id, order_value, timestamp")
			}

			// --- 2. prepare enriched CSV outputs ---
			dir := filepath.Dir(seedCSV)
			usersOut, err := os.Create(filepath.Join(dir, "users_seed.csv"))
			if err != nil {
				return err
			}
			defer usersOut.Close()
			usersW := csv.NewWriter(usersOut)
			if err := usersW.Write([]string{"id", "first_name", "last_name", "email"}); err != nil {
				return err
			}

			ordersOut, err := os.Create(filepath.Join(dir, "orders_enriched.csv"))
			if err != nil {
				return err
			}
			defer ordersOut.Close()
			ordersW := csv.NewWriter(ordersOut)
			if err := ordersW.Write([]string{"order_id", "user_id", "customer_id", "total_bill", "created_at", "updated_at"}); err != nil {
				return err
			}

			passwordHash, err := bcrypt.GenerateFromPassword([]byte(seedDefaultPassword), bcrypt.DefaultCost)
			if err != nil {
				return err
			}

			userUUIDs := make(map[string]uuid.UUID)

			var usersCreated, ordersCreated int

			// --- 3. ETL: per row, ensure user then seed order ---
			for _, row := range rows[1:] {
				if len(row) < 3 {
					continue
				}
				code := row[0]
				orderValue := row[1]
				timestamp := row[2]

				// derive stable UUIDs
				userID := uuid.NewSHA1(seedNamespace, []byte("user:"+code))
				orderID := uuid.NewSHA1(seedNamespace, []byte("order:"+code+"@"+timestamp))

				// seed user (idempotent)
				if _, seen := userUUIDs[code]; !seen {
					res, insErr := db.Insert("users").Rows(goqu.Record{
						"id":         userID,
						"first_name": code,
						"last_name":  code,
						"email":      code + "@seed.local",
						"password":   string(passwordHash),
					}).OnConflict(goqu.DoNothing()).Executor().Exec()
					if insErr != nil {
						return fmt.Errorf("failed to seed user %q: %w", code, insErr)
					}
					if affected, _ := res.RowsAffected(); affected > 0 {
						usersCreated++
					}
					userUUIDs[code] = userID
					if err := usersW.Write([]string{userID.String(), code, code, code + "@seed.local"}); err != nil {
						return err
					}
				}

				// parse values
				total, perr := strconv.ParseFloat(orderValue, 64)
				if perr != nil {
					return fmt.Errorf("invalid order_value %q: %w", orderValue, perr)
				}
				ts, terr := time.Parse(time.RFC3339, timestamp)
				if terr != nil {
					return fmt.Errorf("invalid timestamp %q: %w", timestamp, terr)
				}

				// seed order (idempotent)
				res, insErr := db.Insert("orders").Rows(goqu.Record{
					"id":          orderID,
					"customer_id": userUUIDs[code],
					"total_bill":  total,
					"created_at":  ts,
					"updated_at":  ts,
				}).OnConflict(goqu.DoNothing()).Executor().Exec()
				if insErr != nil {
					return fmt.Errorf("failed to seed order for %q: %w", code, insErr)
				}
				if affected, _ := res.RowsAffected(); affected > 0 {
					ordersCreated++
				}

				if err := ordersW.Write([]string{
					orderID.String(),
					userUUIDs[code].String(),
					code,
					orderValue,
					ts.Format(time.RFC3339),
					ts.Format(time.RFC3339),
				}); err != nil {
					return err
				}
			}

			usersW.Flush()
			ordersW.Flush()

			logger.Info("seed.go", "RunE", "seed completed", utils.LoggerMeta{"users_created": usersCreated, "orders_created": ordersCreated})
			return nil
		},
	}

	cmd.Flags().StringVar(&seedCSV, "csv", "tmp/orders.csv", "path to the orders CSV file")

	return cmd
}
