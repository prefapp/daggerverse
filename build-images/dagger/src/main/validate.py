import dagger
from dagger import dag, function


@function
def foo(self) -> dagger.Container:
    """Returns a container that echoes whatever string argument is provided"""
    return dag.container().from_("alpine:latest").with_exec(["echo", "foo"])
