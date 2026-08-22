package cmd

import (
	"github.com/JayPonda/Product-catalog/server/utils"
	"github.com/spf13/cobra"
)

// registerAppCommands is the single place where every application/custom CLI
// command is wired into rootCmd. Each command is a factory function that
// receives the shared config and logger (and may receive any other
// pre-initialized dependency it needs), so registration stays decoupled from
// where each command file lives. Infra commands (server, migrate) register
// themselves and are intentionally excluded here.
func registerAppCommands(root *cobra.Command, cfg AppConfig, logger *utils.StructuredLogger) {
	root.AddCommand(seedCmd(cfg, logger))
	root.AddCommand(dedupOrdersRemoveCmd(cfg, logger))
}
