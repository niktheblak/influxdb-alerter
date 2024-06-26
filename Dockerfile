FROM golang:1.22 as build

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /go/bin/app

FROM ubuntu:latest
COPY --from=build /go/bin/app /
ENTRYPOINT ["/app", "check"]
