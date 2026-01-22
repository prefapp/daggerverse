package main

import (
	"context"
	"dagger/ci/internal/dagger"
    "strings"

    "gopkg.in/yaml.v3"
)

type Ci struct{}

func ReadTaskFile(ctx context.Context, taskFile *dagger.File) CiData {
    yamlContent, err := taskFile.Contents(ctx)
    if err != nil {
        panic(err)
    } 

    ciData := CiData{}

    err = yaml.Unmarshal([]byte(yamlContent), &ciData)
    if err != nil {
        panic(err)
    } 

    return ciData
}

func (m *Ci) ExecuteTask(ctx context.Context, task CiTask, container *dagger.Container) *dagger.Container {
    if task.Run == "" {
        container.WithExec(strings.Split(task.Run, " "))
    }

    return container
}

func (m *Ci) ExecuteCI(ctx context.Context, taskFile *dagger.File) *dagger.Container {
    ciData := ReadTaskFile(ctx, taskFile)

    dockerImage := ciData.Setup.Technology + ":" + ciData.Setup.Version

    container := dag.Container().From(dockerImage)

    for _, task := range ciData.Defaults.Tasks {
        container = m.ExecuteTask(ctx, task, container)
    }

    return container
}

