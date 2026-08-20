# forgepool

A tiny build-request dispatcher for a CI worker that talks to a pool of warm
build daemons over persistent connections.

Build daemons (think a Gradle daemon or a compile-server) keep artifacts and
analysis caches hot between requests, so `forgepool` keeps connections alive
and reuses them. When a request fails, `forgepool` decides how to retry.

## Package layout

- `internal/dispatch` — the connection `Pool` and the `Dispatch` retry driver.

## Status

Early WIP. The retry policy is being hardened for flaky daemons.
