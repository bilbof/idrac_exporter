# syntax=docker/dockerfile:1

## Build
FROM golang:1.18-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /idrac_exporter

## Deploy
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /idrac_exporter /idrac_exporter

EXPOSE 9348

USER nonroot:nonroot

ENTRYPOINT ["/idrac_exporter"]
