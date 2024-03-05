FROM golang:alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o ./plex-autoscan

FROM scratch

COPY --from=builder /app/plex-autoscan ./plex-autoscan

ENTRYPOINT ["./plex-autoscan"]
