package main

import (
	"context"
	"testing"
)

// Test root directory existence
func TestAssertOutput(t *testing.T) {

	m := DaggerStructureTest{}
	ctx := context.Background()

	ctr := dag.Container().From("alpine")

	expectedOutput := "root\n"
	expectedExitCode := 0

	res, err := m.AssertOutput(ctx, ctr.WithExec([]string{"whoami"}), &AssertOutputOpts{
		ExpectedOutput: &expectedOutput,
		ExpectedExitCode: &expectedExitCode,
	})

	if err != nil || !res {
		t.Fatalf(`Default alpine user should be root`)
	}

	expectedExitCode = 1

	res, err = m.AssertOutput(ctx, ctr.WithExec([]string{"ls", "/foo"}), &AssertOutputOpts{
		ExpectedExitCode: &expectedExitCode,
	})

	if err != nil || !res {
		t.Fatalf(`Expected exit code 1`)
	}

}
