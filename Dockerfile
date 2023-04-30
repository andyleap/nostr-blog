FROM golang:alpine as builder
RUN mkdir /build 
WORKDIR /build 
ADD go.mod go.sum /build/
RUN go mod download
ADD . /build/
RUN go build -o main .
FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/main /app/

WORKDIR /app

ADD static/ static/
CMD ["./main"]
