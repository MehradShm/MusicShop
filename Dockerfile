# ===== Build stage =====
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# produce static binary (no CGO)
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o /out/api ./cmd/api

# ===== Runtime stage =====
FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /out/api /api
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]