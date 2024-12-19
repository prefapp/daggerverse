import { dag, Container, Directory, object, func } from "@dagger.io/dagger"

@object()
class Github {
  /**
   * Returns a container that echoes whatever string argument is provided
   */
  @func()
  containerEcho(stringArg: string): Container {
    return dag.container().from("alpine:latest").withExec(["echo", stringArg])
  }

  /**
   * Returns lines that match a pattern in the files of the provided Directory
   */
  @func()
  async grepDir(directoryArg: Directory, pattern: string): Promise<string> {
    return dag
      .container()
      .from("alpine:latest")
      .withMountedDirectory("/mnt", directoryArg)
      .withWorkdir("/mnt")
      .withExec(["grep", "-R", pattern, "."])
      .stdout()
  }

  /**
   * Creates a commit
   */
  @func()
  async commit(message: string): Promise<void> {
    await dag
      .container()
      .from("alpine/git:latest")
      .withEnv(["GIT_COMMITTER_NAME=John Doe", "");
  }
