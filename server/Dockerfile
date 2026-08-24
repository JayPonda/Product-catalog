# syntax=docker/dockerfile:1

# ---------- Build stage ----------
FROM golang:1.26-alpine AS build

WORKDIR /src

# Copy manifests first so Docker can cache the dependency layer.
# The repo vendors dependencies (vendor/), which also makes this build hermetic.
COPY go.mod go.sum ./
COPY vendor/ vendor/
RUN go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/product-catalog .

# ---------- Runtime stage ----------
FROM alpine:3.22

# Non-root user; ca-certificates for TLS connections to managed Postgres.
# wget (busybox) stays available for container healthchecks.
RUN addgroup -S app && adduser -S app -G app \
    && apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /out/product-catalog /usr/local/bin/product-catalog
# sql-migrate reads ./migrations relative to the working directory.
COPY migrations/ /app/migrations/

USER app
EXPOSE 3300

ENTRYPOINT ["/usr/local/bin/product-catalog"]
CMD ["server"]
