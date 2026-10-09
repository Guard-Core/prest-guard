# prest-guard demo stack

pREST built from a pinned upstream commit with the prest-guard plugin
compiled in, in front of a Postgres instance. The guard runs enforcing with
rate limiting (100 req/min per client IP) and payload inspection on; flip
`PREST_GUARD_PASSIVE=true` in the compose environment to preview instead of
reject.

```bash
docker compose -f examples/docker-stack/docker-compose.yml up --build -d --wait
# any request passes through the guard: rate limits, IP policy, and payload
# inspection run on every route, e.g.
curl -i "http://localhost:8080/databases"
curl -i "http://localhost:8080/databases?_name=%27%20OR%20%271%27%3D%271"
```

The plugin constraint that shapes the Dockerfile: native Go plugins only
load into a host binary sharing their build identity, so the image builds
prestd and `guard.so` in one stage with the same toolchain and prest pinned
(`PREST_REF`) to the upstream commit whose shared dependency pins match the
plugin's `go.mod`. Bump `PREST_REF` when prest main moves.
