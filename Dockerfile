# Build stage
FROM golang:1.25-alpine AS build

RUN apk add --no-cache git

WORKDIR /app

ENV GOPROXY=https://goproxy.io,direct

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main ./cmd/server/main.go

# Run stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /app/main .
# Copy .env if needed (though usually passed by docker-compose)
COPY .env .

EXPOSE 8081

CMD ["./main"]