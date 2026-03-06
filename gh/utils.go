package main

import (
	"context"
	"fmt"
	"gh/internal/dagger"
	"strings"
	"time"
)

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
	waitTime time.Duration,
) error {
	if returnError {
		return errorToReturn
	}

	fmt.Println(msg)

	timer := time.NewTimer(waitTime)
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		timer.Stop()
		return ctx.Err()
	}
}
