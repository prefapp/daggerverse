package main

import (
	"dagger/ci/internal/dagger"
)

type Config struct {
	GlobalImage   string            `yaml:"image"`
	GlobalVars    map[string]string `yaml:"vars"`
	GlobalSecrets []string          `yaml:"secrets"`
	Tasks         []TaskDeclaration `yaml:"tasks"`
}

type TaskDeclaration struct {
	Name      string            `yaml:"name"`
	Image     string            `yaml:"image"`
	Vars      map[string]string `yaml:"vars"`
	Secrets   map[string]string `yaml:"secrets"`
	RunScript string            `yaml:"run"`
	Needs     []string          `yaml:"needs"` // previous tasks that need to be run before this task
}

type Task struct {
	Id              string
	Definition      TaskDeclaration
	Ctx             *dagger.Container
	HasDependencies bool
	Status          string // "Running", "Success", "Failed"
}

type TaskOrder struct {
	CurrentTask int
	Tasks       []Task
	TaskMap     map[string]Task
}
