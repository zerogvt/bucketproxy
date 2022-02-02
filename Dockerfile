# syntax=docker/dockerfile:1

FROM golang:1.16-alpine

WORKDIR /app
RUN mkdir /app/server
COPY server/* /app/server/

COPY go.mod ./
COPY go.sum ./
RUN go mod download

RUN go build -o bucketproxy ./server/

EXPOSE 8080

CMD [ "./bucketproxy" ]
