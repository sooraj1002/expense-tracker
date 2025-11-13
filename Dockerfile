# syntax=docker/dockerfile:1

############################
# Builder image
############################
FROM golang:1.22-bullseye AS builder

WORKDIR /app
ENV CGO_ENABLED=0 GOOS=linux GOTOOLCHAIN=auto

# Cache module downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build the binary
COPY . .
RUN go build -o expense-tracker .

############################
# Runtime image
############################
FROM gcr.io/distroless/base-debian12

WORKDIR /app
ENV PORT=8080

COPY --from=builder /app/expense-tracker /app/expense-tracker

EXPOSE 8080
ENTRYPOINT ["/app/expense-tracker", "serve"]
