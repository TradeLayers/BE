  # Base image
  FROM golang:1.25-alpine AS builder

  # Working directory
  WORKDIR /app

  # Go modules
  COPY go.mod go.sum ./
  RUN go mod download

  # Copy source
  COPY . .
  RUN CGO_ENABLED=0 go build -o server ./cmd/api

  # Run stage
  FROM alpine:3.21
  WORKDIR /app
  COPY --from=builder /app/server .
  COPY --from=builder /app/firebase.json .

  EXPOSE 5000
  CMD ["./server"]