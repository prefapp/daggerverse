// A generated module for Common functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
    
  "fmt"
  "context"

  yaml "gopkg.in/yaml.v3"

)

type BuildImages struct{}

type BuildData struct {

  buildArgs map[string]string

  dockerfile string

  tag string

}


func (m* BuildImages) LoadInfo(ctx context.Context, yamlPath *File) string {

  val, err :=  yamlPath.Contents(ctx)

  buildData := BuildData{}

  if err != nil {

    panic(fmt.Sprintf("Loading yaml: %s", val))
  
  }else{

    return yaml.Unmarshall([]byte(val), &buildData)

  }
    
}

