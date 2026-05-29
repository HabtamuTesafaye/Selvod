FROM golang:1.23-alpine AS builder
RUN apk add --no-cache gcc musl-dev ffmpeg

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o selvod ./cmd/selvod

FROM alpine:3.20
RUN apk add --no-cache ffmpeg ca-certificates shadow

# Create a non-privileged user
RUN groupadd -r selvod && useradd -r -g selvod selvod
RUN mkdir -p /app /data && chown selvod:selvod /app /data

WORKDIR /app
COPY --from=builder /app/selvod .
RUN chown selvod:selvod selvod

# Run as non-root
USER selvod

EXPOSE 8080
CMD ["./selvod"]
