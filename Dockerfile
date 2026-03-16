FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s -extldflags=-static" -trimpath \
    -o /bin/api ./cmd/api

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /bin/api /api
COPY internal/db/migrations/ /migrations/
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]
