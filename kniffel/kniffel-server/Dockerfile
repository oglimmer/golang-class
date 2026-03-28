# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /opt/build

# copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# copy source code
COPY . .

# run tests and build
RUN go test ./...
RUN CGO_ENABLED=0 go build -o kniffel .

# Runtime stage
FROM alpine:3.21

WORKDIR /opt/app

# copy the binary from the build stage
COPY --from=builder /opt/build/kniffel /opt/app/

CMD ["./kniffel"]
