# AGENTS.md

Guidance for AI agents working in this repository.

## What this repo is

prest-guard is a pREST middleware plugin over the guard-core-go engine. It
is adapter code only: detection logic, pattern tables, and check behavior
belong in guard-core-go (and its reference, Python guard-core). If a fix
changes detection behavior, it goes to the engine, not here.

## Hard constraints

- The whole repo is package main: pREST builds plugins from package main.
  Do not split into importable subpackages; keep functions testable instead.
- `GuardMiddlewareLoad` is the symbol the pREST loader looks up
  ([[pluginmiddlewarelist]] file = "guard", func = "Guard"). Renaming it
  breaks every deployment.
- `Load` must never return nil and must never block: config errors and
  failed engine init return immediate 500/503 handlers, and init retries
  happen on a background goroutine.
- Plugins receive no host configuration. All settings come from
  `PREST_GUARD_*` env vars; never read pREST's config file from here.
- Never let the guard pick up pREST's own `REDIS_URL`: the engine defaults
  read that env var, so `buildEngineConfig` pins the Redis fields.
  The same pattern applies to any new engine default that reads the
  environment.
- Error bodies to clients are generic and marshaled via writeJSONError;
  config details go to logs only.

## Gates before push

```bash
gofmt -l .        # must be empty
go vet ./...
go test -race ./...
go build -buildmode=plugin -trimpath -o /tmp/guard.so .   # the artifact builds
```

Versioning: this repo moves on its own change (SemVer). Engine bumps here
follow the guard-core-go releases the plugin is tested against; do not bump
engines speculatively.

## Benchmarks are part of the contract

`make bench` numbers back the README's body-scanning cost table. If a
change moves them by more than ~20%, update the table and say so in the
changelog entry.
