/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/spf13/cobra"
)

// AppConfig is the minimal configuration contract the CLI needs from the host
// application. It embeds utils.DBConfigProvider (for DB access / migrations)
// and adds the HTTP host/port getters used by the server command.
type AppConfig interface {
	utils.DBConfigProvider
	GetHost() string
	GetPort() string
	GetAllowedOrigins() string
	GetDialect() string
}

// Shared dependencies injected by main.main() before the root command executes.
// Command files in this package read these instead of re-parsing the
// environment, keeping a single source of truth for config and logging.
var (
	appConfig AppConfig
	appLogger *utils.StructuredLogger
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Product Catalog server CLI",
	Long:  `Product Catalog server CLI built with Cobra. Use "server" to start the HTTP server and "migrate" to manage database migrations.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
//
// To add a NEW CLI command in the future:
//  1. Create a new file in this directory, e.g. commands/cmd/seed.go
//  2. Define `var seedCmd = &cobra.Command{ Use: "seed", ... }`
//  3. Register it in init(): `func init() { rootCmd.AddCommand(seedCmd) }`
//  4. Use the package-level `appConfig` / `appLogger` for shared dependencies.
func Execute(cfg AppConfig, logger *utils.StructuredLogger) {
	appConfig = cfg
	appLogger = logger

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Subcommands register themselves via their own init() (see serve.go, migrate.go).
	// To add a new command in the future:
	//  1. Create a new file, e.g. commands/cmd/seed.go
	//  2. Define `var seedCmd = &cobra.Command{ Use: "seed", ... }`
	//  3. Register it: `func init() { rootCmd.AddCommand(seedCmd) }`
	//  4. Use the package-level `appConfig` / `appLogger` for shared dependencies.
}
