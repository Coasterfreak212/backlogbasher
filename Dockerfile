# Stage 1: Build single Go binary with embedded web assets & SQLite3
FROM golang:1.23-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.mod ./
COPY main.go ./
COPY internal/ ./internal/
COPY web/ ./web/

ENV CGO_ENABLED=1
RUN go build -trimpath -ldflags="-s -w" -o /out/backlogbasher .

# Stage 2: Minimal Debian Bookworm Slim runtime
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libsqlite3-0 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /out/backlogbasher /app/backlogbasher

RUN mkdir -p /app/data/uploads

ENV PORT=8080
ENV DB_PATH=/app/data/consoleerp.db
ENV UPLOAD_DIR=/app/data/uploads

EXPOSE 8080
VOLUME ["/app/data"]

ENTRYPOINT ["/app/backlogbasher"]
