# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Download dependencies first (cache layer)
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source and build
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

# The server uses relative paths:
#   ../frontend/*.html  (templates)
#   ../frontend/styles/ (static assets)
#   uploads/            (user uploads)
# So the binary must run from /app/backend/
WORKDIR /app/backend

COPY --from=builder /build/server ./server
COPY backend/migrations/ ./migrations/
COPY frontend/ /app/frontend/

# Persistent uploads directory (mount a Railway volume here)
RUN mkdir -p uploads

EXPOSE 8085

CMD ["./server"]
