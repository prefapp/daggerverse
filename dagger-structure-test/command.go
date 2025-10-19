package main

import (
	"context"
	"regexp"
	"dagger/dagger-structure-test/internal/dagger"
	"time"
)

type AssertOutputOpts struct {
	ExpectedOutput *string
	ExpectedError  *string
	ExpectedExitCode *int
	DisableCache *bool
}

// Test a container output
func (m *DaggerStructureTest) AssertOutput(ctx context.Context, container *dagger.Container, options *AssertOutputOpts) (bool, error) {

	ctr := container

	if options.DisableCache != nil && *options.DisableCache {
		ctr = container.WithEnvVariable("CACHEBUSTER", time.Now().String())
	}

	if options.ExpectedOutput != nil {
		output, err := ctr.Stdout(ctx)

		if _, ok := err.(*ExecError); err != nil && !ok {
			return false, err
		}
		if output != *options.ExpectedOutput {
			return false, nil
		}

		match, err := regexp.MatchString(*options.ExpectedOutput, output)

		if err != nil {
			return false, err
		}

		if !match {
			return false, nil
		}
	}

	if options.ExpectedError != nil {
		output, err := ctr.Stderr(ctx)
		
		if _, ok := err.(*ExecError); err != nil && !ok {
			return false, err
		}
		
		match, err := regexp.MatchString(*options.ExpectedError, output)
		
		if err != nil {
			return false, err
		}
		
		if !match {
			return false, nil
		}
		
	}
	
	if options.ExpectedExitCode != nil {
		exitCode, err := m.GetExitCode(ctx, ctr)
		if err != nil {
			return false, err
		}
		if exitCode != *options.ExpectedExitCode {
			return false, nil
		}
	}
	
	return true, nil
}

// Returns a container exit code
func (m *DaggerStructureTest) GetExitCode(ctx context.Context, ctr *dagger.Container) (int, error) {

	_, err := ctr.Sync(ctx)
    if e, ok := err.(*ExecError); ok {
        return e.ExitCode, nil
    } else if err != nil {
		return 0, err
	}
    return 0, nil
}
