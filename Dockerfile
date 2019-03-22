FROM golang:1.11 AS build

RUN mkdir /app
WORKDIR /app

COPY go.mod ./go.mod
COPY go.sum ./go.sum

RUN go get -d -v ./...

COPY main.go .
COPY tellus ./tellus

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

#
# Running Container
#
FROM alpine:3.6

RUN apk add --no-cache git terraform

COPY --from=build /app/main /
ENTRYPOINT ["/main"]
