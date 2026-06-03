// Package main is the composition root for the onit CLI. It is the only place
// in the codebase that imports both core and adapters (hexagonal invariant).
// Keep this file thin: read env, build infrastructure, wire cobra, execute.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	pgadapter "github.com/bredacoder/onit-ai/internal/adapters/postgres"
	"github.com/bredacoder/onit-ai/internal/core"
	"github.com/bredacoder/onit-ai/internal/core/ids"
)

func main() {
	rootCmd, err := buildCmd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// buildCmd constructs the cobra root command with all subcommands wired up.
// It reads DATABASE_URL and ONIT_USER_ID from the environment, builds a real
// postgres adapter, and returns the ready-to-Execute root command.
//
// Separating buildCmd from main() keeps the logic accessible to tests that
// want to exercise the full command tree without os.Exit.
func buildCmd() (*cobra.Command, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL env var is required")
	}

	rawUserID := os.Getenv("ONIT_USER_ID")
	if rawUserID == "" {
		return nil, fmt.Errorf("ONIT_USER_ID env var is required")
	}
	userID := ids.UserID(rawUserID)

	// Use a timeout for initial pool connection to avoid hanging indefinitely.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	adapter := pgadapter.New(pool)

	return newRootCmd(adapter, userID), nil
}

// newRootCmd assembles the cobra root command from its dependencies. All
// command construction goes here so tests can pass alternate implementations
// without touching env or the real database.
func newRootCmd(p core.Persistence, userID ids.UserID) *cobra.Command {
	root := &cobra.Command{
		Use:   "onit",
		Short: "Personal AI agent for local-services tasks",
		Long: `onit is your personal AI agent — it finds local-service providers,
gets quotes, cross-checks your calendar, and schedules on your behalf.`,

		// SilenceUsage prevents cobra from printing usage on every RunE error.
		// SilenceErrors prevents cobra's own error printing; we handle it in main.
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newTasksCmd(p, userID))

	return root
}
