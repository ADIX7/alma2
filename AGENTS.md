# Agent instructions

## Development and tests

- After completing development changes, always run relevant tests and the full
  available test suite (`go test ./...`). On Windows, use the Docker workflow
  below instead of running the suite natively or directly under WSL.
- Fix failures before declaring work complete. If tests cannot run, report why.
- For documentation-only changes, run applicable documentation checks. If no
  executable tests exist yet, state that explicitly; do not claim tests passed.

### Testing on Windows

Start Docker Desktop in Linux containers mode (`docker info` should show a Linux
server). From the repository root, build the image from the current checkout,
then run it only if the build succeeds:

```text
docker build -f Dockerfile.ci -t alma2-ci .
docker run --rm alma2-ci
```

Rebuild after changing code. The container runs the full Go suite, formatting
checks, `go vet`, a build, and the fixture rehearsal. Use the image as-is (no
bind mount or privileged mode) so symlink tests run on the container filesystem
without requiring Windows symlink privileges. For dependency download failures
on TLS-inspecting networks, see the [trusted CA instructions](README.md#run-tests-in-docker).
If Docker is unavailable, report that validation is blocked.

## Version control

- Use `jj` when the repository supports it; otherwise use Git.
