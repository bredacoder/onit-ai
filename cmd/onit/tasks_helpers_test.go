package main

import (
	"bytes"

	"github.com/bredacoder/onit-ai/internal/core"
	"github.com/bredacoder/onit-ai/internal/core/ids"
)

// runTasksCmd wires a root command with the given persistence layer and userID,
// sets args to "tasks", captures stdout and stderr into separate buffers,
// and returns (outBuf, errBuf, executeError).
func runTasksCmd(p core.Persistence, userID ids.UserID) (string, string, error) {
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	cmd := newRootCmd(p, userID)
	cmd.SetArgs([]string{"tasks"})
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()

	return outBuf.String(), errBuf.String(), err
}
