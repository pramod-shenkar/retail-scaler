# Builder stage
FROM golang:1.25 as builder

WORKDIR /app

# Step 1: Copy go.mod and go.sum first to leverage Docker cache
# COPY go.mod go.sum ./

# Step 2: Run go mod tidy to ensure dependencies are fetched
# RUN go mod tidy

# Step 3: Copy the rest of the application code
COPY . .

# Build the producer and consumer
# RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/producer /app/producer/main.go
# RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/consumer /app/consumer/main.go

# # Runtime stage
# FROM gcr.io/distroless/static-debian12:nonroot

# COPY --from=builder /bin/producer /bin/producer
# COPY --from=builder /bin/consumer /bin/consumer

EXPOSE 8080
