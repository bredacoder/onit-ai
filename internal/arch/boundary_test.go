// Package arch_test enforces architecture invariants for the onit module.
// It asserts that no package under internal/core/... imports any package
// under internal/adapters/..., ensuring the hexagonal boundary holds.
package arch_test

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

const (
	corePattern   = "github.com/bredacoder/onit-ai/internal/core/..."
	adapterPrefix = "github.com/bredacoder/onit-ai/internal/adapters/"
)

func TestCoreDoesNotImportAdapters(t *testing.T) {
	t.Parallel()

	out, err := runGoList(corePattern)
	if err != nil {
		t.Fatalf("go list failed: %v", err)
	}

	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			t.Fatalf("unexpected go list output line: %q", line)
		}
		pkgPath := parts[0]
		allImports := parts[1] + " " + parts[2] + " " + parts[3]

		for _, imp := range strings.Fields(allImports) {
			if strings.HasPrefix(imp, adapterPrefix) {
				t.Errorf("core package %q imports adapter %q — dependency inversion violated", pkgPath, imp)
			}
		}
	}
}

// runGoList executes go list for the given pattern and returns the combined output.
func runGoList(pattern string) (string, error) {
	cmd := exec.Command(
		"go", "list",
		"-f", "{{.ImportPath}}|{{join .Imports \" \"}}|{{join .TestImports \" \"}}|{{join .XTestImports \" \"}}",
		pattern,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("go list %q: %w\nstderr: %s", pattern, err, stderr.String())
	}

	return stdout.String(), nil
}
