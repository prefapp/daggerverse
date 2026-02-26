package main

import (
	"context"
	"fmt"
	"gh/internal/dagger"
	"strings"
	"time"
)

var WaitTimeBetweenRetries time.Duration = 2 * time.Second

func extractErrorMessage(err error) string {
	switch e := err.(type) {
	case *dagger.ExecError:
		errorMsg := ""

		if e.Stderr != "" {
			errorMsg += fmt.Sprintf("::error::%s\n", e.Stderr)
		}
		if e.Stdout != "" {
			errorMsg += fmt.Sprintf("::info::%s", e.Stdout)
		}

		return errorMsg
	default:
		return fmt.Sprintf("::error::%s", strings.ReplaceAll(err.Error(), "::error::", ""))
	}
}

func retry(
	ctx context.Context,
	returnError bool,
	errorToReturn error,
	msg string,
) error {
	if returnError {
		return errorToReturn
	}

	fmt.Println(msg)

	timer := time.NewTimer(WaitTimeBetweenRetries)
	select {
	case <-timer.C:
		timer.Stop()
	case <-ctx.Done():
		timer.Stop()
		return ctx.Err()
	}

	return nil
}
