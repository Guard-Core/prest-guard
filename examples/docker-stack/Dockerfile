# syntax=docker/dockerfile:1

# Build context is the repository root:
#   docker build -f examples/docker-stack/Dockerfile .
#
# Native Go plugins only load into a host binary that shares their build
# identity, so prestd and guard.so come out of one builder stage: same
# toolchain, CGO on, and prest pinned to the upstream commit whose shared
# dependency pins match the plugin's go.mod (negroni v3.1.1 today). Bump
# PREST_REF when prest main moves and the pins are re-checked.
ARG PREST_REF=f6f5cd4fa777c6e2f138e9adb125218aba14c06c

FROM golang:1.26-bookworm AS build
ARG PREST_REF
ENV CGO_ENABLED=1
WORKDIR /build

RUN apt-get update && apt-get install -y --no-install-recommends git ca-certificates && \
    rm -rf /var/lib/apt/lists/*
RUN git clone https://github.com/prest/prest.git /build/prest && \
    git -C /build/prest checkout "${PREST_REF}"

COPY go.mod go.sum ./
COPY *.go ./
RUN go build -buildmode=plugin -trimpath -ldflags "-s -w" -o /out/lib/middlewares/guard.so .

RUN cd /build/prest && go build -trimpath -ldflags "-s -w" -o /out/prestd ./cmd/prestd/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl && \
    rm -rf /var/lib/apt/lists/* && \
    useradd --create-home --uid 1000 guard

COPY --from=build /out/prestd /usr/local/bin/prestd
COPY --from=build /out/lib /app/lib
COPY examples/docker-stack/prest.toml /app/prest.toml

LABEL org.opencontainers.image.title="prest-guard demo" \
    org.opencontainers.image.description="Containerized pREST stack with the prest-guard plugin: the guard-core engine in front of pREST's CRUD routes" \
    org.opencontainers.image.source="https://github.com/Guard-Core/prest-guard" \
    org.opencontainers.image.licenses="MIT"

USER guard
WORKDIR /app

ENV PREST_CONF=/app/prest.toml

EXPOSE 8080

ENTRYPOINT ["prestd"]
