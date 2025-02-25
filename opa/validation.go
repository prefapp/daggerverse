package main

import (
	"context"
	"dagger/opa/internal/dagger"
	"fmt"
	"regexp"
	"time"
)

func (m *Opa) Validate(
	ctx context.Context,
	policy *dagger.File,
	data *dagger.File,
	file *dagger.File,
) (*dagger.Container, error) {

	fileName, err := file.Name(ctx)

	if err != nil {

		return nil, err

	}

	policyFileName, err := policy.Name(ctx)

	if err != nil {

		return nil, err

	}

	dataFileName, err := data.Name(ctx)

	if err != nil {

		return nil, err

	}

	fmt.Printf("Validating file %s against policy %s\n", fileName, policyFileName)

	ctr, _ := dag.Container().
		From("openpolicyagent/conftest").
		WithMountedFile(fmt.Sprintf("/validation/%s", policyFileName), policy).
		WithMountedFile(fmt.Sprintf("/validation/%s", dataFileName), data).
		WithMountedFile(fmt.Sprintf("/validation/%s", fileName), file).
		WithWorkdir("/validation").
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithExec([]string{
			"conftest",
			"--rego-version", "v1",
			"--output", "stdout",
			"test", fileName,
			"--data", dataFileName,
			"--policy", policyFileName,
		}, dagger.ContainerWithExecOpts{
			RedirectStderr: "/tmp/stderr",
			RedirectStdout: "/tmp/stdout",
			Expect:         "ANY",
		}).Sync(ctx)

	eC, err := ctr.ExitCode(ctx)

	if err != nil {

		return nil, err

	}

	if eC != 0 {

		stderr, _ := ctr.File("/tmp/stderr").Contents(ctx)
		stdout, _ := ctr.File("/tmp/stdout").Contents(ctx)

		return nil, fmt.Errorf("%s\n%s", strip(stdout), strip(stderr))
	}

	return ctr, nil
}

func strip(str string) string {
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

	var re = regexp.MustCompile(ansi)

	str = re.ReplaceAllString(str, "")

	str = regexp.MustCompile(`\n`).ReplaceAllString(str, "")

	return str
}
