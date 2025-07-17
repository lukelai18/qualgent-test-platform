# --- Build Stage ---
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /app/job-server ./cmd/job-server

# --- Final Stage ---
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/job-server .
# COPY config.yml.example .
CMD ["./job-server"]
