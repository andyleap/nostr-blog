FROM golang:alpine as builder
RUN mkdir /build 
WORKDIR /build 
ADD go.mod go.sum /build/
RUN go mod download
ADD . /build/
RUN go build -o nostr-blog .
FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/nostr-blog /app/

WORKDIR /app
LABEL org.opencontainers.image.source=https://github.com/andyleap/nostr-blog
CMD ["./nostr-blog"]
