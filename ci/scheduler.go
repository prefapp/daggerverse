package main

import (
	"context"
	"dagger/ci/internal/dagger"

	"gopkg.in/yaml.v3"
)

// Returns a container that echoes whatever string argument is provided
func (m *Ci) Scheduler(ctx context.Context, tasksFile *dagger.File) string {
	content, err := tasksFile.Contents(ctx)

	if err != nil {

		panic(err)
	}

	config := Config{}

	err = yaml.Unmarshal([]byte(content), &config)

	taskOrder := TaskOrder{}

	for _, taskDeclaration := range config.Tasks {

		task := Task{
			Id:              taskDeclaration.Name,
			Definition:      taskDeclaration,
			HasDependencies: len(taskDeclaration.Needs) > 0,
			Ctx:             nil,
			Status:          "PENDING",
		}

		taskOrder.Tasks = append(taskOrder.Tasks, task)

		taskOrder.TaskMap[task.Id] = task

	}

	for _, task := range taskOrder.Tasks {

		Runner(task.Id, taskOrder.TaskMap)

	}

	return "Scheduler ran successfully"
}
