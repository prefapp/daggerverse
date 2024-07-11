// A generated module for K6 functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type K6 struct{}

// EnvironmentVariable represents a string that follows the pattern [a-zA-Z0-9]+=[a-zA-Z0-9]+
type EnvironmentVariable string

// IsValid checks if the EnvironmentVariable adheres to the specified format
func (cs EnvironmentVariable) IsValid() bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9_\-]+=[a-zA-Z0-9_\-]+$`)
	return re.MatchString(string(cs))
}

// Parse returns the key and value of the EnvironmentVariable
func (cs EnvironmentVariable) Parse() (string, string) {
	parts := strings.Split(string(cs), "=")
	if len(parts) != 2 {
		panic("Error parsing environment variable")
	}
	return parts[0], parts[1]
}

// Returns lines that match a pattern in the files of the provided Directory
func (m *K6) Run(
	ctx context.Context,
	// The working directory containing the script
	//+required
	workingDir *Directory,
	// k6 Script file to execute
	//+required
	script string,
	// Direcetory to store the results
	//+required
	outputDir string,
	// Environment variables to set
	//+optional
	env []EnvironmentVariable,
	// Virtual users to emulate
	//+optional
	//+default=1
	vus int,
	// Duration of the test
	//+optional
	//+default="1s"
	duration string,
) *Container {
	// We use Glob over Entries because it lists files recursively
	entries, _ := workingDir.Glob(ctx, "*")

	if !slices.Contains(entries, script) {
		panic("script not found")
	}

	workingDirMountPath := "/k6"
	outputDirMountPath := "/output"

	ctr := dag.Container().
		From("ghcr.io/grafana/xk6-dashboard")

	for _, e := range env {
		if !e.IsValid() {
			panic("invalid env variable")
		}

		key, value := e.Parse()

		ctr = ctr.WithEnvVariable(key, value)
	}

	ctr = ctr.WithDirectory(workingDirMountPath, workingDir).
		WithDirectory(outputDirMountPath, dag.Directory()).
		WithUser("root").
		WithExec([]string{
			"run",
			"--vus", strconv.Itoa(vus),
			"--duration", duration,
			"--out", fmt.Sprintf("web-dashboard=export=%s", filepath.Join(outputDirMountPath, "report.html")),
			"--summary-export", filepath.Join(outputDirMountPath, "summary.json"),
			"--console-output", filepath.Join(outputDirMountPath, "logs.txt"),
			filepath.Join(workingDirMountPath, script),
		})

	ctr, err := ctr.Sync(ctx)

	if e, ok := err.(*ExecError); ok {
		panic("Error running QA workflow, exit code: " + strconv.Itoa(e.ExitCode))
	} else if err != nil {
		panic("Unexpected error")
	}

	return ctr
}
