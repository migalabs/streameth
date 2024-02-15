# syntax=docker/dockerfile:1
FROM golang:1.20-alpine as builder

RUN mkdir /app
WORKDIR /app
ADD . .

RUN go build -o ./build/streameth


FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=builder /app/build/streameth ./
ENTRYPOINT ["/streameth"]
