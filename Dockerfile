FROM golang:1.17-alpine3.14 AS build

RUN apk --no-cache add make git gcc libc-dev curl
WORKDIR /app

COPY go.mod go.sum Makefile ./
COPY . .
RUN make build


FROM alpine:3.14.3

COPY --from=build /app/bin/mirage /bin/mirage

ENTRYPOINT [ "/bin/mirage" ]
