package main

import (
	"dagger/ci/internal/dagger"
	"fmt"
)

func Runner(taskId string, taskMap map[string]Task) {

	task := taskMap[taskId]

	ctx, err := ResolveContext(task, taskMap)

	if err != nil {

		panic(err)
	}

	task.Ctx = ctx

	fmt.Printf("Running task %s\n", task.Id)
}

func ResolveContext(task Task, taskMap map[string]Task) (*dagger.Container, error) {

	if task.HasDependencies {

		for _, dependencyId := range task.Definition.Needs {

			switch status := taskMap[dependencyId].Status; status {
			case "PENDING":
				return nil, fmt.Errorf("Task '%s' requires task %s to be completed first", task.Id, dependencyId)
			case "RUNNING":
				return nil, fmt.Errorf("Task '%s' requires task %s to be completed first, is in pending state", task.Id, dependencyId)
			case "FAILED":
				return nil, fmt.Errorf("Task '%s' requires task %s to be completed first, is in failed state", task.Id, dependencyId)
			case "COMPLETED":
				fmt.Printf("Task %s has completed, running task %s\n", dependencyId, task.Id)
			}

		}
		// return the last container in the dependency chain
		return taskMap[task.Definition.Needs[len(task.Definition.Needs)-1]].Ctx, nil

	} else {

		return dag.Container().From(task.Definition.Image), nil

	}

}
