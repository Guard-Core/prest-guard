# Contributing

Small repo, few rules:

- Run the gates before pushing: `gofmt -l .` empty, `go vet ./...` clean,
  `go test -race ./...` green, and the plugin artifact building
  (`make so`).
- One topic per PR. If a fix belongs in the engine (detection behavior,
  check semantics), file it against
  [guard-core-go](https://github.com/Guard-Core/guard-core-go) instead; this
  repo maps and wires, it does not detect.
- New configuration surface needs: raw-string validation in
  `ParseGuardConfig`, a mapping in `buildEngineConfig`, a default in the
  README table, and a test.
- Commit messages follow the conventional commits style used across the
  guard-core ecosystem (`feat:`, `fix:`, `docs:`, `test:`, `chore:`).
