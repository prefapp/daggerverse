package main

import (
	"context"
	"fmt"
)

func PrepareAndPrintError(
	ctx context.Context,
	command string,
	description string,
	errorCaused error,
) error {

	errorM := fmt.Errorf(`
# Error
## 🛑 CRITICAL FAILURE: %s,

%s,

----

Error details:

%s
`,
		command, description, errorCaused,
	)

	return errorM

}
