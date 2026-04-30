FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /usr/local/bin/tidal ./cmd/tidal

FROM alpine:3.21

RUN apk add --no-cache ffmpeg ca-certificates tzdata

COPY --from=builder /usr/local/bin/tidal /usr/local/bin/tidal

EXPOSE 8080

ENTRYPOINT ["tidal"]
CMD ["server"]
