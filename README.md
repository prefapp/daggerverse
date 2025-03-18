[![Pr verify](https://github.com/prefapp/daggerverse/actions/workflows/pr_verify.yaml/badge.svg)](https://github.com/prefapp/daggerverse/actions/workflows/pr_verify.yaml)

# Prefapp daggerverse

### Requirements to develop
- dagger v0.15.3
- go 1.23.4+

### Develop an existing dagger module

```shell
cd build-images

# This will install the internal dagger packages, that are ignored in `.gitignore`
dagger develop
```

### Create a new dagger module

```shell
# Create a folder for a new module
mkdir <my-module-name>
cd <my-module-name>

# Create a module in the source ".", it will follow the monorepo structure for the .gitignore.
dagger init --name=<my-module-name> --sdk=go --source=.
```
