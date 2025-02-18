FROM golang:1.21-alpine AS builder
RUN apk add --no-cache make
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux make build

FROM alpine:latest
RUN apk add --no-cache ca-certificates && \
    adduser -D nonroot
WORKDIR /app
COPY --from=builder /src/build/batcher .
USER nonroot
CMD ["/app/batcher"]