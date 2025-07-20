# --- Build Stage ---
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /app/job-server ./cmd/job-server
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /app/appwright-agent ./cmd/appwright-agent

# --- Final Stage ---
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/job-server .
COPY --from=builder /app/appwright-agent .
# COPY config.yml.example .
CMD ["./job-server"]
