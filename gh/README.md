# 🚀 Dagger GH module

This module is a custom version of [aweris' `gh` module](https://github.com/aweris/daggerverse), with some custom functionality added. This documentation will explain both our custom functions and the already existing ones from the base module.

The general purpouse of the module is to allow easy usage of the `gh` cli program within Dagger, especially when paired with our [`gh-commit` plugin](https://github.com/prefapp/gh-commit) (though any number and kind of plugins can be installed). The module has specialized functions designed to do specific tasks, but it also has some generic functions which require additional programming or user input to achieve an useful result.

## 🔨 Functions

NOTE: Most functions will accept [generic params](#generic-params) in their calls. In order to avoid repeating them within each function explanation, they'll be given their own section and linked whenever needed.

Unless otherwise specified, any given function parameter is optional.

1. **New**: generic function to create a `Gh` golang object. Note this only creates and returns a pointer to a `Gh` golang object, **_not_** a Dagger container, and it's currently not used in favor of more useful functions that return containers. It receives the [generic params `version`, `token` and `repo`](#generic-params). In addition, it has the following specific params:
    - `plugins`: a `[]string`, representing the name of the plugins that should be installed when creating the container. Omitting this parameter (via the command line) or setting it to an empty list (via code) won't install any plugins.
    - `base`: a `*dagger.Container`, to be used as the base of a new container created by using the `Gh` object returned by this function. Omitting this parameter (via the command line) or setting it to an empty list (via code) will create a container using the `alpine/git:latest` image.

2. **Container**: generic function that creates a new `*dagger.Container` and returns it, along with an error if any happened. The container created is customized using the function parameters, which consist of all [generic params](#generic-params) and the following specific ones:
    - `repo`: a `string`, the name of the repo the application is working with, in the format `<owner>/<repo-name>`. This value is used to set the `GH_REPO` environment variable.
    - `pluginNames`: a `[]string`, a list of plugin names to be installed. Omitting or not setting this value will result in no plugins being installed.
    - `pluginVersions`: a `[]string`, a list of versions for each plugin in `pluginNames`. Leaving a version blank or not setting it will download the latest version for that plugin. The values in this parameter will be associated to `pluginNames` positionally, i.e. `pluginNames = [a, b, c, d, e]` and `pluginVersions = [1, "", 2]` will result in the following plugins being downloaded: `a-1`, `b-latest`, `c-2`, `d-latest` and `e-latest`.

3. **Run**: generic function that runs an arbitrary `gh` cli command inside a container created via the previous **Container** (2) function, then returns it and any errors that happened. This function takes  all [generic params](#generic-params), which are then passed to the **Container** (2) function for customization, and the following specific ones:
    - `cmd`: a `string`, representing a `gh` command to be executed inside the container. Any valid `gh` command is supported, and parameters and subcommands can be added by separating them with spaces as usual. The initial `gh` part of the command must be omitted (e.g., in order to execute `gh pr list` the value of `cmd` must be `pr list`). This parameter is *mandatory*.
    - `disableCache`: a `bool`, used to disable Dagger's container cache when is set to `true` by using a `CACHE_BUSTER` environment variable.

4. **Get**: generic function that downloads and returns the `gh` cli binary as a `*dagger.File`, as well as any errors that happen. This function takes the [`ctx`, `version` and `token` generic params](#generic-params), and the following specific ones:
    - `goos`: a `string`, representing the operating system we want to download the binary for. Uses the OS the program is running on when not set. Can be either `linux` or `darwin`.
    - `goarch`: a `string`, representing the architecture we want to download the binary for. Uses the architecture of the PC the program is running on when not set. Can be any valid `gh` architecture (`amd64`, `arm64`, etc)

5. **CreatePR**: specific function used to create a PR in a GitHub repository, returning the link to the PR and any errors that occurred. The PR is created using the `gh pr create` command, and the link is retrieved by a combination of `gh pr list` and `gh pr view`. This function takes the [all generic params](#generic-params), and the following specific ones:
    - `title`: a `string`, the title of the PR to be created. This parameter is *mandatory*.
    - `body`: a `string`, the body of the PR to be created. This parameter is *mandatory*.
    - `branch`: a `string`, the name of the branch from which to create the PR. This parameter is *mandatory*.
    - `repoDir`: a `*dagger.Directory`, the folder where the repository the PR is going to be created in is stored. This parameter is *mandatory*.
    - `labels`: a `[]string`, a list of labels to add to the PR once it has been created. No labels will be added if this parameter is omitted or empty.
    - `ctr`: a `*dagger.Container`, which is used to launch the `gh` commands. If this is omitted or unset, a new container will be created and used instead.

6. **Commit**: specific function used to commit and push changes in a repo to the remote, using our `gh-commit` `gh` plugin. Returns the container used to execute the command plus any errors that occurred. This function takes the [all generic params](#generic-params) and the following specific ones:
    - `repoDir`: a `*dagger.Directory`, the folder where the repository the commit is going to be made in is stored. This parameter is *mandatory*.
    - `branchName`: a `string`, the name of the branch where to commit the changes. This parameter is *mandatory*.
    - `commitMessage`: a `string`, the text of the commit message. This parameter is *mandatory*.
    - `deletePath`: a `string`, the value of the `--delete-path` parameter of the `gh-commit` plugin. Setting this will delete all the files in the specified path before creating the commit.
    - `createEmpty`: a `bool`, whether or not to allow the creation of an empty commit (a commit with no files changed). By default, `false`, so trying to create a commit with no changes will throw an error.
    - `ctr`: a `*dagger.Container`, which is used to launch the `gh` commands. If this is omitted or unset, a new container will be created and used instead.

7. **CommitAndCreatePR**: specific function used to close any PRs that already exist in the current branch, commit the current latest changes and create a new PR with them. Uses the **CreatePR** (5), **Commit** (6) and **DeleteRemoteBranch** (8) functions. Returns the link of the created PR and any errors that occurred. This function takes the [all generic params](#generic-params) and the following specific ones:
    - `repoDir`: a `*dagger.Directory`, the folder where the repository the commit is going to be made in is stored. This parameter is *mandatory*.
    - `branchName`: a `string`, the name of the branch where to commit the changes. This parameter is *mandatory*.
    - `commitMessage`: a `string`, the text of the commit message. This parameter is *mandatory*.
    - `prTitle`: a `string`, the title of the PR to be created. This parameter is *mandatory*.
    - `prBody`: a `string`, the body of the PR to be created. This parameter is *mandatory*.
    - `labels`: a `[]string`, a list of labels to add to the PR once it has been created. No labels will be added if this parameter is omitted or empty.
    - `deletePath`: a `string`, the value of the `--delete-path` parameter of the `gh-commit` plugin. Setting this will delete all the files in the specified path before creating the commit.
    - `createEmpty`: a `bool`, whether or not to allow the creation of an empty commit (a commit with no files changed). By default, `false`, so trying to create a commit with no changes will throw an error.
    - `ctr`: a `*dagger.Container`, which is used to launch the `gh` commands. If this is omitted or unset, a new container will be created and used instead.

8. **DeleteRemoteBranch**: specific function that deletes a branch in the remote repository, closing any PRs opened with it. It returns nothing.  This function takes the [all generic params](#generic-params) and the following specific ones:
    - `repoDir`: a `*dagger.Directory`, the folder where the repository the branch that's going to be deleted from is stored. This parameter is *mandatory*.
    - `branchName`: a `string`, the name of the branch to delete remotely. This parameter is *mandatory*.
    - `ctr`: a `*dagger.Container`, which is used to launch the `gh` commands. If this is omitted or unset, a new container will be created and used instead.


## ⌨️  Generic params

- `ctx`: a [`context.Context`](https://pkg.go.dev/context), used by golang internally. This parameter is always *mandatory*, but when calling a function from the Dagger CLI it is automatically set.
- `version`: a `string`, the version of the `gh` cli to be downloaded. If left empty, the latest version will be downloaded.
- `token`: a `*dagger.Secret`, a GitHub token used for authentication with its services. Though in many functions this parameter is optional it should always be properly set to avoid authentication errors.
- `localGhCliPath`: a `*dagger.File`, a binary file of a `gh` cli program stored locally. When set, no download of the `gh` cli will be performed and this file will be used instead. If this file's version and the `version` parameter differ, a warning will be printed.

## 📃 Credits

- Based on [aweris/daggerverse/gh](https://daggerverse.dev/mod/github.com/aweris/daggerverse/gh)
