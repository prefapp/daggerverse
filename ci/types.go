package main

type CiTask struct {
    Name       string            `yaml:"name"`
    Needs      []string          `yaml:"needs"`
    WorkingDir string            `yaml:"working_dir"`
    Run        string            `yaml:"run"`
    Uses       string            `yaml:"uses"`
    Function   string            `yaml:"function"`
    // We must express this property as a string
    // With       map[string]string `yaml:"with"`
    Image      string            `yaml:"image"`
}

type CiTaskList struct {
    Tasks      []CiTask          `yaml:"tasks"`
}

type CiSetup struct {
    Technology string            `yaml:"technology"`
    Version    string            `yaml:"version"`
}

type CiData struct {
    Setup      CiSetup           `yaml:"setup"`
    Defaults   CiTaskList        `yaml:"defaults"`
}
