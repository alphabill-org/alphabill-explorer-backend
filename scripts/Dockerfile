FROM scratch AS go-dependency
# Empty

FROM golang:1.24-bookworm AS builder
ARG DOCKER_GO_DEPENDENCY=/go-dependency

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN --mount=from=go-dependency,target=$DOCKER_GO_DEPENDENCY \
    go mod download && go mod verify

COPY . .
# Cache Go build artifacts to speed up subsequent builds
RUN --mount=from=go-dependency,target=$DOCKER_GO_DEPENDENCY \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    make build

FROM gcr.io/distroless/base-debian12:debug-nonroot
USER nonroot
SHELL [ "/busybox/sh", "-c" ]
WORKDIR /app
ENV PATH="/app:${PATH}"
COPY --from=builder /usr/src/app/build/abexplorer /app/
COPY --from=builder /usr/src/app/cmd/config.yaml /app/

ENTRYPOINT ["abexplorer", "./config.yaml"]
