# syntax=docker/dockerfile:1

FROM golang:1.18-buster AS build

WORKDIR /app

COPY go.mod .
RUN go mod download

COPY . .

RUN go mod tidy

RUN go build -o /t main/main.go main/inst_list.go


## Deploy

FROM gcr.io/distroless/base-debian10

ENV GO111MODULE=on
ENV GOOGLE_APPLICATION_CREDENTIALS='/data/credentials/creds.json'

WORKDIR /

COPY --from=build /t /t

EXPOSE 8080

USER root:root

ENTRYPOINT ["/t"]
