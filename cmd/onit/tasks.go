package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/bredacoder/onit-ai/internal/core"
	"github.com/bredacoder/onit-ai/internal/core/ids"
	"github.com/bredacoder/onit-ai/internal/core/understanding"
)

// errNoCurrentUser is returned when ONIT_USER_ID is not set.
var errNoCurrentUser = errors.New("no current user configured: set ONIT_USER_ID env var")

// newTasksCmd returns a cobra.Command that lists tasks for the given user.
// Rendering goes to cmd.OutOrStdout() so tests can capture output with
// cmd.SetOut(buf).
func newTasksCmd(p core.Persistence, userID ids.UserID) *cobra.Command {
	return &cobra.Command{
		Use:   "tasks",
		Short: "List your tasks",
		Long:  "List all tasks belonging to the current user.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if userID == "" {
				return errNoCurrentUser
			}

			tasks, err := p.ListTasks(cmd.Context(), userID)
			if err != nil {
				return fmt.Errorf("listing tasks: %w", err)
			}

			return renderTasks(cmd.OutOrStdout(), tasks)
		},
	}
}

// renderTasks writes the task list (or empty-state message) to w.
// Format (one line per task):
//
//	<id>  <service_type>  <state>
func renderTasks(w io.Writer, tasks []understanding.Task) error {
	if len(tasks) == 0 {
		_, err := fmt.Fprintln(w, "no tasks yet")
		return err
	}

	for _, t := range tasks {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", t.ID, t.ServiceType, t.State); err != nil {
			return err
		}
	}

	return nil
}
