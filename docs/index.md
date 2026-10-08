# prest-guard

pREST middleware plugin: opt-in request security middleware for [pREST](https://github.com/prest/prest) (PostgreSQL REST), powered by the [guard-core-go](https://github.com/Guard-Core/guard-core-go) engine.

## What it does

When enabled, the plugin runs every incoming request through the guard-core pipeline before it reaches pREST: IP reputation and bans, rate limits, penetration-attempt detection, and block responses with JSON error bodies. Everything is opt-in and configured via `PREST_GUARD_*` environment variables.

## Installation

See the repository [README](https://github.com/Guard-Core/prest-guard#readme) for build and configuration.
