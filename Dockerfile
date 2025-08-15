# ===== Build stage =====
FROM golang:1.24-alpine AS builder
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Set working directory to /app (container's project root)
WORKDIR /app

# Copy go.mod and go.sum first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the project
COPY . .

# Build the backend app
RUN mkdir -p out && go build -o out/backend_app ./backend

# ===== Runtime stage =====
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/out/backend_app /app/backend_app
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/backend_app"]
